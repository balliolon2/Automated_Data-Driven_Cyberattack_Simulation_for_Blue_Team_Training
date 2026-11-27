package controllers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode"

	"cybersim/dto"
	"cybersim/middlewares"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type fakeExamStore struct {
	sessions        map[string]*models.ExamSession
	sessionQs       map[string][]models.ExamSessionQuestion
	questions       map[string]models.Question
	profiles        map[string][]models.UserSkillProfile
	learnSeen       map[string]map[string]bool
	nextSessCounter int
}

func newFakeExamStore() *fakeExamStore {
	return &fakeExamStore{
		sessions:  make(map[string]*models.ExamSession),
		sessionQs: make(map[string][]models.ExamSessionQuestion),
		questions: make(map[string]models.Question),
		profiles:  make(map[string][]models.UserSkillProfile),
		learnSeen: make(map[string]map[string]bool),
	}
}

func (f *fakeExamStore) seedExamPool(prefix string, countPerDomain map[string]int) {
	subtopicByDomain := make(map[string]string, len(services.OrderedTheoryDomains))
	for _, spec := range services.OrderedTheorySubtopics {
		if _, exists := subtopicByDomain[spec.DomainID]; !exists {
			subtopicByDomain[spec.DomainID] = spec.SubTopicID
		}
	}
	for dom, n := range countPerDomain {
		st := subtopicByDomain[dom]
		for i := 0; i < n; i++ {
			qid := prefix + "-" + dom + "-" + string(rune('a'+i))
			var subPtr *string
			if st != "" {
				s := st
				subPtr = &s
			}
			f.questions[qid] = models.Question{
				QuestionID:    qid,
				DomainID:      dom,
				SubTopicID:    subPtr,
				Type:          "multiple_choice",
				QuestionText:  "Which control mitigates " + dom + " risk item " + qid + "?",
				Options:       map[string]string{"A": "Control A", "B": "Control B", "C": "Control C", "D": "Control D"},
				CorrectAnswer: "A",
				Explanation:   "Correct Answer: A. Control A provides primary defense for " + dom + ". Option B is incorrect. Control B does not address the threat. Option C is incorrect. Control C is unrelated. Option D is incorrect. Control D lacks coverage.",
				Difficulty:    3,
				IsActive:      true,
			}
		}
	}
}

func (f *fakeExamStore) GetActiveSessionForUser(userID string) (*models.ExamSession, error) {
	for _, s := range f.sessions {
		if s.UserID == userID && s.Status == "in_progress" {
			c := *s
			return &c, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeExamStore) HasCompletedPreTest(userID string) (bool, error) {
	for _, s := range f.sessions {
		if s.UserID == userID && s.ExamType == "pre" && s.Status == "completed" {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeExamStore) GetUserSkillProfiles(userID string) ([]models.UserSkillProfile, error) {
	return f.profiles[userID], nil
}

func (f *fakeExamStore) SelectQuestionsByDomains(userID string, domainCounts map[string]int, excludePreAndLearn bool) ([]models.Question, error) {
	excluded := make(map[string]bool)
	if excludePreAndLearn {
		for _, s := range f.sessions {
			if s.UserID == userID && s.ExamType == "pre" {
				for _, esq := range f.sessionQs[s.SessionID] {
					excluded[esq.QuestionID] = true
				}
			}
		}
		for qid := range f.learnSeen[userID] {
			excluded[qid] = true
		}
	}

	var selected []models.Question
	orderedDomains := []string{"domain1", "domain2", "domain3", "domain4", "domain5"}
	for _, dom := range orderedDomains {
		needed := domainCounts[dom]
		taken := 0
		for _, q := range f.questions {
			if q.DomainID == dom && q.IsActive && !excluded[q.QuestionID] {
				selected = append(selected, q)
				taken++
				if taken >= needed {
					break
				}
			}
		}
	}
	return selected, nil
}

func (f *fakeExamStore) CreateExamSession(session *models.ExamSession, questions []models.Question) error {
	if session.SessionID == "" {
		f.nextSessCounter++
		session.SessionID = "sess-" + string(rune('0'+f.nextSessCounter))
	}
	c := *session
	f.sessions[session.SessionID] = &c
	var esqs []models.ExamSessionQuestion
	for idx, q := range questions {
		esqs = append(esqs, models.ExamSessionQuestion{
			SessionID:  session.SessionID,
			QuestionID: q.QuestionID,
			OrderIndex: idx,
		})
	}
	f.sessionQs[session.SessionID] = esqs
	return nil
}

func (f *fakeExamStore) GetSessionQuestions(sessionID string) ([]QuestionResponse, error) {
	esqs := f.sessionQs[sessionID]
	resp := make([]QuestionResponse, len(esqs))
	for i, esq := range esqs {
		q := f.questions[esq.QuestionID]
		resp[i] = QuestionResponse{
			QuestionID:   q.QuestionID,
			DomainID:     q.DomainID,
			Type:         q.Type,
			QuestionText: q.QuestionText,
			Options:      q.Options,
			Translations: services.SanitizeQuestionTranslations(q.Translations, false),
			OrderIndex:   esq.OrderIndex,
			UserAnswer:   esq.UserAnswer,
		}
	}
	return resp, nil
}

func (f *fakeExamStore) FindActiveSessionByIDAndUser(sessionID, userID string) (*models.ExamSession, error) {
	s, ok := f.sessions[sessionID]
	if !ok || s.UserID != userID || s.Status != "in_progress" {
		return nil, gorm.ErrRecordNotFound
	}
	c := *s
	return &c, nil
}

func (f *fakeExamStore) FindSessionByIDAndUser(sessionID, userID string) (*models.ExamSession, error) {
	s, ok := f.sessions[sessionID]
	if !ok || s.UserID != userID {
		return nil, gorm.ErrRecordNotFound
	}
	c := *s
	return &c, nil
}

func (f *fakeExamStore) FindQuestionByID(questionID string) (*models.Question, error) {
	q, ok := f.questions[questionID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	c := q
	return &c, nil
}

func (f *fakeExamStore) FindSessionQuestion(sessionID, questionID string) (*models.ExamSessionQuestion, error) {
	for _, esq := range f.sessionQs[sessionID] {
		if esq.QuestionID == questionID {
			c := esq
			return &c, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeExamStore) SaveExamAnswer(session *models.ExamSession, esq *models.ExamSessionQuestion) (int64, error) {
	esqs := f.sessionQs[session.SessionID]
	var unanswered int64
	for i := range esqs {
		if esqs[i].QuestionID == esq.QuestionID {
			esqs[i] = *esq
		}
		if strings.TrimSpace(esqs[i].UserAnswer) == "" {
			unanswered++
		}
	}
	f.sessionQs[session.SessionID] = esqs
	if stored, ok := f.sessions[session.SessionID]; ok {
		stored.TimeSpentSeconds = session.TimeSpentSeconds
	}
	return unanswered, nil
}

func (f *fakeExamStore) buildOrderedResults(sessionID string) []dto.DetailedExamResultDTO {
	esqs := make([]models.ExamSessionQuestion, len(f.sessionQs[sessionID]))
	copy(esqs, f.sessionQs[sessionID])
	sort.Slice(esqs, func(i, j int) bool {
		return esqs[i].OrderIndex < esqs[j].OrderIndex
	})

	detailed := make([]dto.DetailedExamResultDTO, 0, len(esqs))
	for _, esq := range esqs {
		q := f.questions[esq.QuestionID]
		isCorr := esq.IsCorrect != nil && *esq.IsCorrect
		detailed = append(detailed, dto.DetailedExamResultDTO{
			QuestionID:    q.QuestionID,
			DomainID:      q.DomainID,
			SubTopicID:    q.SubTopicID,
			Type:          q.Type,
			QuestionText:  q.QuestionText,
			Options:       q.Options,
			CorrectAnswer: q.CorrectAnswer,
			Explanation:   q.Explanation,
			Translations:  services.SanitizeQuestionTranslations(q.Translations, true),
			MitreRef:      q.MitreRef,
			OrderIndex:    esq.OrderIndex,
			UserAnswer:    esq.UserAnswer,
			IsCorrect:     isCorr,
		})
	}
	return detailed
}

func (f *fakeExamStore) MarkUnansweredAndFinalize(session *models.ExamSession, userID string) (*services.ExamReviewData, error) {
	if stored, ok := f.sessions[session.SessionID]; ok && stored.Status == "completed" {
		*session = *stored
		return services.SummarizeExamResults(f.buildOrderedResults(stored.SessionID), examDisplayThreshold()), nil
	}
	esqs := f.sessionQs[session.SessionID]
	for i := range esqs {
		if strings.TrimSpace(esqs[i].UserAnswer) == "" {
			esqs[i].UserAnswer = ""
			falseVal := false
			esqs[i].IsCorrect = &falseVal
		}
	}
	f.sessionQs[session.SessionID] = esqs

	detailed := f.buildOrderedResults(session.SessionID)
	reviewData := services.SummarizeExamResults(detailed, examDisplayThreshold())
	denom := session.TotalQuestions
	if denom <= 0 {
		denom = len(detailed)
	}
	scorePct := services.ComputeSessionScore(nil, reviewData.CorrectCount, denom)
	now := time.Now()
	if stored, ok := f.sessions[session.SessionID]; ok {
		stored.Score = &scorePct
		stored.Status = "completed"
		stored.CompletedAt = &now
		stored.TimeSpentSeconds = session.TimeSpentSeconds
		*session = *stored
	} else {
		session.Score = &scorePct
		session.Status = "completed"
		session.CompletedAt = &now
	}

	if session.ExamType == "pre" {
		f.profiles[userID] = nil
		for _, ds := range reviewData.DomainScores {
			f.profiles[userID] = append(f.profiles[userID], models.UserSkillProfile{
				UserID:           userID,
				DomainID:         ds.DomainID,
				ProficiencyScore: ds.Percentage,
			})
		}
	}

	return reviewData, nil
}

func (f *fakeExamStore) FinalizeExamSession(session *models.ExamSession, userID string) (*services.ExamReviewData, error) {
	return f.MarkUnansweredAndFinalize(session, userID)
}

func (f *fakeExamStore) GetExamSessionReviewData(session *models.ExamSession) (*services.ExamReviewData, error) {
	return services.SummarizeExamResults(f.buildOrderedResults(session.SessionID), examDisplayThreshold()), nil
}

func (f *fakeExamStore) GetCompletedExamHistory(userID string) ([]dto.ExamHistoryItemDTO, error) {
	var sessions []*models.ExamSession
	for _, s := range f.sessions {
		if s.UserID == userID && s.Status == "completed" {
			c := *s
			sessions = append(sessions, &c)
		}
	}
	sort.Slice(sessions, func(i, j int) bool {
		var ti, tj time.Time
		if sessions[i].CompletedAt != nil {
			ti = *sessions[i].CompletedAt
		} else {
			ti = sessions[i].StartedAt
		}
		if sessions[j].CompletedAt != nil {
			tj = *sessions[j].CompletedAt
		} else {
			tj = sessions[j].StartedAt
		}
		return ti.After(tj)
	})

	items := make([]dto.ExamHistoryItemDTO, 0, len(sessions))
	for _, sess := range sessions {
		reviewData := services.SummarizeExamResults(f.buildOrderedResults(sess.SessionID), examDisplayThreshold())
		total := sess.TotalQuestions
		if len(reviewData.Results) > 0 {
			total = len(reviewData.Results)
		}
		items = append(items, dto.ExamHistoryItemDTO{
			SessionID:        sess.SessionID,
			ExamType:         sess.ExamType,
			Status:           sess.Status,
			Score:            services.ComputeSessionScore(sess.Score, reviewData.CorrectCount, total),
			CorrectCount:     reviewData.CorrectCount,
			IncorrectCount:   reviewData.IncorrectCount,
			UnansweredCount:  reviewData.UnansweredCount,
			TotalQuestions:   total,
			StartedAt:        sess.StartedAt,
			CompletedAt:      sess.CompletedAt,
			TimeSpentSeconds: sess.TimeSpentSeconds,
		})
	}
	return items, nil
}

func (f *fakeExamStore) GetLatestCompletedPreTestSession(userID string) (*models.ExamSession, error) {
	var latest *models.ExamSession
	for _, s := range f.sessions {
		if s.UserID == userID && s.ExamType == "pre" && s.Status == "completed" {
			if latest == nil {
				c := *s
				latest = &c
				continue
			}
			var ts, tl time.Time
			if s.CompletedAt != nil {
				ts = *s.CompletedAt
			} else {
				ts = s.StartedAt
			}
			if latest.CompletedAt != nil {
				tl = *latest.CompletedAt
			} else {
				tl = latest.StartedAt
			}
			if ts.After(tl) {
				c := *s
				latest = &c
			}
		}
	}
	if latest == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return latest, nil
}

func (f *fakeExamStore) GetRandomSampleQuestion() (*models.Question, error) {
	for _, q := range f.questions {
		c := q
		return &c, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func makeExamRouter(ec *ExamController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	api.GET("/exams/sample", ec.GetSampleQuestion)
	exams := api.Group("/exams", middlewares.AuthMiddleware())
	exams.POST("/pre-test", ec.StartPreTest)
	exams.POST("/post-test", ec.StartPostTest)
	exams.GET("/session", ec.GetActiveSession)
	exams.POST("/submit-answer", ec.SubmitAnswer)
	exams.POST("/sessions/:sessionId/finalize", ec.FinalizeSession)
	exams.GET("/history", ec.GetExamHistory)
	exams.GET("/sessions/:sessionId/review", ec.GetExamReview)
	return r
}

func signExamJWT(t *testing.T, userID string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    "learner",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	signed, err := tok.SignedString([]byte("exam-test-secret"))
	if err != nil {
		t.Fatalf("sign jwt failed: %v", err)
	}
	return signed
}

func containsThaiRune(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Thai) || (r >= 0x0E00 && r <= 0x0E7F) {
			return true
		}
	}
	return false
}

func TestExamController_PostTestGatesAndErrorCodes(t *testing.T) {
	t.Setenv("JWT_SECRET", "exam-test-secret")
	store := newFakeExamStore()
	store.seedExamPool("q1", map[string]int{"domain1": 8, "domain2": 14, "domain3": 10, "domain4": 16, "domain5": 12})
	ec := NewExamControllerWithStore(store)
	r := makeExamRouter(ec)

	tok := signExamJWT(t, "u-exam-1")
	headers := map[string]string{
		"Authorization":   "Bearer " + tok,
		"Accept-Language": "th",
	}

	// 1. GetActiveSession when none exists -> 404 EXAM_SESSION_NOT_FOUND with "message"
	wNoSess := doJSONReq(t, r, http.MethodGet, "/api/exams/session", nil, headers)
	var noSessBody map[string]any
	_ = json.Unmarshal(wNoSess.Body.Bytes(), &noSessBody)
	if wNoSess.Code != http.StatusNotFound || noSessBody["message"] != "No active session found" || noSessBody["code"] != "EXAM_SESSION_NOT_FOUND" {
		t.Fatalf("expected 404 EXAM_SESSION_NOT_FOUND, got %d %+v", wNoSess.Code, noSessBody)
	}

	// 2. StartPostTest before Pre-Test -> 403 EXAM_PRETEST_REQUIRED
	wPreReq := doJSONReq(t, r, http.MethodPost, "/api/exams/post-test", nil, headers)
	var preReqBody map[string]any
	_ = json.Unmarshal(wPreReq.Body.Bytes(), &preReqBody)
	if wPreReq.Code != http.StatusForbidden || preReqBody["eligible"] != false || preReqBody["reason"] != "pre_test_required" || preReqBody["code"] != "EXAM_PRETEST_REQUIRED" {
		t.Fatalf("expected 403 EXAM_PRETEST_REQUIRED, got %d %+v", wPreReq.Code, preReqBody)
	}

	// 3. Complete a pre-test with domain2 below 70% -> 403 EXAM_POSTTEST_LOCKED with params
	store.sessions["pre-done"] = &models.ExamSession{
		SessionID:      "pre-done",
		UserID:         "u-exam-1",
		ExamType:       "pre",
		Status:         "completed",
		TotalQuestions: 30,
	}
	store.profiles["u-exam-1"] = []models.UserSkillProfile{
		{UserID: "u-exam-1", DomainID: "domain1", ProficiencyScore: 80},
		{UserID: "u-exam-1", DomainID: "domain2", ProficiencyScore: 55},
		{UserID: "u-exam-1", DomainID: "domain3", ProficiencyScore: 75},
		{UserID: "u-exam-1", DomainID: "domain4", ProficiencyScore: 80},
		{UserID: "u-exam-1", DomainID: "domain5", ProficiencyScore: 90},
	}

	wLocked := doJSONReq(t, r, http.MethodPost, "/api/exams/post-test", nil, headers)
	var lockedBody map[string]any
	_ = json.Unmarshal(wLocked.Body.Bytes(), &lockedBody)
	if wLocked.Code != http.StatusForbidden || lockedBody["reason"] != "threshold_not_reached" || lockedBody["code"] != "EXAM_POSTTEST_LOCKED" || lockedBody["domain_id"] != "domain2" {
		t.Fatalf("expected 403 EXAM_POSTTEST_LOCKED for domain2, got %d %+v", wLocked.Code, lockedBody)
	}
	params, ok := lockedBody["params"].(map[string]any)
	if !ok || params["domain_id"] != "domain2" || params["threshold"] != float64(70) {
		t.Fatalf("expected params with domain_id and threshold, got %+v", lockedBody["params"])
	}

	// 4. StartPreTest + active session guard -> 400 EXAM_SESSION_ACTIVE
	wStart := doJSONReq(t, r, http.MethodPost, "/api/exams/pre-test", nil, headers)
	if wStart.Code != http.StatusCreated {
		t.Fatalf("expected 201 StartPreTest, got %d: %s", wStart.Code, wStart.Body.String())
	}
	wDup := doJSONReq(t, r, http.MethodPost, "/api/exams/pre-test", nil, headers)
	var dupBody map[string]any
	_ = json.Unmarshal(wDup.Body.Bytes(), &dupBody)
	if wDup.Code != http.StatusBadRequest || dupBody["code"] != "EXAM_SESSION_ACTIVE" || dupBody["session_id"] == "" {
		t.Fatalf("expected 400 EXAM_SESSION_ACTIVE with session_id, got %d %+v", wDup.Code, dupBody)
	}

	// 5. SubmitAnswer error codes
	wInvalidAns := doJSONReq(t, r, http.MethodPost, "/api/exams/submit-answer", map[string]any{
		"session_id":  dupBody["session_id"],
		"question_id": "any-q",
		"user_answer": "",
	}, headers)
	var invAnsBody map[string]any
	_ = json.Unmarshal(wInvalidAns.Body.Bytes(), &invAnsBody)
	if wInvalidAns.Code != http.StatusBadRequest || invAnsBody["code"] != "EXAM_INVALID_ANSWER" {
		t.Fatalf("expected 400 EXAM_INVALID_ANSWER, got %d %+v", wInvalidAns.Code, invAnsBody)
	}

	wBadSess := doJSONReq(t, r, http.MethodPost, "/api/exams/submit-answer", map[string]any{
		"session_id":  "non-existent",
		"question_id": "any-q",
		"user_answer": "A",
	}, headers)
	var badSessBody map[string]any
	_ = json.Unmarshal(wBadSess.Body.Bytes(), &badSessBody)
	if wBadSess.Code != http.StatusForbidden || badSessBody["code"] != "EXAM_SESSION_NOT_FOUND" {
		t.Fatalf("expected 403 EXAM_SESSION_NOT_FOUND, got %d %+v", wBadSess.Code, badSessBody)
	}

	wMissingQ := doJSONReq(t, r, http.MethodPost, "/api/exams/submit-answer", map[string]any{
		"session_id":  dupBody["session_id"],
		"question_id": "does-not-exist",
		"user_answer": "A",
	}, headers)
	var missingQBody map[string]any
	_ = json.Unmarshal(wMissingQ.Body.Bytes(), &missingQBody)
	if wMissingQ.Code != http.StatusNotFound || missingQBody["code"] != "EXAM_QUESTION_NOT_FOUND" {
		t.Fatalf("expected 404 EXAM_QUESTION_NOT_FOUND, got %d %+v", wMissingQ.Code, missingQBody)
	}
}

func TestExamController_StrictEnglishOnlyInvariantUnderThaiLocale(t *testing.T) {
	t.Setenv("JWT_SECRET", "exam-test-secret")
	store := newFakeExamStore()
	store.seedExamPool("en", map[string]int{"domain1": 8, "domain2": 14, "domain3": 10, "domain4": 16, "domain5": 12})
	ec := NewExamControllerWithStore(store)
	r := makeExamRouter(ec)

	// User has preferred_language="th" (represented by Accept-Language: th)
	tok := signExamJWT(t, "u-thai-learner")
	headers := map[string]string{
		"Authorization":   "Bearer " + tok,
		"Accept-Language": "th",
	}

	// Sample endpoint
	wSample := doJSONReq(t, r, http.MethodGet, "/api/exams/sample", nil, headers)
	if wSample.Code != http.StatusOK || containsThaiRune(wSample.Body.String()) {
		t.Fatalf("sample question must be 200 and strictly English, got %d: %s", wSample.Code, wSample.Body.String())
	}

	// Start Pre-Test
	wPre := doJSONReq(t, r, http.MethodPost, "/api/exams/pre-test", nil, headers)
	if wPre.Code != http.StatusCreated || containsThaiRune(wPre.Body.String()) {
		t.Fatalf("pre-test start must be 201 and strictly English, got %d", wPre.Code)
	}

	var preResp struct {
		Session   models.ExamSession `json:"session"`
		Questions []QuestionResponse `json:"questions"`
	}
	_ = json.Unmarshal(wPre.Body.Bytes(), &preResp)
	if len(preResp.Questions) != 30 {
		t.Fatalf("expected 30 pre-test questions, got %d", len(preResp.Questions))
	}

	// Active Session recovery
	wSess := doJSONReq(t, r, http.MethodGet, "/api/exams/session", nil, headers)
	if wSess.Code != http.StatusOK || containsThaiRune(wSess.Body.String()) {
		t.Fatalf("active session must be 200 and strictly English, got %d", wSess.Code)
	}

	// Answer all 30 questions to inspect final explanations & correct_answers
	var lastSubmitBody string
	for _, q := range preResp.Questions {
		wSub := doJSONReq(t, r, http.MethodPost, "/api/exams/submit-answer", map[string]any{
			"session_id":         preResp.Session.SessionID,
			"question_id":        q.QuestionID,
			"user_answer":        "A",
			"time_spent_seconds": 10,
		}, headers)
		if wSub.Code != http.StatusOK {
			t.Fatalf("submit failed: %d %s", wSub.Code, wSub.Body.String())
		}
		lastSubmitBody = wSub.Body.String()
		if containsThaiRune(lastSubmitBody) {
			t.Fatalf("submit response must contain zero Thai runes: %s", lastSubmitBody)
		}
	}

	var finalResp struct {
		Completed bool `json:"completed"`
		Results   []struct {
			QuestionText  string `json:"question_text"`
			Explanation   string `json:"explanation"`
			CorrectAnswer string `json:"correct_answer"`
		} `json:"results"`
	}
	_ = json.Unmarshal([]byte(lastSubmitBody), &finalResp)
	if !finalResp.Completed || len(finalResp.Results) != 30 {
		t.Fatalf("expected completed=true with 30 results, got %+v", finalResp)
	}

	// Start Post-Test (all domain scores are 100% >= 70%)
	wPost := doJSONReq(t, r, http.MethodPost, "/api/exams/post-test", nil, headers)
	if wPost.Code != http.StatusCreated || containsThaiRune(wPost.Body.String()) {
		t.Fatalf("post-test start must be 201 and strictly English, got %d: %s", wPost.Code, wPost.Body.String())
	}
}

func TestExamController_PreSubmissionSanitizationAndPostSubmissionTranslations(t *testing.T) {
	t.Setenv("JWT_SECRET", "exam-test-secret")
	store := newFakeExamStore()
	store.seedExamPool("bi", map[string]int{"domain1": 8, "domain2": 14, "domain3": 10, "domain4": 16, "domain5": 12})
	for qid, q := range store.questions {
		q.Translations = map[string]any{
			"th": map[string]any{
				"question_text": "คำถามภาษาไทยสำหรับ " + qid,
				"options": map[string]any{
					"A": "ตัวเลือก ก",
					"B": "ตัวเลือก ข",
					"C": "ตัวเลือก ค",
					"D": "ตัวเลือก ง",
				},
				"explanation": "คำตอบที่ถูกต้อง: A. เหตุผลภาษาไทย ตัวเลือก B ไม่ถูกต้อง ผิด ตัวเลือก C ไม่ถูกต้อง ผิด ตัวเลือก D ไม่ถูกต้อง ผิด",
			},
		}
		store.questions[qid] = q
	}

	ec := NewExamControllerWithStore(store)
	r := makeExamRouter(ec)
	tok := signExamJWT(t, "u-bilingual-1")
	headers := map[string]string{
		"Authorization":   "Bearer " + tok,
		"Accept-Language": "th",
	}

	// 1. Sample endpoint includes full translation including explanation
	wSample := doJSONReq(t, r, http.MethodGet, "/api/exams/sample", nil, headers)
	if wSample.Code != http.StatusOK {
		t.Fatalf("sample failed: %d", wSample.Code)
	}
	var sampleBody map[string]any
	_ = json.Unmarshal(wSample.Body.Bytes(), &sampleBody)
	sampleTh := sampleBody["translations"].(map[string]any)["th"].(map[string]any)
	sampleExp, hasSampleExp := sampleTh["explanation"].(string)
	if !hasSampleExp || sampleExp == "" || containsThaiRune(sampleBody["question_text"].(string)) {
		t.Fatalf("sample should keep canonical EN stem and include th explanation: %+v", sampleBody)
	}
	if _, hasActive := sampleBody["is_active"]; hasActive {
		t.Fatalf("sample DTO must not leak internal is_active column: %+v", sampleBody)
	}

	// 2. Start Pre-Test & Session Recovery keep canonical EN and strip th.explanation
	wPre := doJSONReq(t, r, http.MethodPost, "/api/exams/pre-test", nil, headers)
	if wPre.Code != http.StatusCreated {
		t.Fatalf("pre-test start failed: %d", wPre.Code)
	}
	var preBody struct {
		Session   models.ExamSession `json:"session"`
		Questions []map[string]any   `json:"questions"`
	}
	_ = json.Unmarshal(wPre.Body.Bytes(), &preBody)
	for _, q := range preBody.Questions {
		if containsThaiRune(q["question_text"].(string)) {
			t.Fatalf("canonical question_text must stay English: %v", q["question_text"])
		}
		thMap := q["translations"].(map[string]any)["th"].(map[string]any)
		if thStem, ok := thMap["question_text"].(string); !ok || thStem == "" || thMap["options"] == nil {
			t.Fatalf("expected th stem and options in active exam, got %+v", thMap)
		}
		if _, leaked := thMap["explanation"]; leaked {
			t.Fatalf("active pre-test leaked translations.th.explanation: %+v", thMap)
		}
	}

	wSess := doJSONReq(t, r, http.MethodGet, "/api/exams/session", nil, headers)
	if wSess.Code != http.StatusOK {
		t.Fatalf("session recovery failed: %d", wSess.Code)
	}
	var sessBody struct {
		Questions []map[string]any `json:"questions"`
	}
	_ = json.Unmarshal(wSess.Body.Bytes(), &sessBody)
	thRec := sessBody.Questions[0]["translations"].(map[string]any)["th"].(map[string]any)
	if _, leaked := thRec["explanation"]; leaked {
		t.Fatalf("session recovery leaked translations.th.explanation")
	}

	// 3. Answer all 30 and check Review exposes full th.explanation
	for _, q := range preBody.Questions {
		wSub := doJSONReq(t, r, http.MethodPost, "/api/exams/submit-answer", map[string]any{
			"session_id":  preBody.Session.SessionID,
			"question_id": q["question_id"],
			"user_answer": "A",
		}, headers)
		if wSub.Code != http.StatusOK {
			t.Fatalf("submit failed: %d", wSub.Code)
		}
	}

	wRev := doJSONReq(t, r, http.MethodGet, "/api/exams/sessions/"+preBody.Session.SessionID+"/review", nil, headers)
	if wRev.Code != http.StatusOK {
		t.Fatalf("review failed: %d", wRev.Code)
	}
	var revBody struct {
		Results []map[string]any `json:"results"`
	}
	_ = json.Unmarshal(wRev.Body.Bytes(), &revBody)
	thRev := revBody.Results[0]["translations"].(map[string]any)["th"].(map[string]any)
	if thRev["explanation"] == nil || thRev["explanation"] == "" {
		t.Fatalf("expected th.explanation in exam review, got %+v", thRev)
	}
}

func TestExamController_FinalizeGuardReviewHistoryAndProficiencyInvariance(t *testing.T) {
	t.Setenv("JWT_SECRET", "exam-test-secret")
	store := newFakeExamStore()
	store.seedExamPool("rev", map[string]int{"domain1": 8, "domain2": 14, "domain3": 10, "domain4": 16, "domain5": 12})
	ec := NewExamControllerWithStore(store)
	r := makeExamRouter(ec)

	tokA := signExamJWT(t, "u-review-a")
	tokB := signExamJWT(t, "u-review-b")
	hA := map[string]string{"Authorization": "Bearer " + tokA}
	hB := map[string]string{"Authorization": "Bearer " + tokB}

	// 1. Unauthenticated access rejected with 401
	wUnauth := doJSONReq(t, r, http.MethodGet, "/api/exams/history", nil, nil)
	if wUnauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated history, got %d", wUnauth.Code)
	}

	// 2. Empty history initially
	wEmptyHist := doJSONReq(t, r, http.MethodGet, "/api/exams/history", nil, hA)
	if wEmptyHist.Code != http.StatusOK {
		t.Fatalf("expected 200 on empty history, got %d: %s", wEmptyHist.Code, wEmptyHist.Body.String())
	}
	var emptyHist dto.ExamHistoryResponse
	if err := json.Unmarshal(wEmptyHist.Body.Bytes(), &emptyHist); err != nil {
		t.Fatalf("unmarshal empty history: %v", err)
	}
	if len(emptyHist.Sessions) != 0 || emptyHist.LatestPreSessionID != nil || emptyHist.LatestPostSessionID != nil {
		t.Fatalf("expected empty history response, got %+v", emptyHist)
	}

	// 3. Start Pre-Test for user A
	wPre := doJSONReq(t, r, http.MethodPost, "/api/exams/pre-test", nil, hA)
	if wPre.Code != http.StatusCreated {
		t.Fatalf("expected 201 StartPreTest, got %d: %s", wPre.Code, wPre.Body.String())
	}
	var preStart struct {
		Session   models.ExamSession `json:"session"`
		Questions []QuestionResponse `json:"questions"`
	}
	_ = json.Unmarshal(wPre.Body.Bytes(), &preStart)
	preID := preStart.Session.SessionID

	// Shuffle stored sessionQs order to prove GetExamReview returns order_index ASC regardless of storage order
	storedPreQs := store.sessionQs[preID]
	for i, j := 0, len(storedPreQs)-1; i < j; i, j = i+1, j-1 {
		storedPreQs[i], storedPreQs[j] = storedPreQs[j], storedPreQs[i]
	}

	// 4. Completion guard: GET /review while in_progress returns 403 EXAM_NOT_COMPLETED with zero answer leakage
	wInProg := doJSONReq(t, r, http.MethodGet, "/api/exams/sessions/"+preID+"/review", nil, hA)
	if wInProg.Code != http.StatusForbidden {
		t.Fatalf("expected 403 EXAM_NOT_COMPLETED while in_progress, got %d: %s", wInProg.Code, wInProg.Body.String())
	}
	var inProgBody dto.ErrorResponse
	_ = json.Unmarshal(wInProg.Body.Bytes(), &inProgBody)
	if inProgBody.Code != "EXAM_NOT_COMPLETED" || inProgBody.Params["exam_type"] != "pre" {
		t.Fatalf("unexpected 403 body: %+v", inProgBody)
	}
	if strings.Contains(wInProg.Body.String(), "Correct Answer:") {
		t.Fatalf("answers leaked while exam in_progress")
	}

	// 5. Ownership guard: user B cannot inspect or finalize user A's session (404)
	wForeignRev := doJSONReq(t, r, http.MethodGet, "/api/exams/sessions/"+preID+"/review", nil, hB)
	if wForeignRev.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for foreign review access, got %d", wForeignRev.Code)
	}
	wForeignFin := doJSONReq(t, r, http.MethodPost, "/api/exams/sessions/"+preID+"/finalize", map[string]any{"time_spent_seconds": 120}, hB)
	if wForeignFin.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for foreign finalize access, got %d", wForeignFin.Code)
	}

	// 6. Partial submission: answer 18 correct ("A"), 10 wrong ("B"), leave 2 blank
	for idx, q := range preStart.Questions {
		if idx >= 28 {
			continue
		}
		ans := "A"
		if idx >= 18 {
			ans = "B"
		}
		wAns := doJSONReq(t, r, http.MethodPost, "/api/exams/submit-answer", map[string]any{
			"session_id":         preID,
			"question_id":        q.QuestionID,
			"user_answer":        ans,
			"time_spent_seconds": 0,
		}, hA)
		if wAns.Code != http.StatusOK {
			t.Fatalf("submit answer %d failed: %d", idx, wAns.Code)
		}
	}

	// Finalize with elapsed time 515s
	wFin := doJSONReq(t, r, http.MethodPost, "/api/exams/sessions/"+preID+"/finalize", map[string]any{
		"time_spent_seconds": 515,
	}, hA)
	if wFin.Code != http.StatusOK {
		t.Fatalf("expected 200 FinalizeSession, got %d: %s", wFin.Code, wFin.Body.String())
	}
	var finResp dto.FinalizeExamResponse
	if err := json.Unmarshal(wFin.Body.Bytes(), &finResp); err != nil {
		t.Fatalf("unmarshal finalize response: %v", err)
	}
	if !finResp.Completed || finResp.CorrectCount != 18 || finResp.IncorrectCount != 10 || finResp.UnansweredCount != 2 || finResp.Score != 60 || finResp.TimeSpentSeconds != 515 {
		t.Fatalf("unexpected partial finalize counts: %+v", finResp)
	}
	if len(finResp.DomainScores) != 5 || len(finResp.Results) != 30 {
		t.Fatalf("expected 5 domain scores and 30 results, got %d / %d", len(finResp.DomainScores), len(finResp.Results))
	}
	if !finResp.Results[28].IsUnanswered || !finResp.Results[29].IsUnanswered || finResp.Results[28].IsCorrect {
		t.Fatalf("expected questions 29-30 marked unanswered & incorrect")
	}

	// Idempotent finalize call preserves completed_at & score
	firstCompletedAt := *store.sessions[preID].CompletedAt
	wFin2 := doJSONReq(t, r, http.MethodPost, "/api/exams/sessions/"+preID+"/finalize", map[string]any{
		"time_spent_seconds": 999,
	}, hA)
	if wFin2.Code != http.StatusOK {
		t.Fatalf("expected 200 idempotent finalize, got %d", wFin2.Code)
	}
	if !store.sessions[preID].CompletedAt.Equal(firstCompletedAt) || *store.sessions[preID].TimeSpentSeconds != 515 {
		t.Fatalf("idempotent finalize mutated completed_at or time_spent_seconds")
	}

	// 7. GET /review on completed Pre-Test verifies order_index ASC, domain order, sub_topic_title, nil pre_comparison
	wPreRev := doJSONReq(t, r, http.MethodGet, "/api/exams/sessions/"+preID+"/review", nil, hA)
	if wPreRev.Code != http.StatusOK {
		t.Fatalf("expected 200 Pre-Test review, got %d: %s", wPreRev.Code, wPreRev.Body.String())
	}
	var preRev dto.ExamReviewResponse
	if err := json.Unmarshal(wPreRev.Body.Bytes(), &preRev); err != nil {
		t.Fatalf("unmarshal pre review: %v", err)
	}
	if preRev.PreComparison != nil || len(preRev.Results) != 30 || len(preRev.DomainScores) != 5 {
		t.Fatalf("unexpected pre review envelope: preComp=%v res=%d dom=%d", preRev.PreComparison, len(preRev.Results), len(preRev.DomainScores))
	}
	for i, item := range preRev.Results {
		if item.OrderIndex != i {
			t.Fatalf("results not sorted by order_index ASC at %d: got %d", i, item.OrderIndex)
		}
		if item.SubTopicID == nil || item.SubTopicTitle == "" || item.DomainName == "" {
			t.Fatalf("missing domain/subtopic enrichment on item %d: %+v", i, item)
		}
	}
	for i, ds := range preRev.DomainScores {
		if ds.DomainID != services.OrderedTheoryDomains[i].DomainID {
			t.Fatalf("domain_scores not in canonical order at %d: got %s", i, ds.DomainID)
		}
	}

	// 8. Simulate simulation training raising user_skill_profiles to 82.5 across all 5 domains
	for i := range store.profiles["u-review-a"] {
		store.profiles["u-review-a"][i].ProficiencyScore = 82.5
	}

	// Start Post-Test and complete with 27/30 correct (90%) via SubmitAnswer
	wPost := doJSONReq(t, r, http.MethodPost, "/api/exams/post-test", nil, hA)
	if wPost.Code != http.StatusCreated {
		t.Fatalf("expected 201 StartPostTest, got %d: %s", wPost.Code, wPost.Body.String())
	}
	var postStart struct {
		Session   models.ExamSession `json:"session"`
		Questions []QuestionResponse `json:"questions"`
	}
	_ = json.Unmarshal(wPost.Body.Bytes(), &postStart)
	postID := postStart.Session.SessionID

	var lastPostSubmit dto.SubmitAnswerResponse
	for idx, q := range postStart.Questions {
		ans := "A"
		if idx >= 27 {
			ans = "C"
		}
		timeInc := 0
		if idx == len(postStart.Questions)-1 {
			timeInc = 420
		}
		wSub := doJSONReq(t, r, http.MethodPost, "/api/exams/submit-answer", map[string]any{
			"session_id":         postID,
			"question_id":        q.QuestionID,
			"user_answer":        ans,
			"time_spent_seconds": timeInc,
		}, hA)
		if wSub.Code != http.StatusOK {
			t.Fatalf("post submit %d failed: %d", idx, wSub.Code)
		}
		_ = json.Unmarshal(wSub.Body.Bytes(), &lastPostSubmit)
	}
	if !lastPostSubmit.Completed || len(lastPostSubmit.DomainScores) != 5 || lastPostSubmit.SessionID != postID {
		t.Fatalf("expected completed SubmitAnswerResponse with 5 domain_scores and session_id, got %+v", lastPostSubmit)
	}

	// Proficiency Invariance: user_skill_profiles must remain 82.5 across all 5 domains after Post-Test!
	for _, prof := range store.profiles["u-review-a"] {
		if prof.ProficiencyScore != 82.5 {
			t.Fatalf("Proficiency Invariance violated on %s: got %v, want 82.5", prof.DomainID, prof.ProficiencyScore)
		}
	}

	// 9. Post-Test review returns pre_comparison against latest completed Pre-Test
	wPostRev := doJSONReq(t, r, http.MethodGet, "/api/exams/sessions/"+postID+"/review", nil, hA)
	if wPostRev.Code != http.StatusOK {
		t.Fatalf("expected 200 Post-Test review, got %d: %s", wPostRev.Code, wPostRev.Body.String())
	}
	var postRev dto.ExamReviewResponse
	if err := json.Unmarshal(wPostRev.Body.Bytes(), &postRev); err != nil {
		t.Fatalf("unmarshal post review: %v", err)
	}
	if postRev.PreComparison == nil {
		t.Fatalf("expected non-nil pre_comparison on post-test review")
	}
	if postRev.PreComparison.PreSessionID != preID || postRev.PreComparison.PreScore != 60 || postRev.PreComparison.PostScore != 90 || postRev.PreComparison.ScoreDelta != 30 {
		t.Fatalf("unexpected pre_comparison values: %+v", postRev.PreComparison)
	}
	if len(postRev.PreComparison.DomainDeltas) != 5 {
		t.Fatalf("expected 5 domain_deltas, got %d", len(postRev.PreComparison.DomainDeltas))
	}

	// 10. History excludes abandoned/in_progress sessions and lists completed post + pre DESC
	store.sessions["abandoned-x"] = &models.ExamSession{
		SessionID:      "abandoned-x",
		UserID:         "u-review-a",
		ExamType:       "pre",
		Status:         "abandoned",
		TotalQuestions: 30,
		StartedAt:      time.Now(),
	}
	wHist := doJSONReq(t, r, http.MethodGet, "/api/exams/history", nil, hA)
	if wHist.Code != http.StatusOK {
		t.Fatalf("expected 200 history, got %d", wHist.Code)
	}
	var hist dto.ExamHistoryResponse
	_ = json.Unmarshal(wHist.Body.Bytes(), &hist)
	if len(hist.Sessions) != 2 || hist.LatestPostSessionID == nil || *hist.LatestPostSessionID != postID || hist.LatestPreSessionID == nil || *hist.LatestPreSessionID != preID {
		t.Fatalf("unexpected exam history: %+v", hist)
	}
	if hist.LatestPreToPostDelta == nil || *hist.LatestPreToPostDelta != 30 {
		t.Fatalf("expected server-computed latest_pre_to_post_delta=30, got %v", hist.LatestPreToPostDelta)
	}
	if hist.Sessions[0].SessionID != postID || hist.Sessions[1].SessionID != preID {
		t.Fatalf("expected DESC completed ordering (post then pre), got %+v", hist.Sessions)
	}
}

func TestExamController_PostReviewWithoutPreTestReturnsNilComparison(t *testing.T) {
	t.Setenv("JWT_SECRET", "exam-test-secret")
	store := newFakeExamStore()
	store.seedExamPool("solo", map[string]int{"domain1": 4, "domain2": 7, "domain3": 5, "domain4": 8, "domain5": 6})
	ec := NewExamControllerWithStore(store)
	r := makeExamRouter(ec)

	tok := signExamJWT(t, "u-no-pre")
	now := time.Now()
	score := 80.0
	store.sessions["post-only"] = &models.ExamSession{
		SessionID:      "post-only",
		UserID:         "u-no-pre",
		ExamType:       "post",
		Status:         "completed",
		Score:          &score,
		TotalQuestions: 30,
		CompletedAt:    &now,
	}
	var esqs []models.ExamSessionQuestion
	idx := 0
	for qid := range store.questions {
		trueVal := true
		esqs = append(esqs, models.ExamSessionQuestion{
			SessionID:  "post-only",
			QuestionID: qid,
			OrderIndex: idx,
			UserAnswer: "A",
			IsCorrect:  &trueVal,
		})
		idx++
	}
	store.sessionQs["post-only"] = esqs

	w := doJSONReq(t, r, http.MethodGet, "/api/exams/sessions/post-only/review", nil, map[string]string{
		"Authorization": "Bearer " + tok,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var rev dto.ExamReviewResponse
	_ = json.Unmarshal(w.Body.Bytes(), &rev)
	if rev.PreComparison != nil {
		t.Fatalf("expected nil pre_comparison when no completed pre-test exists, got %+v", rev.PreComparison)
	}
}


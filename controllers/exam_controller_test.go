package controllers

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
	"unicode"

	"cybersim/middlewares"
	"cybersim/models"

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
	for dom, n := range countPerDomain {
		for i := 0; i < n; i++ {
			qid := prefix + "-" + dom + "-" + string(rune('a'+i))
			f.questions[qid] = models.Question{
				QuestionID:    qid,
				DomainID:      dom,
				Type:          "multiple_choice",
				QuestionText:  "Which control mitigates " + dom + " risk item " + qid + "?",
				Options:       map[string]string{"A": "Control A", "B": "Control B", "C": "Control C", "D": "Control D"},
				CorrectAnswer: "A",
				Explanation:   "Control A provides primary defense for " + dom + ".",
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

func (f *fakeExamStore) FinalizeExamSession(session *models.ExamSession, userID string) (int64, []ExamDomainScore, []ExamDetailedResult, error) {
	esqs := f.sessionQs[session.SessionID]
	var correctCount int64
	domTotals := make(map[string]int64)
	domCorrect := make(map[string]int64)
	detailed := make([]ExamDetailedResult, 0, len(esqs))

	for _, esq := range esqs {
		q := f.questions[esq.QuestionID]
		isCorr := esq.IsCorrect != nil && *esq.IsCorrect
		if isCorr {
			correctCount++
			domCorrect[q.DomainID]++
		}
		domTotals[q.DomainID]++
		detailed = append(detailed, ExamDetailedResult{
			Question:   q,
			UserAnswer: esq.UserAnswer,
			IsCorrect:  isCorr,
		})
	}

	scorePct := float64(correctCount) / float64(session.TotalQuestions) * 100
	now := time.Now()
	if stored, ok := f.sessions[session.SessionID]; ok {
		stored.Score = &scorePct
		stored.Status = "completed"
		stored.CompletedAt = &now
		*session = *stored
	}

	var scores []ExamDomainScore
	if session.ExamType == "pre" {
		f.profiles[userID] = nil
		for dom, tot := range domTotals {
			corr := domCorrect[dom]
			pct := float64(corr) / float64(tot) * 100
			scores = append(scores, ExamDomainScore{
				DomainID:     dom,
				TotalCount:   tot,
				CorrectCount: corr,
				Percentage:   pct,
			})
			f.profiles[userID] = append(f.profiles[userID], models.UserSkillProfile{
				UserID:           userID,
				DomainID:         dom,
				ProficiencyScore: pct,
			})
		}
	}

	return correctCount, scores, detailed, nil
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

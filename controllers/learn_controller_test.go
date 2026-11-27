package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cybersim/dto"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type fakeLearnStore struct {
	profiles          map[string][]models.UserSkillProfile
	progress          map[string]map[string]models.UserTheoryProgress
	subProgress       map[string]map[string]models.UserTheorySubtopicProgress
	sampleBySubtopic  map[string][]models.Question
	sampleByDomain    map[string][]models.Question
	sampleAssignments map[string]map[string][]models.UserTheorySampleQuestionDetail
	preferredLangs    map[string]string
}

func newFakeLearnStore() *fakeLearnStore {
	return &fakeLearnStore{
		profiles:          make(map[string][]models.UserSkillProfile),
		progress:          make(map[string]map[string]models.UserTheoryProgress),
		subProgress:       make(map[string]map[string]models.UserTheorySubtopicProgress),
		sampleBySubtopic:  make(map[string][]models.Question),
		sampleByDomain:    make(map[string][]models.Question),
		sampleAssignments: make(map[string]map[string][]models.UserTheorySampleQuestionDetail),
	}
}

func (f *fakeLearnStore) seedSamplePool(domainID, subTopicID string, qs []models.Question) {
	for i := range qs {
		st := subTopicID
		qs[i].DomainID = domainID
		qs[i].SubTopicID = &st
	}
	f.sampleBySubtopic[subTopicID] = append(f.sampleBySubtopic[subTopicID], qs...)
	f.sampleByDomain[domainID] = append(f.sampleByDomain[domainID], qs...)
}

func (f *fakeLearnStore) GetUserSkillProfiles(userID string) ([]models.UserSkillProfile, error) {
	return f.profiles[userID], nil
}

func (f *fakeLearnStore) GetTheoryProgress(userID string) ([]models.UserTheoryProgress, error) {
	m := f.progress[userID]
	out := make([]models.UserTheoryProgress, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out, nil
}

func (f *fakeLearnStore) GetSubtopicProgress(userID string) ([]models.UserTheorySubtopicProgress, error) {
	m := f.subProgress[userID]
	out := make([]models.UserTheorySubtopicProgress, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out, nil
}

func (f *fakeLearnStore) GetSubtopicSampleProgress(userID string) (map[string]int, map[string]int, error) {
	attempted := make(map[string]int)
	total := make(map[string]int)
	for subID, items := range f.sampleAssignments[userID] {
		total[subID] = len(items)
		for _, item := range items {
			if item.Assignment.SubmittedAnswer != nil && *item.Assignment.SubmittedAnswer != "" {
				attempted[subID]++
			}
		}
	}
	return attempted, total, nil
}

func (f *fakeLearnStore) MarkTheoryCompleted(userID, domainID string) error {
	if _, ok := f.progress[userID]; !ok {
		f.progress[userID] = make(map[string]models.UserTheoryProgress)
	}
	if _, exists := f.progress[userID][domainID]; exists {
		return nil // idempotent
	}
	f.progress[userID][domainID] = models.UserTheoryProgress{
		UserID:      userID,
		DomainID:    domainID,
		CompletedAt: time.Now(),
	}
	return nil
}

func (f *fakeLearnStore) CompleteSubtopic(userID, role, subTopicID, domainID string) (dto.CompleteSubtopicResponse, error) {
	specList := services.SubtopicsForDomain(domainID)
	if role != "learner" {
		return dto.CompleteSubtopicResponse{
			SubTopicID:                 subTopicID,
			DomainID:                   domainID,
			IsCompleted:                true,
			DomainAutoCompleted:        false,
			TotalSubtopicsInDomain:     len(specList),
			CompletedSubtopicCount:     0,
			TotalSubtopics:             len(services.OrderedTheorySubtopics),
			CompletedSubtopics:         []string{},
			CompletedDomains:           []string{},
			TotalDomains:               len(services.OrderedTheoryDomains),
			TheoryCompleted:            false,
			PreviewMode:                true,
		}, nil
	}

	items := f.sampleAssignments[userID][subTopicID]
	attempted := 0
	for _, it := range items {
		if it.Assignment.SubmittedAnswer != nil && *it.Assignment.SubmittedAnswer != "" {
			attempted++
		}
	}
	if len(items) == 0 || attempted < len(items) {
		return dto.CompleteSubtopicResponse{}, errSampleIncomplete
	}

	if _, ok := f.subProgress[userID]; !ok {
		f.subProgress[userID] = make(map[string]models.UserTheorySubtopicProgress)
	}
	subNewlyCompleted := false
	if _, exists := f.subProgress[userID][subTopicID]; !exists {
		f.subProgress[userID][subTopicID] = models.UserTheorySubtopicProgress{
			UserID:      userID,
			SubTopicID:  subTopicID,
			CompletedAt: time.Now(),
		}
		subNewlyCompleted = true
	}

	completedSubInDomain := 0
	for _, sp := range specList {
		if _, ok := f.subProgress[userID][sp.SubTopicID]; ok {
			completedSubInDomain++
		}
	}
	domainAutoCompleted := false
	if len(specList) > 0 && completedSubInDomain == len(specList) {
		if subNewlyCompleted {
			if _, ok := f.progress[userID]; !ok {
				f.progress[userID] = make(map[string]models.UserTheoryProgress)
			}
			f.progress[userID][domainID] = models.UserTheoryProgress{
				UserID:      userID,
				DomainID:    domainID,
				CompletedAt: time.Now(),
			}
			domainAutoCompleted = true
		} else if _, exists := f.progress[userID][domainID]; !exists {
			_ = f.MarkTheoryCompleted(userID, domainID)
			domainAutoCompleted = true
		}
	}

	completedSub := make([]string, 0, len(f.subProgress[userID]))
	for _, s := range services.OrderedTheorySubtopics {
		if _, ok := f.subProgress[userID][s.SubTopicID]; ok {
			completedSub = append(completedSub, s.SubTopicID)
		}
	}
	progRows, _ := f.GetTheoryProgress(userID)
	completedDomains, _ := buildEffectiveCompletedDomains(progRows)

	return dto.CompleteSubtopicResponse{
		SubTopicID:                 subTopicID,
		DomainID:                   domainID,
		IsCompleted:                true,
		DomainAutoCompleted:        domainAutoCompleted,
		CompletedSubtopicsInDomain: completedSubInDomain,
		TotalSubtopicsInDomain:     len(specList),
		CompletedSubtopicCount:     len(completedSub),
		TotalSubtopics:             len(services.OrderedTheorySubtopics),
		CompletedSubtopics:         completedSub,
		CompletedDomains:           completedDomains,
		CompletedCount:             len(completedDomains),
		TotalDomains:               len(services.OrderedTheoryDomains),
		TheoryCompleted:            services.IsTheoryGateSatisfied(completedDomains),
		PreviewMode:                false,
	}, nil
}

func (f *fakeLearnStore) LoadModules() ([]services.TheoryModuleContent, error) {
	return services.LoadTheoryModules(nil)
}

func (f *fakeLearnStore) LoadSubtopics() ([]services.TheorySubtopicContent, error) {
	return services.LoadTheorySubtopics(nil)
}

func (f *fakeLearnStore) LoadModulesByLocale(locale string) ([]services.TheoryModuleContent, error) {
	return services.LoadTheoryModulesByLocale(nil, locale)
}

func (f *fakeLearnStore) LoadSubtopicsByLocale(locale string) ([]services.TheorySubtopicContent, error) {
	return services.LoadTheorySubtopicsByLocale(nil, locale)
}

func (f *fakeLearnStore) GetUserPreferredLanguage(userID string) string {
	if lang, ok := f.preferredLangs[userID]; ok {
		return lang
	}
	return "en"
}

func (f *fakeLearnStore) GetOrAssignSampleQuestions(userID, role, subTopicID, domainID string) ([]models.UserTheorySampleQuestionDetail, error) {
	if role != "learner" {
		return f.drawSamples(subTopicID, domainID, nil), nil
	}
	if _, ok := f.sampleAssignments[userID]; !ok {
		f.sampleAssignments[userID] = make(map[string][]models.UserTheorySampleQuestionDetail)
	}
	if existing, ok := f.sampleAssignments[userID][subTopicID]; ok && len(existing) > 0 {
		out := make([]models.UserTheorySampleQuestionDetail, len(existing))
		copy(out, existing)
		return out, nil
	}
	used := make(map[string]bool)
	for _, items := range f.sampleAssignments[userID] {
		for _, item := range items {
			used[item.Assignment.QuestionID] = true
		}
	}
	assigned := f.drawSamples(subTopicID, domainID, used)
	for i := range assigned {
		assigned[i].Assignment.UserID = userID
	}
	f.sampleAssignments[userID][subTopicID] = assigned
	out := make([]models.UserTheorySampleQuestionDetail, len(assigned))
	copy(out, assigned)
	return out, nil
}

func (f *fakeLearnStore) drawSamples(subTopicID, domainID string, used map[string]bool) []models.UserTheorySampleQuestionDetail {
	var details []models.UserTheorySampleQuestionDetail
	picked := make(map[string]bool)
	order := 1
	for _, q := range f.sampleBySubtopic[subTopicID] {
		if len(details) >= 2 {
			break
		}
		if used != nil && used[q.QuestionID] {
			continue
		}
		picked[q.QuestionID] = true
		details = append(details, models.UserTheorySampleQuestionDetail{
			Assignment: models.UserTheorySampleQuestion{
				SubTopicID: subTopicID,
				QuestionID: q.QuestionID,
				OrderIndex: order,
				IsFallback: false,
			},
			Question: q,
		})
		order++
	}
	for _, q := range f.sampleByDomain[domainID] {
		if len(details) >= 2 {
			break
		}
		if picked[q.QuestionID] || (used != nil && used[q.QuestionID]) {
			continue
		}
		picked[q.QuestionID] = true
		details = append(details, models.UserTheorySampleQuestionDetail{
			Assignment: models.UserTheorySampleQuestion{
				SubTopicID: subTopicID,
				QuestionID: q.QuestionID,
				OrderIndex: order,
				IsFallback: true,
			},
			Question: q,
		})
		order++
	}
	return details
}

func (f *fakeLearnStore) SubmitSampleQuestionAttempt(userID, role, subTopicID, questionID, answer string) (dto.AttemptSampleQuestionResponse, error) {
	if role != "learner" {
		for _, q := range f.sampleByDomain {
			for _, item := range q {
				if item.QuestionID == questionID {
					return dto.AttemptSampleQuestionResponse{
						SubTopicID:      subTopicID,
						QuestionID:      questionID,
						SubmittedAnswer: answer,
						IsCorrect:       strings.EqualFold(answer, item.CorrectAnswer),
						CorrectAnswer:   item.CorrectAnswer,
						Explanation:     item.Explanation,
						Translations:    services.SanitizeQuestionTranslations(item.Translations, true),
						AttemptedCount:  2,
						RequiredCount:   2,
						ReadyToComplete: true,
						PreviewMode:     true,
					}, nil
				}
			}
		}
		return dto.AttemptSampleQuestionResponse{}, errSampleNotFound
	}

	items := f.sampleAssignments[userID][subTopicID]
	targetIdx := -1
	for i := range items {
		if items[i].Assignment.QuestionID == questionID {
			targetIdx = i
			break
		}
	}
	if targetIdx < 0 {
		return dto.AttemptSampleQuestionResponse{}, errSampleNotFound
	}

	item := &items[targetIdx]
	already := item.Assignment.SubmittedAnswer != nil && *item.Assignment.SubmittedAnswer != ""
	if !already {
		ansCopy := answer
		isCorr := strings.EqualFold(answer, item.Question.CorrectAnswer)
		now := time.Now()
		item.Assignment.SubmittedAnswer = &ansCopy
		item.Assignment.IsCorrect = &isCorr
		item.Assignment.AnsweredAt = &now
	}

	attempted := 0
	for _, it := range items {
		if it.Assignment.SubmittedAnswer != nil && *it.Assignment.SubmittedAnswer != "" {
			attempted++
		}
	}
	reqCount := len(items)
	isCorrVal := item.Assignment.IsCorrect != nil && *item.Assignment.IsCorrect
	return dto.AttemptSampleQuestionResponse{
		SubTopicID:       subTopicID,
		QuestionID:       questionID,
		SubmittedAnswer:  *item.Assignment.SubmittedAnswer,
		IsCorrect:        isCorrVal,
		CorrectAnswer:    item.Question.CorrectAnswer,
		Explanation:      item.Question.Explanation,
		Translations:     services.SanitizeQuestionTranslations(item.Question.Translations, true),
		AttemptedCount:   attempted,
		RequiredCount:    reqCount,
		ReadyToComplete:  reqCount > 0 && attempted >= reqCount,
		AlreadySubmitted: already,
		PreviewMode:      false,
	}, nil
}

func setupLearnRouter(lc *LearnController, userID, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if userID != "" {
			c.Set("user_id", userID)
		}
		if role != "" {
			c.Set("role", role)
		}
		c.Next()
	})
	r.GET("/api/learn/domains", lc.GetDomains)
	r.POST("/api/learn/domains/:domainId/complete", lc.CompleteDomain)
	r.GET("/api/learn/subtopics/:subTopicId", lc.GetSubTopic)
	r.POST("/api/learn/subtopics/:subTopicId/sample-questions/:questionId/attempt", lc.AttemptSampleQuestion)
	r.POST("/api/learn/subtopics/:subTopicId/complete", lc.CompleteSubTopic)
	return r
}

func TestLearnController_PreTestGateAndProgression(t *testing.T) {
	store := newFakeLearnStore()
	lc := NewLearnControllerWithStore(store)

	t.Run("learner before pre-test receives 403 pre_test_required on GET and POST", func(t *testing.T) {
		r := setupLearnRouter(lc, "learner-1", "learner")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/learn/domains", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("GET /api/learn/domains status = %d, want 403", w.Code)
		}
		var errResp dto.TheoryGateErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if errResp.Reason != "pre_test_required" || errResp.Unlocked == nil || *errResp.Unlocked {
			t.Fatalf("unexpected error payload: %+v", errResp)
		}

		wPost := httptest.NewRecorder()
		reqPost, _ := http.NewRequest("POST", "/api/learn/domains/domain1/complete", nil)
		r.ServeHTTP(wPost, reqPost)
		if wPost.Code != http.StatusForbidden {
			t.Fatalf("POST /api/learn/domains/domain1/complete status = %d, want 403", wPost.Code)
		}
	})

	t.Run("specialist without pre-test can access all 5 theory modules", func(t *testing.T) {
		r := setupLearnRouter(lc, "spec-1", "specialist")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/learn/domains", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("GET /api/learn/domains status = %d, want 200", w.Code)
		}
		var resp dto.LearnDomainsResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if !resp.Unlocked || len(resp.Domains) != 5 {
			t.Fatalf("expected unlocked=true and 5 domains, got unlocked=%v len=%d", resp.Unlocked, len(resp.Domains))
		}
	})

	t.Run("learner after pre-test sees proficiency, recommended focus, and completes 5/5 idempotently", func(t *testing.T) {
		store.profiles["learner-2"] = []models.UserSkillProfile{
			{UserID: "learner-2", DomainID: "domain1", ProficiencyScore: 75.0},
			{UserID: "learner-2", DomainID: "domain2", ProficiencyScore: 55.0},
			{UserID: "learner-2", DomainID: "domain3", ProficiencyScore: 80.0},
			{UserID: "learner-2", DomainID: "domain4", ProficiencyScore: 40.0},
			{UserID: "learner-2", DomainID: "domain5", ProficiencyScore: 72.0},
		}
		r := setupLearnRouter(lc, "learner-2", "learner")

		// Invalid domain ID rejected with 400
		wBad := httptest.NewRecorder()
		reqBad, _ := http.NewRequest("POST", "/api/learn/domains/domain99/complete", nil)
		r.ServeHTTP(wBad, reqBad)
		if wBad.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid domainId, got %d", wBad.Code)
		}

		// Seed 2 sample questions for each of the 28 canonical sub-topics
		for _, spec := range services.OrderedTheorySubtopics {
			store.seedSamplePool(spec.DomainID, spec.SubTopicID, []models.Question{
				{
					QuestionID:    "q-" + spec.SubTopicID + "-1",
					DomainID:      spec.DomainID,
					Type:          "multiple_choice",
					QuestionText:  "Question 1 for " + spec.SubTopicID,
					Options:       map[string]string{"A": "OptA", "B": "OptB", "C": "OptC", "D": "OptD"},
					CorrectAnswer: "A",
					Explanation:   "Exp 1",
					IsActive:      true,
				},
				{
					QuestionID:    "q-" + spec.SubTopicID + "-2",
					DomainID:      spec.DomainID,
					Type:          "multiple_choice",
					QuestionText:  "Question 2 for " + spec.SubTopicID,
					Options:       map[string]string{"A": "OptA", "B": "OptB", "C": "OptC", "D": "OptD"},
					CorrectAnswer: "B",
					Explanation:   "Exp 2",
					IsActive:      true,
				},
			})
		}

		// Direct domain completion works immediately without completing child sub-topics (idempotent + Staff preview_mode)
		wDirect := httptest.NewRecorder()
		reqDirect, _ := http.NewRequest("POST", "/api/learn/domains/domain2/complete", nil)
		r.ServeHTTP(wDirect, reqDirect)
		if wDirect.Code != http.StatusOK {
			t.Fatalf("expected 200 on direct POST /domains/domain2/complete, got %d: %s", wDirect.Code, wDirect.Body.String())
		}
		var directResp dto.CompleteTheoryDomainResponse
		if err := json.Unmarshal(wDirect.Body.Bytes(), &directResp); err != nil {
			t.Fatalf("unmarshal directResp: %v", err)
		}
		if !directResp.IsCompleted || directResp.CompletedCount != 1 || directResp.TotalDomains != 5 || directResp.TheoryCompleted {
			t.Fatalf("unexpected direct domain2 complete payload: %+v", directResp)
		}

		// Staff direct domain complete returns preview_mode=true without mutating learner state
		rSpec := setupLearnRouter(lc, "spec-1", "specialist")
		wSpec := httptest.NewRecorder()
		reqSpec, _ := http.NewRequest("POST", "/api/learn/domains/domain1/complete", nil)
		rSpec.ServeHTTP(wSpec, reqSpec)
		if wSpec.Code != http.StatusOK {
			t.Fatalf("expected 200 on staff direct domain complete, got %d", wSpec.Code)
		}
		var specDone dto.CompleteTheoryDomainResponse
		_ = json.Unmarshal(wSpec.Body.Bytes(), &specDone)
		if !specDone.PreviewMode || !specDone.IsCompleted {
			t.Fatalf("expected staff preview_mode=true on domain complete, got %+v", specDone)
		}

		// Optional Deep Dive: complete sub-topic 1.1 still requires attempting its 2 sample questions first (400 otherwise)
		wSubPremature := httptest.NewRecorder()
		reqSubPremature, _ := http.NewRequest("POST", "/api/learn/subtopics/1.1/complete", nil)
		r.ServeHTTP(wSubPremature, reqSubPremature)
		if wSubPremature.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 sample_questions_incomplete before attempts, got %d: %s", wSubPremature.Code, wSubPremature.Body.String())
		}

		// Complete remaining domains directly (domain1, 3, 4, 5) alongside Optional Deep Dive auto-completion
		for _, spec := range services.OrderedTheorySubtopics {
			wSubGet := httptest.NewRecorder()
			reqSubGet, _ := http.NewRequest("GET", "/api/learn/subtopics/"+spec.SubTopicID, nil)
			r.ServeHTTP(wSubGet, reqSubGet)
			if wSubGet.Code != http.StatusOK {
				t.Fatalf("GET subtopic %s failed: %d", spec.SubTopicID, wSubGet.Code)
			}
			var subDetail dto.LearnSubtopicDetailResponse
			_ = json.Unmarshal(wSubGet.Body.Bytes(), &subDetail)
			for _, sq := range subDetail.SampleQuestions {
				ansBody, _ := json.Marshal(dto.AttemptSampleQuestionRequest{Answer: "A"})
				wAtt := httptest.NewRecorder()
				reqAtt, _ := http.NewRequest("POST", "/api/learn/subtopics/"+spec.SubTopicID+"/sample-questions/"+sq.QuestionID+"/attempt", bytes.NewReader(ansBody))
				reqAtt.Header.Set("Content-Type", "application/json")
				r.ServeHTTP(wAtt, reqAtt)
				if wAtt.Code != http.StatusOK {
					t.Fatalf("attempt %s/%s failed: %d", spec.SubTopicID, sq.QuestionID, wAtt.Code)
				}
			}

			wSubDone := httptest.NewRecorder()
			reqSubDone, _ := http.NewRequest("POST", "/api/learn/subtopics/"+spec.SubTopicID+"/complete", nil)
			r.ServeHTTP(wSubDone, reqSubDone)
			if wSubDone.Code != http.StatusOK {
				t.Fatalf("complete subtopic %s status = %d, want 200: %s", spec.SubTopicID, wSubDone.Code, wSubDone.Body.String())
			}
			if spec.SubTopicID == "1.4" {
				var c14 dto.CompleteSubtopicResponse
				_ = json.Unmarshal(wSubDone.Body.Bytes(), &c14)
				if !c14.DomainAutoCompleted || c14.CompletedCount != 2 || c14.TheoryCompleted {
					t.Fatalf("expected 1.4 completion to auto-complete domain1 alongside domain2 (2/5), got: %+v", c14)
				}
			}
		}

		// Idempotent POST /domains/domain1/complete stays 200 and keeps 5/5
		wComp := httptest.NewRecorder()
		reqComp, _ := http.NewRequest("POST", "/api/learn/domains/domain1/complete", nil)
		r.ServeHTTP(wComp, reqComp)
		if wComp.Code != http.StatusOK {
			t.Fatalf("POST /domains/domain1/complete after 4/4 subtopics = %d, want 200", wComp.Code)
		}

		wGet := httptest.NewRecorder()
		reqGet, _ := http.NewRequest("GET", "/api/learn/domains", nil)
		r.ServeHTTP(wGet, reqGet)
		if wGet.Code != http.StatusOK {
			t.Fatalf("GET /api/learn/domains status = %d, want 200", wGet.Code)
		}
		var resp dto.LearnDomainsResponse
		if err := json.Unmarshal(wGet.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if !resp.TheoryCompleted || resp.CompletedCount != 5 || resp.CompletedSubtopicCount != 28 || len(resp.CompletedDomains) != 5 {
			t.Fatalf("expected 5/5 domains & 28/28 subtopics theory_completed=true, got %+v", resp)
		}
		if !resp.Domains[1].RecommendedFocus || !resp.Domains[3].RecommendedFocus || resp.Domains[0].RecommendedFocus {
			t.Fatalf("unexpected recommended_focus flags: d1=%v d2=%v d4=%v",
				resp.Domains[0].RecommendedFocus, resp.Domains[1].RecommendedFocus, resp.Domains[3].RecommendedFocus)
		}
		// Proficiency score must remain untouched
		if resp.Domains[3].ProficiencyScore != 40.0 {
			t.Fatalf("proficiency score drifted after theory completion: got %v, want 40.0", resp.Domains[3].ProficiencyScore)
		}
	})
}

func TestLearnController_SubtopicsAndSampleQuestionLock(t *testing.T) {
	store := newFakeLearnStore()
	store.profiles["learner-sub"] = []models.UserSkillProfile{
		{UserID: "learner-sub", DomainID: "domain1", ProficiencyScore: 60.0},
	}
	store.seedSamplePool("domain1", "1.1", []models.Question{
		{
			QuestionID:    "q-11-a",
			DomainID:      "domain1",
			Type:          "multiple_choice",
			QuestionText:  "Control A?",
			Options:       map[string]string{"A": "Prev", "B": "Det", "C": "Corr", "D": "Comp"},
			CorrectAnswer: "A",
			Explanation:   "Preventive stops before impact.",
			IsActive:      true,
		},
		{
			QuestionID:    "q-11-b",
			DomainID:      "domain1",
			Type:          "multiple_choice",
			QuestionText:  "Control B?",
			Options:       map[string]string{"A": "Tech", "B": "Phys", "C": "Admin", "D": "Ops"},
			CorrectAnswer: "B",
			Explanation:   "Physical badge readers.",
			IsActive:      true,
		},
	})
	// Only 1 question in 1.3; domain fallback must supply 2nd from unused domain1 pool
	store.seedSamplePool("domain1", "1.3", []models.Question{
		{
			QuestionID:    "q-13-a",
			DomainID:      "domain1",
			Type:          "multiple_choice",
			QuestionText:  "Change board?",
			Options:       map[string]string{"A": "CAB", "B": "SOC", "C": "NOC", "D": "HR"},
			CorrectAnswer: "A",
			Explanation:   "CAB reviews changes.",
			IsActive:      true,
		},
	})
	store.seedSamplePool("domain1", "1.4", []models.Question{
		{
			QuestionID:    "q-14-fallback",
			DomainID:      "domain1",
			Type:          "multiple_choice",
			QuestionText:  "HSM root?",
			Options:       map[string]string{"A": "HSM", "B": "TLS", "C": "SHA", "D": "DNS"},
			CorrectAnswer: "A",
			Explanation:   "HSM secures CA keys.",
			IsActive:      true,
		},
	})

	lc := NewLearnControllerWithStore(store)
	r := setupLearnRouter(lc, "learner-sub", "learner")

	// 1. GET /api/learn/domains includes nested 28 sub-topics metadata
	wDom := httptest.NewRecorder()
	reqDom, _ := http.NewRequest("GET", "/api/learn/domains", nil)
	r.ServeHTTP(wDom, reqDom)
	if wDom.Code != http.StatusOK {
		t.Fatalf("GET /api/learn/domains status = %d, want 200", wDom.Code)
	}
	var domResp dto.LearnDomainsResponse
	if err := json.Unmarshal(wDom.Body.Bytes(), &domResp); err != nil {
		t.Fatalf("unmarshal domResp: %v", err)
	}
	if domResp.TotalSubtopics != 28 || len(domResp.Domains[0].Subtopics) != 4 {
		t.Fatalf("expected 28 total_subtopics and 4 in domain1, got %d / %d", domResp.TotalSubtopics, len(domResp.Domains[0].Subtopics))
	}

	// 2. First GET /api/learn/subtopics/1.3 locks 1 primary + 1 domain fallback and hides answers
	wSub1 := httptest.NewRecorder()
	reqSub1, _ := http.NewRequest("GET", "/api/learn/subtopics/1.3", nil)
	r.ServeHTTP(wSub1, reqSub1)
	if wSub1.Code != http.StatusOK {
		t.Fatalf("GET /api/learn/subtopics/1.3 status = %d, want 200: %s", wSub1.Code, wSub1.Body.String())
	}
	var subResp dto.LearnSubtopicDetailResponse
	if err := json.Unmarshal(wSub1.Body.Bytes(), &subResp); err != nil {
		t.Fatalf("unmarshal subResp: %v", err)
	}
	if len(subResp.SampleQuestions) != 2 {
		t.Fatalf("expected 2 locked sample questions, got %d", len(subResp.SampleQuestions))
	}
	if subResp.SampleQuestions[0].QuestionID != "q-13-a" || subResp.SampleQuestions[0].IsFallback {
		t.Fatalf("first draw should be primary non-fallback q-13-a, got %+v", subResp.SampleQuestions[0])
	}
	if !subResp.SampleQuestions[1].IsFallback {
		t.Fatalf("second draw should be marked is_fallback=true, got %+v", subResp.SampleQuestions[1])
	}
	for _, sq := range subResp.SampleQuestions {
		if sq.CorrectAnswer != nil || sq.Explanation != nil {
			t.Fatalf("answers/explanations must be omitted before submission: %+v", sq)
		}
	}

	// 3. Submitting 1st attempt evaluates immediately, locks answer, and reveals explanation
	bodyBytes, _ := json.Marshal(dto.AttemptSampleQuestionRequest{Answer: "b"})
	wAtt1 := httptest.NewRecorder()
	reqAtt1, _ := http.NewRequest("POST", "/api/learn/subtopics/1.3/sample-questions/q-13-a/attempt", bytes.NewReader(bodyBytes))
	reqAtt1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wAtt1, reqAtt1)
	if wAtt1.Code != http.StatusOK {
		t.Fatalf("attempt 1 status = %d, want 200: %s", wAtt1.Code, wAtt1.Body.String())
	}
	var attResp1 dto.AttemptSampleQuestionResponse
	if err := json.Unmarshal(wAtt1.Body.Bytes(), &attResp1); err != nil {
		t.Fatalf("unmarshal attResp1: %v", err)
	}
	if attResp1.IsCorrect || attResp1.SubmittedAnswer != "B" || attResp1.CorrectAnswer != "A" || attResp1.Explanation == "" || attResp1.ReadyToComplete {
		t.Fatalf("unexpected first-attempt grading response: %+v", attResp1)
	}

	// 4. Re-submitting with "A" must NOT overwrite locked wrong answer "B"
	retryBytes, _ := json.Marshal(dto.AttemptSampleQuestionRequest{Answer: "A"})
	wAttRetry := httptest.NewRecorder()
	reqAttRetry, _ := http.NewRequest("POST", "/api/learn/subtopics/1.3/sample-questions/q-13-a/attempt", bytes.NewReader(retryBytes))
	reqAttRetry.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wAttRetry, reqAttRetry)
	var attRetry dto.AttemptSampleQuestionResponse
	_ = json.Unmarshal(wAttRetry.Body.Bytes(), &attRetry)
	if !attRetry.AlreadySubmitted || attRetry.SubmittedAnswer != "B" || attRetry.IsCorrect {
		t.Fatalf("first submitted answer must stay locked, got %+v", attRetry)
	}

	// 5. Specialist preview returns preview_mode=true without mutating learner sample rows
	rSpec := setupLearnRouter(lc, "spec-preview", "specialist")
	wSpec := httptest.NewRecorder()
	reqSpec, _ := http.NewRequest("GET", "/api/learn/subtopics/1.1", nil)
	rSpec.ServeHTTP(wSpec, reqSpec)
	if wSpec.Code != http.StatusOK {
		t.Fatalf("specialist GET subtopic status = %d, want 200", wSpec.Code)
	}
	var specResp dto.LearnSubtopicDetailResponse
	_ = json.Unmarshal(wSpec.Body.Bytes(), &specResp)
	if !specResp.PreviewMode || len(specResp.SampleQuestions) != 2 {
		t.Fatalf("expected specialist preview_mode=true with 2 questions, got %+v", specResp)
	}
}

func TestBuildPostTestExclusionQuery_SQLIncludesPreTestAndLearnSamples(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "host=localhost user=test dbname=test port=5432 sslmode=disable",
		PreferSimpleProtocol: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("gorm.Open DryRun failed: %v", err)
	}

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var qList []models.Question
		return tx.Where("domain_id = ? AND is_active = true AND question_id NOT IN (?)",
			"domain1",
			BuildPostTestExclusionQuery(tx, "user-abc"),
		).Order("RANDOM()").Limit(4).Find(&qList)
	})

	if !strings.Contains(sql, "exam_session_questions") || !strings.Contains(sql, "exam_type = 'pre'") {
		t.Fatalf("expected Pre-Test exclusion in generated SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "user_theory_sample_questions") {
		t.Fatalf("expected Learn sample question exclusion in generated SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "UNION") {
		t.Fatalf("expected UNION in exclusion subquery, got: %s", sql)
	}
}

func TestLearnController_DirectDomainEssentialsAndOptionalDeepDiveCoexist(t *testing.T) {
	store := newFakeLearnStore()
	userID := "essentials-learner"

	store.profiles[userID] = []models.UserSkillProfile{
		{UserID: userID, DomainID: "domain1", ProficiencyScore: 19.0},
		{UserID: userID, DomainID: "domain2", ProficiencyScore: 45.0},
		{UserID: userID, DomainID: "domain3", ProficiencyScore: 60.0},
		{UserID: userID, DomainID: "domain4", ProficiencyScore: 55.0},
		{UserID: userID, DomainID: "domain5", ProficiencyScore: 70.0},
	}

	lc := NewLearnControllerWithStore(store)
	r := setupLearnRouter(lc, userID, "learner")

	// Completing 4 of 5 Domain Essentials directly leaves theory_completed=false
	for i := 0; i < 4; i++ {
		domID := services.OrderedTheoryDomains[i].DomainID
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/learn/domains/"+domID+"/complete", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("POST /api/learn/domains/%s/complete = %d, want 200: %s", domID, w.Code, w.Body.String())
		}
		var resp dto.CompleteTheoryDomainResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.CompletedCount != i+1 || resp.TheoryCompleted {
			t.Fatalf("step %d: expected CompletedCount=%d TheoryCompleted=false, got %+v", i+1, i+1, resp)
		}
	}

	// Completing the 5th Domain Essential sets theory_completed=true with 0/28 subtopics
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest("POST", "/api/learn/domains/domain5/complete", nil)
	r.ServeHTTP(w5, req5)
	if w5.Code != http.StatusOK {
		t.Fatalf("POST /api/learn/domains/domain5/complete = %d, want 200", w5.Code)
	}
	var resp5 dto.CompleteTheoryDomainResponse
	_ = json.Unmarshal(w5.Body.Bytes(), &resp5)
	if !resp5.TheoryCompleted || resp5.CompletedCount != 5 || len(resp5.CompletedDomains) != 5 {
		t.Fatalf("expected 5/5 theory_completed=true, got %+v", resp5)
	}

	// GET /api/learn/domains confirms 5/5 completed domains and 0/28 optional subtopics
	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest("GET", "/api/learn/domains", nil)
	r.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Fatalf("GET /api/learn/domains status = %d, want 200", wGet.Code)
	}
	var domResp dto.LearnDomainsResponse
	_ = json.Unmarshal(wGet.Body.Bytes(), &domResp)
	if !domResp.TheoryCompleted || domResp.CompletedCount != 5 || domResp.CompletedSubtopicCount != 0 {
		t.Fatalf("expected theory_completed=true with 5/5 domains & 0/28 optional subtopics, got %+v", domResp)
	}
	for _, d := range domResp.Domains {
		if !d.IsCompleted || d.CompletedAt == nil {
			t.Fatalf("domain %s should be marked completed directly: %+v", d.DomainID, d)
		}
	}
	if domResp.Domains[0].ProficiencyScore != 19.0 {
		t.Fatalf("proficiency_score drifted: got %v, want 19.0", domResp.Domains[0].ProficiencyScore)
	}
}

func TestLearnController_LocalizationPrecedenceErrorCodesAndEnglishOnlySamples(t *testing.T) {
	store := newFakeLearnStore()
	userID := "bilingual-learner"
	store.profiles[userID] = []models.UserSkillProfile{
		{UserID: userID, DomainID: "domain1", ProficiencyScore: 60.0},
	}
	store.seedSamplePool("domain1", "1.1", []models.Question{
		{
			QuestionID:    "q-en-11-a",
			DomainID:      "domain1",
			Type:          "multiple_choice",
			QuestionText:  "Which control category includes firewalls and encryption?",
			Options:       map[string]any{"A": "Technical", "B": "Managerial", "C": "Physical", "D": "Operational"},
			CorrectAnswer: "A",
			Explanation:   "Firewalls and encryption are technical controls.",
			IsActive:      true,
		},
		{
			QuestionID:    "q-en-11-b",
			DomainID:      "domain1",
			Type:          "multiple_choice",
			QuestionText:  "Which control type stops unauthorized access before impact?",
			Options:       map[string]any{"A": "Detective", "B": "Preventive", "C": "Deterrent", "D": "Corrective"},
			CorrectAnswer: "B",
			Explanation:   "Preventive controls stop unauthorized actions before they occur.",
			IsActive:      true,
		},
	})
	store.preferredLangs = map[string]string{userID: "th"}

	lc := NewLearnControllerWithStore(store)
	r := setupLearnRouter(lc, userID, "learner")

	// 1. DB preferred_language="th" wins when ?lang and Accept-Language omitted
	wDB := httptest.NewRecorder()
	reqDB, _ := http.NewRequest("GET", "/api/learn/domains", nil)
	r.ServeHTTP(wDB, reqDB)
	if wDB.Code != http.StatusOK {
		t.Fatalf("GET /api/learn/domains status = %d", wDB.Code)
	}
	var dbResp dto.LearnDomainsResponse
	_ = json.Unmarshal(wDB.Body.Bytes(), &dbResp)
	if !containsThaiRune(dbResp.Domains[0].Title) || !containsThaiRune(dbResp.Domains[0].Subtopics[0].Title) {
		t.Fatalf("expected Thai domain1 and subtopic 1.1 from user preferred_language=th, got dom=%q sub=%q",
			dbResp.Domains[0].Title, dbResp.Domains[0].Subtopics[0].Title)
	}

	// 2. ?lang=en overrides DB preferred_language=th and Accept-Language=th
	wEnOverride := httptest.NewRecorder()
	reqEn, _ := http.NewRequest("GET", "/api/learn/domains?lang=en", nil)
	reqEn.Header.Set("Accept-Language", "th")
	r.ServeHTTP(wEnOverride, reqEn)
	var enResp dto.LearnDomainsResponse
	_ = json.Unmarshal(wEnOverride.Body.Bytes(), &enResp)
	if containsThaiRune(enResp.Domains[0].Title) || containsThaiRune(enResp.Domains[0].Content) {
		t.Fatalf("?lang=en must override Accept-Language and DB to return English, got title=%q", enResp.Domains[0].Title)
	}

	// 3. ?lang=th returns Thai markdown while keeping sample_questions 100% English (zero Thai runes)
	wSub := httptest.NewRecorder()
	reqSub, _ := http.NewRequest("GET", "/api/learn/subtopics/1.1?lang=th", nil)
	r.ServeHTTP(wSub, reqSub)
	if wSub.Code != http.StatusOK {
		t.Fatalf("GET /api/learn/subtopics/1.1?lang=th status = %d", wSub.Code)
	}
	var subResp dto.LearnSubtopicDetailResponse
	_ = json.Unmarshal(wSub.Body.Bytes(), &subResp)
	if !containsThaiRune(subResp.Title) || !containsThaiRune(subResp.Content) {
		t.Fatalf("expected Thai subtopic 1.1 title and content, got title=%q", subResp.Title)
	}
	if len(subResp.SampleQuestions) != 2 {
		t.Fatalf("expected 2 sample questions, got %d", len(subResp.SampleQuestions))
	}
	for _, sq := range subResp.SampleQuestions {
		sqJSON, _ := json.Marshal(sq)
		if containsThaiRune(string(sqJSON)) {
			t.Fatalf("sample_questions must remain strictly English (zero Thai runes): %s", string(sqJSON))
		}
	}

	// Attempting sample question returns English explanation under Accept-Language: th
	attemptBody, _ := json.Marshal(dto.AttemptSampleQuestionRequest{Answer: "A"})
	wAtt := httptest.NewRecorder()
	reqAtt, _ := http.NewRequest("POST", "/api/learn/subtopics/1.1/sample-questions/"+subResp.SampleQuestions[0].QuestionID+"/attempt", bytes.NewReader(attemptBody))
	reqAtt.Header.Set("Content-Type", "application/json")
	reqAtt.Header.Set("Accept-Language", "th")
	r.ServeHTTP(wAtt, reqAtt)
	if wAtt.Code != http.StatusOK || containsThaiRune(wAtt.Body.String()) {
		t.Fatalf("sample attempt response must stay strictly English: code=%d body=%s", wAtt.Code, wAtt.Body.String())
	}

	// 4. Canonical Learn error codes
	wIncomplete := httptest.NewRecorder()
	reqInc, _ := http.NewRequest("POST", "/api/learn/subtopics/1.1/complete", nil)
	r.ServeHTTP(wIncomplete, reqInc)
	var incBody map[string]any
	_ = json.Unmarshal(wIncomplete.Body.Bytes(), &incBody)
	if wIncomplete.Code != http.StatusBadRequest || incBody["code"] != "LEARN_SAMPLE_INCOMPLETE" {
		t.Fatalf("expected 400 LEARN_SAMPLE_INCOMPLETE, got %d %+v", wIncomplete.Code, incBody)
	}

	wBadSub := httptest.NewRecorder()
	reqBadSub, _ := http.NewRequest("GET", "/api/learn/subtopics/9.9", nil)
	r.ServeHTTP(wBadSub, reqBadSub)
	var badSubBody map[string]any
	_ = json.Unmarshal(wBadSub.Body.Bytes(), &badSubBody)
	if wBadSub.Code != http.StatusNotFound || badSubBody["code"] != "LEARN_SUBTOPIC_NOT_FOUND" {
		t.Fatalf("expected 404 LEARN_SUBTOPIC_NOT_FOUND, got %d %+v", wBadSub.Code, badSubBody)
	}

	wBadDom := httptest.NewRecorder()
	reqBadDom, _ := http.NewRequest("POST", "/api/learn/domains/domain9/complete", nil)
	r.ServeHTTP(wBadDom, reqBadDom)
	var badDomBody map[string]any
	_ = json.Unmarshal(wBadDom.Body.Bytes(), &badDomBody)
	if wBadDom.Code != http.StatusBadRequest || badDomBody["code"] != "LEARN_DOMAIN_NOT_FOUND" {
		t.Fatalf("expected 400 LEARN_DOMAIN_NOT_FOUND, got %d %+v", wBadDom.Code, badDomBody)
	}

	wPreGate := httptest.NewRecorder()
	rLocked := setupLearnRouter(lc, "no-pretest-user", "learner")
	reqPre, _ := http.NewRequest("GET", "/api/learn/domains", nil)
	rLocked.ServeHTTP(wPreGate, reqPre)
	var preBody map[string]any
	_ = json.Unmarshal(wPreGate.Body.Bytes(), &preBody)
	if wPreGate.Code != http.StatusForbidden || preBody["code"] != "LEARN_PRETEST_REQUIRED" {
		t.Fatalf("expected 403 LEARN_PRETEST_REQUIRED, got %d %+v", wPreGate.Code, preBody)
	}
}

func TestLearnController_SampleQuestionTranslationsPreAndPostAttempt(t *testing.T) {
	store := newFakeLearnStore()
	userID := "learn-bi-user"
	store.profiles[userID] = []models.UserSkillProfile{{UserID: userID, DomainID: "domain1", ProficiencyScore: 60.0}}
	store.seedSamplePool("domain1", "1.1", []models.Question{
		{
			QuestionID:    "q-th-11-a",
			DomainID:      "domain1",
			Type:          "multiple_choice",
			QuestionText:  "Canonical EN question A?",
			Options:       map[string]any{"A": "One", "B": "Two"},
			CorrectAnswer: "A",
			Explanation:   "Correct Answer: A. Right. Option B is incorrect. Wrong.",
			Translations: map[string]any{
				"th": map[string]any{
					"question_text": "คำถามไทย A?",
					"options":       map[string]any{"A": "หนึ่ง", "B": "สอง"},
					"explanation":   "คำตอบที่ถูกต้อง: A. ถูก ตัวเลือก B ไม่ถูกต้อง ผิด",
				},
			},
			IsActive: true,
		},
		{
			QuestionID:    "q-th-11-b",
			DomainID:      "domain1",
			Type:          "multiple_choice",
			QuestionText:  "Canonical EN question B?",
			Options:       map[string]any{"A": "Yes", "B": "No"},
			CorrectAnswer: "B",
			Explanation:   "Correct Answer: B. Right. Option A is incorrect. Wrong.",
			Translations: map[string]any{
				"th": map[string]any{
					"question_text": "คำถามไทย B?",
					"options":       map[string]any{"A": "ใช่", "B": "ไม่ใช่"},
					"explanation":   "คำตอบที่ถูกต้อง: B. ถูก ตัวเลือก A ไม่ถูกต้อง ผิด",
				},
			},
			IsActive: true,
		},
	})

	lc := NewLearnControllerWithStore(store)
	r := setupLearnRouter(lc, userID, "learner")

	wPre := httptest.NewRecorder()
	reqPre, _ := http.NewRequest("GET", "/api/learn/subtopics/1.1?lang=th", nil)
	r.ServeHTTP(wPre, reqPre)
	if wPre.Code != http.StatusOK {
		t.Fatalf("GET subtopic failed: %d", wPre.Code)
	}
	var preResp dto.LearnSubtopicDetailResponse
	_ = json.Unmarshal(wPre.Body.Bytes(), &preResp)
	firstTh := preResp.SampleQuestions[0].Translations.(map[string]any)["th"].(map[string]any)
	if firstTh["question_text"] != "คำถามไทย A?" || firstTh["options"] == nil {
		t.Fatalf("expected th question_text and options pre-attempt, got %+v", firstTh)
	}
	if _, leaked := firstTh["explanation"]; leaked {
		t.Fatalf("unattempted sample leaked translations.th.explanation: %+v", firstTh)
	}

	attemptBody, _ := json.Marshal(dto.AttemptSampleQuestionRequest{Answer: "A"})
	wAtt := httptest.NewRecorder()
	reqAtt, _ := http.NewRequest("POST", "/api/learn/subtopics/1.1/sample-questions/q-th-11-a/attempt", bytes.NewReader(attemptBody))
	reqAtt.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wAtt, reqAtt)
	if wAtt.Code != http.StatusOK {
		t.Fatalf("attempt failed: %d %s", wAtt.Code, wAtt.Body.String())
	}
	var attResp dto.AttemptSampleQuestionResponse
	_ = json.Unmarshal(wAtt.Body.Bytes(), &attResp)
	attTh := attResp.Translations.(map[string]any)["th"].(map[string]any)
	if attExp, ok := attTh["explanation"].(string); !ok || attExp == "" {
		t.Fatalf("expected th.explanation in attempt response, got %+v", attTh)
	}

	wPost := httptest.NewRecorder()
	reqPost, _ := http.NewRequest("GET", "/api/learn/subtopics/1.1?lang=th", nil)
	r.ServeHTTP(wPost, reqPost)
	var postResp dto.LearnSubtopicDetailResponse
	_ = json.Unmarshal(wPost.Body.Bytes(), &postResp)
	post0 := postResp.SampleQuestions[0].Translations.(map[string]any)["th"].(map[string]any)
	post1 := postResp.SampleQuestions[1].Translations.(map[string]any)["th"].(map[string]any)
	post0Exp, hasPost0Exp := post0["explanation"].(string)
	_, hasPost1Exp := post1["explanation"]
	if !hasPost0Exp || post0Exp == "" || hasPost1Exp {
		t.Fatalf("expected answered sample 0 to expose th.explanation and unanswered sample 1 to omit it: 0=%+v 1=%+v", post0, post1)
	}
}

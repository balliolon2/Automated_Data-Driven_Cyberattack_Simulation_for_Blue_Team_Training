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
	if _, exists := f.subProgress[userID][subTopicID]; !exists {
		f.subProgress[userID][subTopicID] = models.UserTheorySubtopicProgress{
			UserID:      userID,
			SubTopicID:  subTopicID,
			CompletedAt: time.Now(),
		}
	}

	completedSubInDomain := 0
	for _, sp := range specList {
		if _, ok := f.subProgress[userID][sp.SubTopicID]; ok {
			completedSubInDomain++
		}
	}
	domainAutoCompleted := false
	if len(specList) > 0 && completedSubInDomain == len(specList) {
		if _, exists := f.progress[userID][domainID]; !exists {
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
	completedDomains, _ := buildOrderedCompletedDomains(progRows)

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
		TheoryCompleted:            services.IsTheorySubtopicGateSatisfied(completedSub),
		PreviewMode:                false,
	}, nil
}

func (f *fakeLearnStore) LoadModules() ([]services.TheoryModuleContent, error) {
	return services.LoadTheoryModules(nil)
}

func (f *fakeLearnStore) LoadSubtopics() ([]services.TheorySubtopicContent, error) {
	return services.LoadTheorySubtopics(nil)
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

		// Anti-bypass: direct POST /domains/domain1/complete blocked while child subtopics incomplete
		wBypass := httptest.NewRecorder()
		reqBypass, _ := http.NewRequest("POST", "/api/learn/domains/domain1/complete", nil)
		r.ServeHTTP(wBypass, reqBypass)
		if wBypass.Code != http.StatusForbidden {
			t.Fatalf("expected 403 subtopics_incomplete on premature domain complete, got %d: %s", wBypass.Code, wBypass.Body.String())
		}
		var bypassErr dto.TheoryGateErrorResponse
		_ = json.Unmarshal(wBypass.Body.Bytes(), &bypassErr)
		if bypassErr.Reason != "subtopics_incomplete" || len(bypassErr.MissingSubtopics) != 4 {
			t.Fatalf("unexpected anti-bypass error payload: %+v", bypassErr)
		}

		// Complete sub-topic 1.1 requires attempting its 2 sample questions first (400 otherwise)
		wSubPremature := httptest.NewRecorder()
		reqSubPremature, _ := http.NewRequest("POST", "/api/learn/subtopics/1.1/complete", nil)
		r.ServeHTTP(wSubPremature, reqSubPremature)
		if wSubPremature.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 sample_questions_incomplete before attempts, got %d: %s", wSubPremature.Code, wSubPremature.Body.String())
		}

		// Study and complete all 28 sub-topics, verifying domain auto-completion + idempotency
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
				if !c14.DomainAutoCompleted || c14.CompletedCount != 1 || c14.TheoryCompleted {
					t.Fatalf("expected 1.4 completion to auto-complete domain1: %+v", c14)
				}
			}
		}

		// Idempotent POST /domains/domain1/complete is now allowed cleanly (200)
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

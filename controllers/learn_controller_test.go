package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cybersim/dto"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
)

type fakeLearnStore struct {
	profiles map[string][]models.UserSkillProfile
	progress map[string]map[string]models.UserTheoryProgress
}

func newFakeLearnStore() *fakeLearnStore {
	return &fakeLearnStore{
		profiles: make(map[string][]models.UserSkillProfile),
		progress: make(map[string]map[string]models.UserTheoryProgress),
	}
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

func (f *fakeLearnStore) LoadModules() ([]services.TheoryModuleContent, error) {
	return services.LoadTheoryModules(nil)
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

		// Complete domain1 twice (idempotent)
		for attempt := 0; attempt < 2; attempt++ {
			wComp := httptest.NewRecorder()
			reqComp, _ := http.NewRequest("POST", "/api/learn/domains/domain1/complete", nil)
			r.ServeHTTP(wComp, reqComp)
			if wComp.Code != http.StatusOK {
				t.Fatalf("attempt %d status = %d, want 200", attempt, wComp.Code)
			}
			var compResp dto.CompleteTheoryDomainResponse
			if err := json.Unmarshal(wComp.Body.Bytes(), &compResp); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}
			if compResp.CompletedCount != 1 || compResp.TheoryCompleted {
				t.Fatalf("unexpected completion state after domain1: %+v", compResp)
			}
		}

		// Complete remaining 4 domains
		for _, d := range []string{"domain2", "domain3", "domain4", "domain5"} {
			wComp := httptest.NewRecorder()
			reqComp, _ := http.NewRequest("POST", "/api/learn/domains/"+d+"/complete", nil)
			r.ServeHTTP(wComp, reqComp)
			if wComp.Code != http.StatusOK {
				t.Fatalf("complete %s status = %d, want 200", d, wComp.Code)
			}
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
		if !resp.TheoryCompleted || resp.CompletedCount != 5 || len(resp.CompletedDomains) != 5 {
			t.Fatalf("expected 5/5 theory_completed=true, got %+v", resp)
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

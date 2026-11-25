package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"cybersim/dto"
	"cybersim/middlewares"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type fakeNFRStore struct {
	mu       sync.Mutex
	facts    services.NFRRawDatabaseFacts
	pingErr  error
	runs     map[string]*models.NFREvaluationRun
	order    []string
	nextSeq  int
	forceErr error
}

func newFakeNFRStore() *fakeNFRStore {
	csuq := 6.20
	tlx := 31.5
	examMed := 840.0
	simMed := 390.0
	return &fakeNFRStore{
		facts: services.NFRRawDatabaseFacts{
			DBStatus:               "up",
			DBLatencyMs:            3.2,
			ActiveQuestionsCount:   540,
			TheorySubtopicsCount:   28,
			ThaiTheoryCount:        33,
			ActiveScenariosCount:   15,
			TotalSimSessions:       6,
			SessionsWithSnapshot:   6,
			SnapshotsCompleteMeta:  6,
			AIGeneratedSessions:    4,
			StaticFallbackSessions: 2,
			SystemFailureSessions:  0,
			ValidSnapshotCount:     6,
			TotalSnapshotCount:     6,
			FallbackReasons:        map[string]int{"LLM_API_KEY not configured": 2},
			CSUQOverallMean:        &csuq,
			NASATLXOverallMean:     &tlx,
			QuestionnaireCount:     4,
			ExamTotalCount:         12,
			ExamCompletedCount:     11,
			ExamMedianSeconds:      &examMed,
			SimFinishedCount:       6,
			SimCompletedCount:      6,
			SimMedianSeconds:       &simMed,
		},
		runs: make(map[string]*models.NFREvaluationRun),
	}
}

func (f *fakeNFRStore) PingDB(ctx context.Context) (float64, error) {
	if f.pingErr != nil {
		return 15.0, f.pingErr
	}
	return 2.1, nil
}

func (f *fakeNFRStore) GatherLiveNFRFacts(ctx context.Context) (services.NFRRawDatabaseFacts, error) {
	if f.forceErr != nil {
		return services.NFRRawDatabaseFacts{}, f.forceErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := f.facts
	if f.pingErr != nil {
		cp.DBStatus = "down"
	}
	return cp, nil
}

func (f *fakeNFRStore) CreateNFRRun(run *models.NFREvaluationRun) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.forceErr != nil {
		return f.forceErr
	}
	f.nextSeq++
	if run.RunID == "" {
		run.RunID = "11111111-1111-1111-1111-11111111111" + string(rune('0'+f.nextSeq))
	}
	if run.CreatedAt.IsZero() {
		run.CreatedAt = time.Now().UTC()
	}
	cp := *run
	f.runs[run.RunID] = &cp
	f.order = append([]string{run.RunID}, f.order...)
	return nil
}

func (f *fakeNFRStore) ListNFRRuns(limit int) ([]models.NFREvaluationRun, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]models.NFREvaluationRun, 0, len(f.order))
	for _, id := range f.order {
		if r, ok := f.runs[id]; ok {
			out = append(out, *r)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (f *fakeNFRStore) GetNFRRunByID(runID string) (*models.NFREvaluationRun, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.runs[runID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *r
	return &cp, nil
}

func (f *fakeNFRStore) DeleteNFRRun(runID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.runs[runID]; !ok {
		return false, nil
	}
	delete(f.runs, runID)
	return true, nil
}

func signNFRTestToken(t *testing.T, userID, role, secret string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID,
		"role":     role,
		"nickname": "tester_" + role,
		"exp":      time.Now().Add(time.Hour).Unix(),
	})
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token failed: %v", err)
	}
	return signed
}

func setupNFRTestRouter(t *testing.T, store *fakeNFRStore, telemetry *services.NFRTelemetryStore) (*gin.Engine, *NFRController) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "super-strong-test-jwt-secret-key-at-least-32-chars")
	t.Setenv("LLM_API_KEY", "sk-test-secret-must-never-appear-in-response")

	r := gin.New()
	r.Use(middlewares.TelemetryMiddleware(telemetry))

	nc := NewNFRControllerWithStore(store, telemetry, r)
	api := r.Group("/api")
	api.GET("/health", nc.HealthCheck)
	api.GET("/exams/sample", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	api.GET("/learn/domains", middlewares.AuthMiddleware(), func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"domains": []any{}}) })
	api.GET("/exams/session", middlewares.AuthMiddleware(), func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"session": nil}) })

	admin := api.Group("/admin", middlewares.AuthMiddleware(), middlewares.RequireAdmin())
	admin.GET("/nfr/live", nc.GetLiveNFR)
	admin.POST("/nfr/telemetry/reset", nc.ResetTelemetryWindow)
	admin.POST("/nfr/runs", nc.CreateNFRRun)
	admin.GET("/nfr/runs", nc.ListNFRRuns)
	admin.GET("/nfr/runs/:runId", nc.GetNFRRun)
	admin.DELETE("/nfr/runs/:runId", nc.DeleteNFRRun)
	admin.POST("/nfr/probe/scalability", nc.RunScalabilityProbe)
	admin.GET("/nfr/export", nc.ExportNFRReport)
	return r, nc
}

func TestNFRController_HealthCheckLiveAndDegraded(t *testing.T) {
	store := newFakeNFRStore()
	telemetry := services.NewNFRTelemetryStore(100)
	r, _ := setupNFRTestRouter(t, store, telemetry)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 healthy, got %d: %s", rec.Code, rec.Body.String())
	}
	var h dto.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &h); err != nil {
		t.Fatalf("decode health failed: %v", err)
	}
	if h.Status != "healthy" || h.DBStatus != "up" || h.DBLatencyMs <= 0 {
		t.Fatalf("unexpected healthy payload: %+v", h)
	}

	store.pingErr = errors.New("connection refused")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if rec2.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 degraded, got %d: %s", rec2.Code, rec2.Body.String())
	}
	var h2 dto.HealthResponse
	_ = json.Unmarshal(rec2.Body.Bytes(), &h2)
	if h2.Status != "degraded" || h2.DBStatus != "down" {
		t.Fatalf("unexpected degraded payload: %+v", h2)
	}
}

func TestNFRController_RBACLiveAuditIsolationAndReset(t *testing.T) {
	store := newFakeNFRStore()
	telemetry := services.NewNFRTelemetryStore(100)
	r, _ := setupNFRTestRouter(t, store, telemetry)
	secret := "super-strong-test-jwt-secret-key-at-least-32-chars"

	learnerTok := signNFRTestToken(t, "u-learner", "learner", secret)
	specialistTok := signNFRTestToken(t, "u-spec", "specialist", secret)
	adminTok := signNFRTestToken(t, "u-admin", "admin", secret)

	// Unauthenticated and non-admin callers must be blocked on every Admin NFR route
	for _, ep := range []struct{ method, path string }{
		{http.MethodGet, "/api/admin/nfr/live"},
		{http.MethodPost, "/api/admin/nfr/telemetry/reset"},
		{http.MethodGet, "/api/admin/nfr/runs"},
		{http.MethodPost, "/api/admin/nfr/runs"},
		{http.MethodPost, "/api/admin/nfr/probe/scalability"},
		{http.MethodGet, "/api/admin/nfr/export?format=csv"},
	} {
		recUnauth := httptest.NewRecorder()
		r.ServeHTTP(recUnauth, httptest.NewRequest(ep.method, ep.path, nil))
		if recUnauth.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s unauth: expected 401, got %d", ep.method, ep.path, recUnauth.Code)
		}

		for _, tok := range []string{learnerTok, specialistTok} {
			req := httptest.NewRequest(ep.method, ep.path, nil)
			req.Header.Set("Authorization", "Bearer "+tok)
			recForb := httptest.NewRecorder()
			r.ServeHTTP(recForb, req)
			if recForb.Code != http.StatusForbidden {
				t.Fatalf("%s %s non-admin: expected 403, got %d", ep.method, ep.path, recForb.Code)
			}
		}
	}

	// Reset before measuring live audit isolation, then send 1 genuine user request
	telemetry.ResetWindow()
	recSample := httptest.NewRecorder()
	r.ServeHTTP(recSample, httptest.NewRequest(http.MethodGet, "/api/exams/sample", nil))
	if recSample.Code != http.StatusOK {
		t.Fatalf("sample call failed: %d", recSample.Code)
	}

	reqLive := httptest.NewRequest(http.MethodGet, "/api/admin/nfr/live", nil)
	reqLive.Header.Set("Authorization", "Bearer "+adminTok)
	recLive := httptest.NewRecorder()
	r.ServeHTTP(recLive, reqLive)
	if recLive.Code != http.StatusOK {
		t.Fatalf("GET /live failed: %d %s", recLive.Code, recLive.Body.String())
	}

	bodyStr := recLive.Body.String()
	if strings.Contains(bodyStr, secret) || strings.Contains(bodyStr, "sk-test-secret-must-never-appear-in-response") {
		t.Fatalf("Zero-Secret Exposure violated! Raw secret leaked in response")
	}

	var rep dto.NFRReportResponse
	if err := json.Unmarshal(recLive.Body.Bytes(), &rep); err != nil {
		t.Fatalf("decode live report failed: %v", err)
	}
	if len(rep.Dimensions) != 10 {
		t.Fatalf("expected 10 dimensions, got %d", len(rep.Dimensions))
	}
	// Internal 6 security probe requests must NOT be counted in passive telemetry!
	if rep.Runtime.TotalRequests != 1 {
		t.Fatalf("expected exactly 1 passive request before /live finished (6 security probes isolated), got %d", rep.Runtime.TotalRequests)
	}

	// Reset telemetry window cleanly
	reqReset := httptest.NewRequest(http.MethodPost, "/api/admin/nfr/telemetry/reset", nil)
	reqReset.Header.Set("Authorization", "Bearer "+adminTok)
	recReset := httptest.NewRecorder()
	r.ServeHTTP(recReset, reqReset)
	if recReset.Code != http.StatusOK {
		t.Fatalf("POST /telemetry/reset failed: %d %s", recReset.Code, recReset.Body.String())
	}
	if telemetry.Snapshot("release", "up", 1.0).TotalRequests != 0 {
		t.Fatalf("expected 0 requests in telemetry after reset")
	}
}

func TestNFRController_ProbeRunsLifecycleAndExport(t *testing.T) {
	store := newFakeNFRStore()
	telemetry := services.NewNFRTelemetryStore(100)
	r, nc := setupNFRTestRouter(t, store, telemetry)
	secret := "super-strong-test-jwt-secret-key-at-least-32-chars"
	adminTok := signNFRTestToken(t, "u-admin", "admin", secret)

	// 1. Invalid probe preset & out-of-cap custom concurrency -> 400
	badProbeReq := httptest.NewRequest(http.MethodPost, "/api/admin/nfr/probe/scalability", bytes.NewBufferString(`{"preset":"extreme_1000"}`))
	badProbeReq.Header.Set("Authorization", "Bearer "+adminTok)
	badProbeReq.Header.Set("Content-Type", "application/json")
	recBadProbe := httptest.NewRecorder()
	r.ServeHTTP(recBadProbe, badProbeReq)
	if recBadProbe.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 NFR_INVALID_PROBE_PARAMS, got %d: %s", recBadProbe.Code, recBadProbe.Body.String())
	}

	overCapReq := httptest.NewRequest(http.MethodPost, "/api/admin/nfr/probe/scalability", bytes.NewBufferString(`{"concurrency":25,"total_requests":50}`))
	overCapReq.Header.Set("Authorization", "Bearer "+adminTok)
	overCapReq.Header.Set("Content-Type", "application/json")
	recOverCap := httptest.NewRecorder()
	r.ServeHTTP(recOverCap, overCapReq)
	if recOverCap.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 NFR_INVALID_PROBE_PARAMS for concurrency > 20, got %d: %s", recOverCap.Code, recOverCap.Body.String())
	}

	// Non-UUID run IDs return 404 rather than 500 across get, delete, and export
	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/admin/nfr/runs/not-a-uuid"},
		{http.MethodDelete, "/api/admin/nfr/runs/not-a-uuid"},
		{http.MethodGet, "/api/admin/nfr/export?run_id=not-a-uuid"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req.Header.Set("Authorization", "Bearer "+adminTok)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 NFR_RUN_NOT_FOUND for %s %s, got %d: %s", tc.method, tc.path, rec.Code, rec.Body.String())
		}
	}

	// 2. Concurrent probe mutex -> 409
	nc.probeRunning.Store(true)
	busyReq := httptest.NewRequest(http.MethodPost, "/api/admin/nfr/probe/scalability", bytes.NewBufferString(`{"preset":"light"}`))
	busyReq.Header.Set("Authorization", "Bearer "+adminTok)
	busyReq.Header.Set("Content-Type", "application/json")
	recBusy := httptest.NewRecorder()
	r.ServeHTTP(recBusy, busyReq)
	nc.probeRunning.Store(false)
	if recBusy.Code != http.StatusConflict {
		t.Fatalf("expected 409 NFR_PROBE_ALREADY_RUNNING, got %d: %s", recBusy.Code, recBusy.Body.String())
	}

	// 3. Valid light probe with save_run=true -> 200, zero passive telemetry pollution from the 50 internal requests
	telemetry.ResetWindow()
	probeReq := httptest.NewRequest(http.MethodPost, "/api/admin/nfr/probe/scalability", bytes.NewBufferString(`{"preset":"light","save_run":true,"run_label":"=ProbeTest"}`))
	probeReq.Header.Set("Authorization", "Bearer "+adminTok)
	probeReq.Header.Set("Content-Type", "application/json")
	recProbe := httptest.NewRecorder()
	r.ServeHTTP(recProbe, probeReq)
	if recProbe.Code != http.StatusOK {
		t.Fatalf("expected 200 scalability probe, got %d: %s", recProbe.Code, recProbe.Body.String())
	}
	var probeResp dto.ScalabilityProbeResponse
	_ = json.Unmarshal(recProbe.Body.Bytes(), &probeResp)
	if probeResp.SavedRun == nil || len(probeResp.Benchmark.Levels) != 1 || probeResp.Benchmark.Levels[0].TotalRequests != 50 {
		t.Fatalf("unexpected probe response: %+v", probeResp)
	}
	if got := telemetry.Snapshot("release", "up", 1.0).TotalRequests; got != 1 {
		// Only the POST /probe/scalability HTTP request itself was counted; its 50 internal requests were isolated!
		t.Fatalf("expected exactly 1 HTTP sample (probe POST only; 50 internal requests isolated), got %d", got)
	}

	// 4. Oversized benchmark upload (> 1MB) -> 413
	oversized := `{"run_label":"big","notes":"` + strings.Repeat("a", (1<<20)+10) + `"}`
	reqBig := httptest.NewRequest(http.MethodPost, "/api/admin/nfr/runs", bytes.NewBufferString(oversized))
	reqBig.Header.Set("Authorization", "Bearer "+adminTok)
	reqBig.Header.Set("Content-Type", "application/json")
	recBig := httptest.NewRecorder()
	r.ServeHTTP(recBig, reqBig)
	if recBig.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 NFR_PAYLOAD_TOO_LARGE, got %d", recBig.Code)
	}

	// 5. CSV Export with UTF-8 BOM and formula neutralization ('=ProbeTest)
	runID := probeResp.SavedRun.RunID
	reqCSV := httptest.NewRequest(http.MethodGet, "/api/admin/nfr/export?format=csv&run_id="+runID, nil)
	reqCSV.Header.Set("Authorization", "Bearer "+adminTok)
	recCSV := httptest.NewRecorder()
	r.ServeHTTP(recCSV, reqCSV)
	if recCSV.Code != http.StatusOK {
		t.Fatalf("expected 200 CSV export, got %d: %s", recCSV.Code, recCSV.Body.String())
	}
	if !bytes.HasPrefix(recCSV.Body.Bytes(), []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatalf("expected UTF-8 BOM prefix on CSV")
	}
	if !strings.Contains(recCSV.Body.String(), "'=ProbeTest") {
		t.Fatalf("expected formula neutralization in CSV export")
	}

	// 6. Get, Delete & 404 verification
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/admin/nfr/runs/"+runID, nil)
	reqDel.Header.Set("Authorization", "Bearer "+adminTok)
	recDel := httptest.NewRecorder()
	r.ServeHTTP(recDel, reqDel)
	if recDel.Code != http.StatusOK {
		t.Fatalf("expected 200 delete run, got %d", recDel.Code)
	}

	reqMissing := httptest.NewRequest(http.MethodGet, "/api/admin/nfr/runs/"+runID, nil)
	reqMissing.Header.Set("Authorization", "Bearer "+adminTok)
	recMissing := httptest.NewRecorder()
	r.ServeHTTP(recMissing, reqMissing)
	if recMissing.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", recMissing.Code)
	}
}

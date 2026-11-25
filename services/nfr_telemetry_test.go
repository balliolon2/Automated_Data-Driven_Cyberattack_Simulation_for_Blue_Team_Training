package services

import (
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func almostEqualNFR(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func TestNFRTelemetryStore_ZeroSamplesAndR7Percentiles(t *testing.T) {
	store := NewNFRTelemetryStore(100)

	empty := store.Snapshot("release", "up", 2.5)
	if empty.TotalRequests != 0 {
		t.Fatalf("expected 0 total requests, got %d", empty.TotalRequests)
	}
	if empty.NonLLMLatency.P50 != nil || empty.NonLLMLatency.P95 != nil || empty.NonLLMLatency.P99 != nil {
		t.Fatalf("expected nil percentiles when 0 samples exist, got %+v", empty.NonLLMLatency)
	}
	if empty.AvailabilityPct != nil || empty.ErrorRate5xxPct != nil {
		t.Fatalf("expected nil availability/error rate when 0 requests exist")
	}

	// Known dataset: [10, 20, 30, 40, 50]
	// R-7 percentiles:
	// p50 -> idx = 0.50 * 4 = 2.0 => 30.0
	// p95 -> idx = 0.95 * 4 = 3.8 => 40 + 0.8*(50-40) = 48.0
	// p99 -> idx = 0.99 * 4 = 3.96 => 40 + 0.96*(50-40) = 49.6
	for _, d := range []float64{50, 10, 40, 20, 30} {
		store.RecordHTTPRequest("GET /api/learn/domains", http.StatusOK, d)
	}
	// Simulation start must be classified separately from non_llm_api
	store.RecordHTTPRequest("POST /api/simulation/start", http.StatusOK, 12000)

	snap := store.Snapshot("release", "up", 1.8)
	if snap.TotalRequests != 6 || snap.NonLLMRequests != 5 || snap.LLMRequests != 1 {
		t.Fatalf("unexpected request counts: total=%d nonLLM=%d llm=%d", snap.TotalRequests, snap.NonLLMRequests, snap.LLMRequests)
	}
	if snap.NonLLMLatency.P50 == nil || !almostEqualNFR(*snap.NonLLMLatency.P50, 30.0, 1e-4) {
		t.Fatalf("expected non-LLM p50=30.0, got %v", snap.NonLLMLatency.P50)
	}
	if snap.NonLLMLatency.P95 == nil || !almostEqualNFR(*snap.NonLLMLatency.P95, 48.0, 1e-4) {
		t.Fatalf("expected non-LLM p95=48.0, got %v", snap.NonLLMLatency.P95)
	}
	if snap.NonLLMLatency.P99 == nil || !almostEqualNFR(*snap.NonLLMLatency.P99, 49.6, 1e-4) {
		t.Fatalf("expected non-LLM p99=49.6, got %v", snap.NonLLMLatency.P99)
	}
	if snap.LLMLatency.Count != 1 || snap.LLMLatency.P95 == nil || !almostEqualNFR(*snap.LLMLatency.P95, 12000.0, 1e-4) {
		t.Fatalf("expected LLM p95=12000.0, got %+v", snap.LLMLatency)
	}
}

func TestNFRTelemetryStore_RingBufferWrapAroundAndResetWindow(t *testing.T) {
	store := NewNFRTelemetryStore(4)
	baseProc := store.ProcessStartedAt()

	// Write 6 requests into capacity-4 buffer: first 2 (1000ms) must be evicted from percentiles
	store.RecordHTTPRequest("GET /api/a", http.StatusInternalServerError, 1000)
	store.RecordHTTPRequest("GET /api/a", http.StatusBadRequest, 1000)
	for _, d := range []float64{10, 20, 30, 40} {
		store.RecordHTTPRequest("GET /api/a", http.StatusOK, d)
	}

	snap := store.Snapshot("release", "up", 2.0)
	if snap.TotalRequests != 6 {
		t.Fatalf("expected cumulative window request count 6, got %d", snap.TotalRequests)
	}
	if snap.Status5xx != 1 || snap.Status4xx != 1 || snap.Status2xx != 4 {
		t.Fatalf("unexpected status counts: 2xx=%d 4xx=%d 5xx=%d", snap.Status2xx, snap.Status4xx, snap.Status5xx)
	}
	if snap.NonLLMLatency.Count != 4 || snap.NonLLMLatency.Max == nil || !almostEqualNFR(*snap.NonLLMLatency.Max, 40.0, 1e-4) {
		t.Fatalf("expected bounded ring buffer max=40.0 after wrap, got %+v", snap.NonLLMLatency)
	}
	// Availability = (6 - 1) / 6 * 100 = 83.33%
	if snap.AvailabilityPct == nil || !almostEqualNFR(*snap.AvailabilityPct, 83.33, 0.02) {
		t.Fatalf("expected availability 83.33%%, got %v", snap.AvailabilityPct)
	}

	time.Sleep(5 * time.Millisecond)
	resetAt := store.ResetWindow()
	if !resetAt.After(baseProc) {
		t.Fatalf("expected window reset timestamp after process start")
	}
	after := store.Snapshot("release", "up", 1.0)
	if !after.ProcessStartedAt.Equal(baseProc) {
		t.Fatalf("processStartedAt must remain immutable across ResetWindow")
	}
	if after.TotalRequests != 0 || after.Status5xx != 0 || after.NonLLMLatency.Count != 0 {
		t.Fatalf("expected zeroed telemetry after ResetWindow, got %+v", after)
	}
}

func TestNFRTelemetryStore_ProbeTokenIsolationAndLLMRecording(t *testing.T) {
	store := NewNFRTelemetryStore(10)
	tok := store.InternalProbeToken()
	if len(tok) != 64 {
		t.Fatalf("expected 64-char hex internal probe token, got len=%d", len(tok))
	}

	validReq := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	validReq.Header.Set(InternalNFRProbeHeader, tok)
	if !store.IsInternalProbeRequest(validReq) {
		t.Fatalf("expected matching internal probe header to be recognized")
	}

	spoofReq := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	spoofReq.Header.Set(InternalNFRProbeHeader, "spoofed-token")
	if store.IsInternalProbeRequest(spoofReq) {
		t.Fatalf("spoofed probe header must not be trusted")
	}

	store.RecordLLMCall(LLMCallRecord{
		DurationMs:     1850.5,
		Attempts:       2,
		Retries:        1,
		Outcome:        "ai_success",
		FallbackReason: "",
		Provider:       "https://user:secret@api.openai.com/v1/chat/completions?api_key=sk-leak",
		Model:          "glm-5.3-flash",
	})
	snap := store.Snapshot("release", "up", 2.0)
	if snap.LLMCalls.TotalCalls != 1 || snap.LLMCalls.SuccessCount != 1 || snap.LLMCalls.TotalRetries != 1 {
		t.Fatalf("unexpected LLM call summary: %+v", snap.LLMCalls)
	}
	if len(snap.LLMCalls.Providers) != 1 || snap.LLMCalls.Providers[0] != "https://api.openai.com" {
		t.Fatalf("provider must be sanitized to scheme+host only, got %v", snap.LLMCalls.Providers)
	}
}

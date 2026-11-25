package services

import (
	"crypto/rand"
	"encoding/hex"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cybersim/dto"
)

const (
	// InternalNFRProbeHeader identifies synthetic in-process audit/concurrency probe requests
	// so they are never counted in passive live user telemetry.
	InternalNFRProbeHeader = "X-SOC-Internal-NFR-Probe"

	DefaultHTTPTelemetryCapacity     = 10000
	DefaultLLMTelemetryCapacity      = 500
	DefaultEndpointHistoryCapacity   = 1000
	telemetryRouteLLMSimulationStart = "POST /api/simulation/start"
)

type httpSample struct {
	Route      string
	Category   string
	StatusCode int
	DurationMs float64
	RecordedAt time.Time
}

// LLMCallRecord captures timing and outcome metadata for one LLM scenario generation call.
type LLMCallRecord struct {
	DurationMs     float64
	Attempts       int
	Retries        int
	Outcome        string // ai_success | fallback | error
	FallbackReason string
	Provider       string
	Model          string
	RecordedAt     time.Time
}

type routeAggregate struct {
	Route      string
	Category   string
	Count      int
	ErrorCount int
	Durations  []float64
	NextIdx    int
}

// NFRTelemetryStore collects passive runtime HTTP and LLM telemetry inside bounded in-memory ring buffers.
type NFRTelemetryStore struct {
	mu               sync.RWMutex
	processStartedAt time.Time
	windowStartedAt  time.Time
	probeToken       string

	httpCap  int
	httpRing []httpSample
	httpIdx  int
	httpLen  int

	llmCap  int
	llmRing []LLMCallRecord
	llmIdx  int
	llmLen  int

	totalRequests  int
	nonLLMRequests int
	llmRequests    int
	status2xx      int
	status3xx      int
	status4xx      int
	status5xx      int

	inFlightCurrent        atomic.Int64
	peakConcurrentRequests atomic.Int64

	routes             map[string]*routeAggregate
	latestScalability  *dto.NFRScalabilityBenchmarkDTO
}

var (
	globalNFRTelemetryOnce sync.Once
	globalNFRTelemetryInst *NFRTelemetryStore
)

// GlobalNFRTelemetry returns the process-wide runtime NFR telemetry collector.
func GlobalNFRTelemetry() *NFRTelemetryStore {
	globalNFRTelemetryOnce.Do(func() {
		globalNFRTelemetryInst = NewNFRTelemetryStore(DefaultHTTPTelemetryCapacity)
	})
	return globalNFRTelemetryInst
}

// NewNFRTelemetryStore initializes a bounded NFR telemetry store with an unspoofable probe token.
func NewNFRTelemetryStore(capacity int) *NFRTelemetryStore {
	if capacity <= 0 {
		capacity = DefaultHTTPTelemetryCapacity
	}
	now := time.Now().UTC()
	return &NFRTelemetryStore{
		processStartedAt: now,
		windowStartedAt:  now,
		probeToken:       generateRandomHexToken(32),
		httpCap:          capacity,
		httpRing:         make([]httpSample, capacity),
		llmCap:           DefaultLLMTelemetryCapacity,
		llmRing:          make([]LLMCallRecord, DefaultLLMTelemetryCapacity),
		routes:           make(map[string]*routeAggregate),
	}
}

func generateRandomHexToken(numBytes int) string {
	buf := make([]byte, numBytes)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(buf)
}

// RoundTelemetry2 rounds val to 2 decimal places, guarding against NaN/Inf.
func RoundTelemetry2(val float64) float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0
	}
	return math.Round(val*100) / 100
}

// SanitizeProviderHost reduces a raw LLM endpoint URL to scheme://host so paths, query tokens, or userinfo are never retained.
func SanitizeProviderHost(rawURL string) string {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return ""
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return "unknown-provider"
	}
	scheme := parsed.Scheme
	if scheme == "" {
		scheme = "https"
	}
	return scheme + "://" + parsed.Host
}

// ProcessStartedAt returns the immutable timestamp when the telemetry store started.
func (s *NFRTelemetryStore) ProcessStartedAt() time.Time {
	return s.processStartedAt
}

// UptimeSeconds returns whole seconds elapsed since process startup.
func (s *NFRTelemetryStore) UptimeSeconds() int64 {
	sec := int64(time.Since(s.processStartedAt).Seconds())
	if sec < 0 {
		return 0
	}
	return sec
}

// InternalProbeToken returns the per-process secret header value used to isolate internal probes.
func (s *NFRTelemetryStore) InternalProbeToken() string {
	return s.probeToken
}

// IsInternalProbeRequest reports whether req carries the authentic per-process probe isolation token.
func (s *NFRTelemetryStore) IsInternalProbeRequest(req *http.Request) bool {
	if req == nil {
		return false
	}
	headerVal := strings.TrimSpace(req.Header.Get(InternalNFRProbeHeader))
	return headerVal != "" && headerVal == s.probeToken
}

// IncrementInFlight updates active and peak concurrency counters for passive requests.
func (s *NFRTelemetryStore) IncrementInFlight() {
	cur := s.inFlightCurrent.Add(1)
	for {
		peak := s.peakConcurrentRequests.Load()
		if cur <= peak || s.peakConcurrentRequests.CompareAndSwap(peak, cur) {
			break
		}
	}
}

// DecrementInFlight decrements the active in-flight request counter.
func (s *NFRTelemetryStore) DecrementInFlight() {
	if s.inFlightCurrent.Add(-1) < 0 {
		s.inFlightCurrent.Store(0)
	}
}

// RecordHTTPRequest records one completed HTTP request into the bounded ring buffer.
func (s *NFRTelemetryStore) RecordHTTPRequest(route string, statusCode int, durationMs float64) {
	if durationMs < 0 || math.IsNaN(durationMs) || math.IsInf(durationMs, 0) {
		durationMs = 0
	}
	normRoute := strings.TrimSpace(route)
	if normRoute == "" {
		normRoute = "<unmatched>"
	}
	category := "non_llm_api"
	if normRoute == telemetryRouteLLMSimulationStart {
		category = "llm_simulation_start"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalRequests++
	if category == "llm_simulation_start" {
		s.llmRequests++
	} else {
		s.nonLLMRequests++
	}

	switch {
	case statusCode >= 500:
		s.status5xx++
	case statusCode >= 400:
		s.status4xx++
	case statusCode >= 300:
		s.status3xx++
	default:
		s.status2xx++
	}

	s.httpRing[s.httpIdx] = httpSample{
		Route:      normRoute,
		Category:   category,
		StatusCode: statusCode,
		DurationMs: durationMs,
		RecordedAt: time.Now().UTC(),
	}
	s.httpIdx = (s.httpIdx + 1) % s.httpCap
	if s.httpLen < s.httpCap {
		s.httpLen++
	}

	agg, exists := s.routes[normRoute]
	if !exists {
		agg = &routeAggregate{
			Route:     normRoute,
			Category:  category,
			Durations: make([]float64, 0, 64),
		}
		s.routes[normRoute] = agg
	}
	agg.Count++
	if statusCode >= 500 {
		agg.ErrorCount++
	}
	if len(agg.Durations) < DefaultEndpointHistoryCapacity {
		agg.Durations = append(agg.Durations, durationMs)
	} else {
		agg.Durations[agg.NextIdx] = durationMs
		agg.NextIdx = (agg.NextIdx + 1) % DefaultEndpointHistoryCapacity
	}
}

// RecordLLMCall stores sanitized LLM generation latency and outcome metadata.
func (s *NFRTelemetryStore) RecordLLMCall(rec LLMCallRecord) {
	if rec.DurationMs < 0 || math.IsNaN(rec.DurationMs) || math.IsInf(rec.DurationMs, 0) {
		rec.DurationMs = 0
	}
	if rec.RecordedAt.IsZero() {
		rec.RecordedAt = time.Now().UTC()
	}
	rec.Provider = SanitizeProviderHost(rec.Provider)
	rec.Model = strings.TrimSpace(rec.Model)
	if rec.Outcome == "" {
		rec.Outcome = "error"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.llmRing[s.llmIdx] = rec
	s.llmIdx = (s.llmIdx + 1) % s.llmCap
	if s.llmLen < s.llmCap {
		s.llmLen++
	}
}

// ResetWindow clears passive HTTP and LLM telemetry counters while preserving processStartedAt.
func (s *NFRTelemetryStore) ResetWindow() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	s.windowStartedAt = now
	s.httpIdx = 0
	s.httpLen = 0
	s.llmIdx = 0
	s.llmLen = 0
	s.totalRequests = 0
	s.nonLLMRequests = 0
	s.llmRequests = 0
	s.status2xx = 0
	s.status3xx = 0
	s.status4xx = 0
	s.status5xx = 0
	s.peakConcurrentRequests.Store(s.inFlightCurrent.Load())
	s.routes = make(map[string]*routeAggregate)
	return now
}

// SetScalabilityBenchmark caches the most recent in-app concurrency probe result in memory.
func (s *NFRTelemetryStore) SetScalabilityBenchmark(bench *dto.NFRScalabilityBenchmarkDTO) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if bench == nil {
		s.latestScalability = nil
		return
	}
	cp := *bench
	cp.TargetEndpoints = append([]string(nil), bench.TargetEndpoints...)
	cp.Levels = append([]dto.NFRScalabilityLevelDTO(nil), bench.Levels...)
	s.latestScalability = &cp
}

// GetScalabilityBenchmark returns a copy of the latest cached in-app concurrency probe result, if any.
func (s *NFRTelemetryStore) GetScalabilityBenchmark() *dto.NFRScalabilityBenchmarkDTO {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.latestScalability == nil {
		return nil
	}
	cp := *s.latestScalability
	cp.TargetEndpoints = append([]string(nil), s.latestScalability.TargetEndpoints...)
	cp.Levels = append([]dto.NFRScalabilityLevelDTO(nil), s.latestScalability.Levels...)
	return &cp
}

// ComputeLatencyPercentiles computes R-7 latency percentiles in milliseconds (nil fields when empty).
func ComputeLatencyPercentiles(values []float64) dto.NFRPercentilesDTO {
	n := len(values)
	if n == 0 {
		return dto.NFRPercentilesDTO{Count: 0}
	}
	sorted := make([]float64, n)
	copy(sorted, values)
	sort.Float64s(sorted)

	sum := 0.0
	for _, v := range sorted {
		sum += v
	}

	minV := RoundTelemetry2(sorted[0])
	maxV := RoundTelemetry2(sorted[n-1])
	meanV := RoundTelemetry2(sum / float64(n))
	p50 := RoundTelemetry2(quantileR7(sorted, 0.50))
	p95 := RoundTelemetry2(quantileR7(sorted, 0.95))
	p99 := RoundTelemetry2(quantileR7(sorted, 0.99))

	return dto.NFRPercentilesDTO{
		Count: n,
		Min:   &minV,
		Mean:  &meanV,
		P50:   &p50,
		P95:   &p95,
		P99:   &p99,
		Max:   &maxV,
	}
}

// Snapshot returns a point-in-time read-only view of passive HTTP and LLM telemetry.
func (s *NFRTelemetryStore) Snapshot(ginMode, dbStatus string, dbLatencyMs float64) dto.NFRRuntimeSnapshotDTO {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nonLLMDurations := make([]float64, 0, s.httpLen)
	simRouteDurations := make([]float64, 0)

	for i := 0; i < s.httpLen; i++ {
		sample := s.httpRing[i]
		if sample.Category == "llm_simulation_start" {
			simRouteDurations = append(simRouteDurations, sample.DurationMs)
		} else {
			nonLLMDurations = append(nonLLMDurations, sample.DurationMs)
		}
	}

	llmCallDurations := make([]float64, 0, s.llmLen)
	llmSummary := dto.NFRLLMTelemetryDTO{
		Providers:     []string{},
		Models:        []string{},
		OutcomeCounts: make(map[string]int),
	}
	providerSet := make(map[string]bool)
	modelSet := make(map[string]bool)

	for i := 0; i < s.llmLen; i++ {
		rec := s.llmRing[i]
		llmSummary.TotalCalls++
		llmSummary.TotalAttempts += rec.Attempts
		llmSummary.TotalRetries += rec.Retries
		llmSummary.OutcomeCounts[rec.Outcome]++
		switch rec.Outcome {
		case "ai_success":
			llmSummary.SuccessCount++
		case "fallback":
			llmSummary.FallbackCount++
		default:
			llmSummary.ErrorCount++
		}
		if rec.Provider != "" && !providerSet[rec.Provider] {
			providerSet[rec.Provider] = true
			llmSummary.Providers = append(llmSummary.Providers, rec.Provider)
		}
		if rec.Model != "" && !modelSet[rec.Model] {
			modelSet[rec.Model] = true
			llmSummary.Models = append(llmSummary.Models, rec.Model)
		}
		llmCallDurations = append(llmCallDurations, rec.DurationMs)
	}
	sort.Strings(llmSummary.Providers)
	sort.Strings(llmSummary.Models)
	llmSummary.Latency = ComputeLatencyPercentiles(llmCallDurations)

	// Prefer direct LLM call latency if recorded; otherwise surface route latency for POST /api/simulation/start.
	llmPercentiles := llmSummary.Latency
	if llmPercentiles.Count == 0 && len(simRouteDurations) > 0 {
		llmPercentiles = ComputeLatencyPercentiles(simRouteDurations)
	}

	endpoints := make([]dto.NFREndpointLatencyDTO, 0, len(s.routes))
	for _, agg := range s.routes {
		pct := ComputeLatencyPercentiles(agg.Durations)
		endpoints = append(endpoints, dto.NFREndpointLatencyDTO{
			Route:      agg.Route,
			Category:   agg.Category,
			Count:      agg.Count,
			ErrorCount: agg.ErrorCount,
			MeanMs:     pct.Mean,
			P50Ms:      pct.P50,
			P95Ms:      pct.P95,
			P99Ms:      pct.P99,
			MaxMs:      pct.Max,
		})
	}
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Count != endpoints[j].Count {
			return endpoints[i].Count > endpoints[j].Count
		}
		return endpoints[i].Route < endpoints[j].Route
	})

	var availPtr *float64
	var err5xxPtr *float64
	if s.totalRequests > 0 {
		avail := RoundTelemetry2(float64(s.totalRequests-s.status5xx) / float64(s.totalRequests) * 100.0)
		err5xx := RoundTelemetry2(float64(s.status5xx) / float64(s.totalRequests) * 100.0)
		availPtr = &avail
		err5xxPtr = &err5xx
	}

	return dto.NFRRuntimeSnapshotDTO{
		ProcessStartedAt:       s.processStartedAt,
		WindowStartedAt:        s.windowStartedAt,
		UptimeSeconds:          s.UptimeSeconds(),
		GinMode:                ginMode,
		TotalRequests:          s.totalRequests,
		NonLLMRequests:         s.nonLLMRequests,
		LLMRequests:            s.llmRequests,
		Status2xx:              s.status2xx,
		Status3xx:              s.status3xx,
		Status4xx:              s.status4xx,
		Status5xx:              s.status5xx,
		AvailabilityPct:        availPtr,
		ErrorRate5xxPct:        err5xxPtr,
		InFlightCurrent:        int(s.inFlightCurrent.Load()),
		PeakConcurrentRequests: int(s.peakConcurrentRequests.Load()),
		DBStatus:               dbStatus,
		DBLatencyMs:            RoundTelemetry2(dbLatencyMs),
		NonLLMLatency:          ComputeLatencyPercentiles(nonLLMDurations),
		LLMLatency:             llmPercentiles,
		LLMCalls:               llmSummary,
		Endpoints:              endpoints,
	}
}

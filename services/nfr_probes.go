package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cybersim/dto"

	"github.com/golang-jwt/jwt/v5"
)

const (
	MaxSafeProbeWorkers  = 20
	MaxSafeProbeRequests = 200
	SafeProbeTimeout     = 15 * time.Second
)

var probeTargetEndpoints = []string{
	"/api/health",
	"/api/exams/sample",
	"/api/learn/domains",
}

// resolveProbeJWTSecret mirrors middlewares.getJWTSecret without creating an import cycle.
func resolveProbeJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("supersecretkey")
	}
	return []byte(secret)
}

func signSyntheticProbeJWT(userID, role string) string {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"role":     role,
		"nickname": "nfr_probe_" + role,
		"exp":      time.Now().Add(30 * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(resolveProbeJWTSecret())
	if err != nil {
		return ""
	}
	return signed
}

func execInternalProbeRequest(
	ctx context.Context,
	router http.Handler,
	telemetry *NFRTelemetryStore,
	method, path, bearer string,
) int {
	req, err := http.NewRequestWithContext(ctx, method, path, nil)
	if err != nil {
		return http.StatusInternalServerError
	}
	if telemetry != nil {
		req.Header.Set(InternalNFRProbeHeader, telemetry.InternalProbeToken())
	}
	if bearer != "" {
		req.Header.Set("Authorization", bearer)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec.Code
}

// RunInProcessSecurityProbe executes 6 read-only httptest probes against the live router
// using the internal probe isolation header so live passive telemetry remains untouched.
func RunInProcessSecurityProbe(
	router http.Handler,
	adminBearer string,
	telemetry *NFRTelemetryStore,
) []dto.NFRCheckItemDTO {
	if router == nil {
		return []dto.NFRCheckItemDTO{
			{
				Key:      "security_probe_router",
				Measured: "router unavailable",
				Unit:     "status",
				Target:   "live Gin router",
				Status:   nfrStatusNotMeasured,
			},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	learnerJWT := signSyntheticProbeJWT("00000000-0000-0000-0000-000000000001", "learner")
	specialistJWT := signSyntheticProbeJWT("00000000-0000-0000-0000-000000000002", "specialist")
	if strings.TrimSpace(adminBearer) == "" {
		if tok := signSyntheticProbeJWT("00000000-0000-0000-0000-000000000003", "admin"); tok != "" {
			adminBearer = "Bearer " + tok
		}
	}

	type probeSpec struct {
		key      string
		path     string
		bearer   string
		target   string
		expected func(code int) bool
	}

	specs := []probeSpec{
		{
			key:      "unauth_admin_nfr_blocked",
			path:     "/api/admin/nfr/live",
			bearer:   "",
			target:   "401",
			expected: func(code int) bool { return code == http.StatusUnauthorized },
		},
		{
			key:      "forged_jwt_rejected",
			path:     "/api/admin/nfr/live",
			bearer:   "Bearer invalid.jwt.token",
			target:   "401",
			expected: func(code int) bool { return code == http.StatusUnauthorized },
		},
		{
			key:      "learner_rbac_denied",
			path:     "/api/admin/nfr/live",
			bearer:   "Bearer " + learnerJWT,
			target:   "403",
			expected: func(code int) bool { return code == http.StatusForbidden },
		},
		{
			key:      "specialist_rbac_denied",
			path:     "/api/admin/nfr/live",
			bearer:   "Bearer " + specialistJWT,
			target:   "403",
			expected: func(code int) bool { return code == http.StatusForbidden },
		},
		{
			key:      "unauth_exam_blocked",
			path:     "/api/exams/session",
			bearer:   "",
			target:   "401",
			expected: func(code int) bool { return code == http.StatusUnauthorized },
		},
		{
			key:      "sqli_query_param_safe",
			path:     "/api/learn/domains?lang=%27%20OR%20%271%27%3D%271",
			bearer:   adminBearer,
			target:   "< 500",
			expected: func(code int) bool { return code > 0 && code < http.StatusInternalServerError },
		},
	}

	results := make([]dto.NFRCheckItemDTO, 0, len(specs))
	for _, spec := range specs {
		code := execInternalProbeRequest(ctx, router, telemetry, http.MethodGet, spec.path, spec.bearer)
		status := nfrStatusFail
		if spec.expected(code) {
			status = nfrStatusPass
		}
		results = append(results, dto.NFRCheckItemDTO{
			Key:      spec.key,
			Measured: fmt.Sprintf("%d", code),
			Unit:     "status",
			Target:   spec.target,
			Status:   status,
		})
	}
	return results
}

// ResolveScalabilityProbeParams resolves preset names ("light", "standard", "burst") or
// bounded custom worker/request overrides while enforcing MaxSafeProbeWorkers and MaxSafeProbeRequests.
func ResolveScalabilityProbeParams(preset string, customWorkers, customReqs int) (string, int, int, error) {
	if customWorkers != 0 || customReqs != 0 {
		if customWorkers <= 0 || customWorkers > MaxSafeProbeWorkers || customReqs <= 0 || customReqs > MaxSafeProbeRequests {
			return "", 0, 0, fmt.Errorf("custom probe parameters exceed safe caps (max %d workers, %d requests)", MaxSafeProbeWorkers, MaxSafeProbeRequests)
		}
		label := strings.TrimSpace(preset)
		if label == "" {
			label = "custom"
		}
		return label, customWorkers, customReqs, nil
	}
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "light":
		return "light", 5, 50, nil
	case "", "standard":
		return "standard", 10, 100, nil
	case "burst":
		return "burst", 20, 200, nil
	default:
		return "", 0, 0, fmt.Errorf("invalid scalability preset %q", preset)
	}
}

// RunSafeScalabilityProbe runs a bounded, read-only, telemetry-isolated concurrency wave against the live Gin router.
func RunSafeScalabilityProbe(
	router http.Handler,
	adminBearer string,
	telemetry *NFRTelemetryStore,
	preset string,
	customWorkers int,
	customReqs int,
) (dto.NFRScalabilityBenchmarkDTO, error) {
	if router == nil {
		return dto.NFRScalabilityBenchmarkDTO{}, fmt.Errorf("router is unavailable")
	}
	normPreset, workers, totalReqs, err := ResolveScalabilityProbeParams(preset, customWorkers, customReqs)
	if err != nil {
		return dto.NFRScalabilityBenchmarkDTO{}, err
	}

	if strings.TrimSpace(adminBearer) == "" {
		if tok := signSyntheticProbeJWT("00000000-0000-0000-0000-000000000003", "admin"); tok != "" {
			adminBearer = "Bearer " + tok
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), SafeProbeTimeout)
	defer cancel()

	var nextIdx atomic.Int64
	durations := make([]float64, totalReqs)
	codes := make([]int, totalReqs)

	startAll := time.Now()
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for {
				idx := int(nextIdx.Add(1) - 1)
				if idx >= totalReqs {
					return
				}
				if ctx.Err() != nil {
					codes[idx] = http.StatusGatewayTimeout
					durations[idx] = float64(SafeProbeTimeout.Milliseconds())
					continue
				}
				target := probeTargetEndpoints[idx%len(probeTargetEndpoints)]
				t0 := time.Now()
				code := execInternalProbeRequest(ctx, router, telemetry, http.MethodGet, target, adminBearer)
				durations[idx] = float64(time.Since(t0).Microseconds()) / 1000.0
				codes[idx] = code
			}
		}()
	}
	wg.Wait()
	elapsedMs := float64(time.Since(startAll).Microseconds()) / 1000.0
	if elapsedMs <= 0 {
		elapsedMs = 1
	}

	successCount := 0
	errorCount := 0
	for _, code := range codes {
		if code >= 200 && code < 400 {
			successCount++
		} else {
			errorCount++
		}
	}

	pct := ComputeLatencyPercentiles(durations)
	errPct := RoundTelemetry2(float64(errorCount) / float64(totalReqs) * 100.0)
	rps := RoundTelemetry2(float64(totalReqs) / (elapsedMs / 1000.0))

	level := dto.NFRScalabilityLevelDTO{
		Concurrency:   workers,
		TotalRequests: totalReqs,
		SuccessCount:  successCount,
		ErrorCount:    errorCount,
		ErrorRatePct:  errPct,
		ThroughputRPS: rps,
		MeanMs:        ptrValOrZero(pct.Mean),
		P50Ms:         ptrValOrZero(pct.P50),
		P95Ms:         ptrValOrZero(pct.P95),
		P99Ms:         ptrValOrZero(pct.P99),
		MaxMs:         ptrValOrZero(pct.Max),
		DurationMs:    RoundTelemetry2(elapsedMs),
	}

	bench := dto.NFRScalabilityBenchmarkDTO{
		Mode:            "in_app_probe",
		Preset:          normPreset,
		ExecutedAt:      time.Now().UTC(),
		TargetEndpoints: append([]string(nil), probeTargetEndpoints...),
		Levels:          []dto.NFRScalabilityLevelDTO{level},
	}
	if telemetry != nil {
		telemetry.SetScalabilityBenchmark(&bench)
	}
	return bench, nil
}

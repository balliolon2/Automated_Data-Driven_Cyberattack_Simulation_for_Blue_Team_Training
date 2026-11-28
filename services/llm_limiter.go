package services

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	DefaultLLMMaxConcurrency = 8
	MaxLLMMaxConcurrency     = 64
)

// LLMLimiter enforces non-blocking bounded concurrency around runtime LLM scenario synthesis.
type LLMLimiter struct {
	sem chan struct{}
}

// ResolveBoundedIntEnv reads an integer environment variable clamped to [minVal, maxVal].
// Missing, unparseable, or below-min values return defaultVal cleanly.
func ResolveBoundedIntEnv(key string, defaultVal, minVal, maxVal int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < minVal {
		return defaultVal
	}
	if parsed > maxVal {
		return maxVal
	}
	return parsed
}

// ResolveLLMMaxConcurrency reads SIMULATION_LLM_MAX_CONCURRENCY (0 disables live LLM; default 8; max 64).
func ResolveLLMMaxConcurrency() int {
	return ResolveBoundedIntEnv(
		"SIMULATION_LLM_MAX_CONCURRENCY",
		DefaultLLMMaxConcurrency,
		0,
		MaxLLMMaxConcurrency,
	)
}

// NewLLMLimiter constructs a buffered semaphore limiter with the requested capacity.
// Capacity <= 0 disables all acquisitions cleanly.
func NewLLMLimiter(capacity int) *LLMLimiter {
	if capacity <= 0 {
		return &LLMLimiter{}
	}
	return &LLMLimiter{
		sem: make(chan struct{}, capacity),
	}
}

// TryAcquire attempts to reserve one permit without blocking callers.
// Returned release functions are idempotent and safe under concurrent calls.
func (l *LLMLimiter) TryAcquire() (release func(), acquired bool) {
	if l == nil || l.sem == nil {
		return nil, false
	}
	select {
	case l.sem <- struct{}{}:
		var once sync.Once
		return func() {
			once.Do(func() {
				<-l.sem
			})
		}, true
	default:
		return nil, false
	}
}

var (
	defaultLLMLimiterOnce sync.Once
	defaultLLMLimiter     *LLMLimiter
)

// TryAcquireLLMSlot acquires a permit from the process-wide default LLM concurrency limiter.
func TryAcquireLLMSlot() (release func(), acquired bool) {
	defaultLLMLimiterOnce.Do(func() {
		defaultLLMLimiter = NewLLMLimiter(ResolveLLMMaxConcurrency())
	})
	return defaultLLMLimiter.TryAcquire()
}

package services

import (
	"os"
	"strconv"
	"strings"

	"cybersim/dto"
)

// NFRRubricVersion identifies the active numeric acceptance ruleset.
const NFRRubricVersion = "nfr-rubric-v1"

// DefaultNFRThresholds returns the canonical nfr-rubric-v1 target and warning cutoffs.
func DefaultNFRThresholds() dto.NFRThresholdsDTO {
	return dto.NFRThresholdsDTO{
		RubricVersion:             NFRRubricVersion,
		NonLLMP95TargetMs:         300.0,
		NonLLMP95WarnMs:           500.0,
		NonLLMP99TargetMs:         800.0,
		NonLLMP99WarnMs:           1200.0,
		LLMP95TargetMs:            20000.0,
		LLMP95WarnMs:              45000.0,
		DBPingTargetMs:            50.0,
		DBPingWarnMs:              150.0,
		AvailabilityTargetPct:     99.0,
		AvailabilityWarnPct:       95.0,
		ScalabilityP95TargetMs:    500.0,
		ScalabilityP95WarnMs:      1000.0,
		ScalabilityErrorTargetPct: 1.0,
		ScalabilityErrorWarnPct:   5.0,
		CoverageTargetPct:         60.0,
		CoverageWarnPct:           40.0,
		CSUQTargetMean:            5.0,
		CSUQWarnMean:              4.0,
		TLXTargetMax:              55.0,
		TLXWarnMax:                70.0,
	}
}

func envPositiveFloat(key string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

// LoadNFRThresholdsFromEnv reads canonical thresholds and applies validated positive environment overrides.
func LoadNFRThresholdsFromEnv() dto.NFRThresholdsDTO {
	t := DefaultNFRThresholds()
	t.NonLLMP95TargetMs = envPositiveFloat("NFR_P95_TARGET_MS", t.NonLLMP95TargetMs)
	if t.NonLLMP95WarnMs < t.NonLLMP95TargetMs {
		t.NonLLMP95WarnMs = t.NonLLMP95TargetMs * 1.5
	}
	t.NonLLMP99TargetMs = envPositiveFloat("NFR_P99_TARGET_MS", t.NonLLMP99TargetMs)
	if t.NonLLMP99WarnMs < t.NonLLMP99TargetMs {
		t.NonLLMP99WarnMs = t.NonLLMP99TargetMs * 1.5
	}
	t.LLMP95TargetMs = envPositiveFloat("NFR_LLM_P95_TARGET_MS", t.LLMP95TargetMs)
	if t.LLMP95WarnMs < t.LLMP95TargetMs {
		t.LLMP95WarnMs = t.LLMP95TargetMs * 2.0
	}
	t.DBPingTargetMs = envPositiveFloat("NFR_DB_PING_TARGET_MS", t.DBPingTargetMs)
	if t.DBPingWarnMs < t.DBPingTargetMs {
		t.DBPingWarnMs = t.DBPingTargetMs * 3.0
	}
	t.AvailabilityTargetPct = envPositiveFloat("NFR_AVAILABILITY_TARGET_PCT", t.AvailabilityTargetPct)
	if t.AvailabilityTargetPct > 100 {
		t.AvailabilityTargetPct = 100
	}
	if t.AvailabilityWarnPct > t.AvailabilityTargetPct {
		t.AvailabilityWarnPct = t.AvailabilityTargetPct - 4.0
		if t.AvailabilityWarnPct < 0 {
			t.AvailabilityWarnPct = 0
		}
	}
	t.ScalabilityP95TargetMs = envPositiveFloat("NFR_SCALABILITY_P95_MS", t.ScalabilityP95TargetMs)
	if t.ScalabilityP95WarnMs < t.ScalabilityP95TargetMs {
		t.ScalabilityP95WarnMs = t.ScalabilityP95TargetMs * 2.0
	}
	t.ScalabilityErrorTargetPct = envPositiveFloat("NFR_SCALABILITY_ERR_PCT", t.ScalabilityErrorTargetPct)
	t.CoverageTargetPct = envPositiveFloat("NFR_COVERAGE_TARGET_PCT", t.CoverageTargetPct)
	if t.CoverageTargetPct > 100 {
		t.CoverageTargetPct = 100
	}
	if t.CoverageWarnPct > t.CoverageTargetPct {
		t.CoverageWarnPct = t.CoverageTargetPct * 0.66
	}
	return t
}

// NFRPostureInput records boolean/length security posture flags without exposing any secret material.
type NFRPostureInput struct {
	JWTSecretSet            bool
	JWTSecretIsDefault      bool
	JWTSecretLength         int
	SwaggerEnabled          bool
	BypassPostTestThreshold bool
	CORSWildcard            bool
	GinMode                 string
}

// LoadNFRPostureInput inspects active runtime environment flags safely without retaining raw secret values.
func LoadNFRPostureInput(ginMode string) NFRPostureInput {
	rawJWT := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	isDefault := rawJWT == "" || rawJWT == "supersecretkey" || rawJWT == "secret_jwt_key_default"
	frontendURL := strings.TrimSpace(os.Getenv("FRONTEND_URL"))

	return NFRPostureInput{
		JWTSecretSet:            rawJWT != "",
		JWTSecretIsDefault:      isDefault,
		JWTSecretLength:         len(rawJWT),
		SwaggerEnabled:          strings.EqualFold(strings.TrimSpace(os.Getenv("ENABLE_SWAGGER")), "true"),
		BypassPostTestThreshold: strings.EqualFold(strings.TrimSpace(os.Getenv("BYPASS_POST_TEST_THRESHOLD")), "true"),
		CORSWildcard:            frontendURL == "*",
		GinMode:                 strings.TrimSpace(ginMode),
	}
}

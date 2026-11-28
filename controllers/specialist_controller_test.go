package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestSessionSnapshotDrivesSpecialistEvidenceReview verifies that the review
// uses the immutable ScenarioSnapshot rather than the live scenarios row. The
// two scenarios deliberately disagree about which log index proves the finding.
func TestSessionSnapshotDrivesSpecialistEvidenceReview(t *testing.T) {
	snapshotScenario, err := decodeScenarioSnapshot(models.ScenarioSnapshot{
		SnapshotID: "snap-1",
		ScenarioID: "scen-snapshot",
		ScenarioData: map[string]interface{}{
			"scenario_id": "scen-snapshot",
			"title":       "Snapshot title",
			"initial_logs": []interface{}{
				map[string]interface{}{"message": "snapshot index zero"},
				map[string]interface{}{"message": "snapshot evidence"},
			},
			"expected_outcomes": map[string]interface{}{
				"key_findings": []interface{}{
					map[string]interface{}{
						"id":                   "f1",
						"description":          "Snapshot finding",
						"domain_id":            "domain1",
						"points":               10,
						"evidence_log_indices": []interface{}{1},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("decodeScenarioSnapshot() error = %v", err)
	}

	liveScenario := models.Scenario{
		ScenarioID: "scen-snapshot",
		Title:      "Live edited title",
		InitialLogs: []interface{}{
			map[string]interface{}{"message": "live evidence"},
			map[string]interface{}{"message": "live index one"},
		},
		ExpectedOutcomes: map[string]interface{}{
			"key_findings": []interface{}{
				map[string]interface{}{
					"id":                   "f1",
					"description":          "Live finding",
					"domain_id":            "domain1",
					"points":               10,
					"evidence_log_indices": []interface{}{0},
				},
			},
		},
	}

	actions := []models.SessionAction{
		{
			ActionType: services.SelectEvidenceActionType,
			StepOrder:  1,
			Payload:    map[string]interface{}{"operation": "toggle", "log_index": 1, "selected": true},
		},
	}
	session := models.SimulationSession{
		SessionID:     "sess-snapshot",
		ScenarioID:    "scen-snapshot",
		RubricVersion: services.RubricVersion,
	}

	snapshotReview := services.BuildEvidenceReviewForSession(snapshotScenario, session, actions)
	liveReview := services.BuildEvidenceReviewForSession(liveScenario, session, actions)

	if len(snapshotReview.EvidenceFindings) != 1 || !snapshotReview.EvidenceFindings[0].Found {
		t.Fatalf("snapshot review must credit the finding from snapshot evidence, got %+v", snapshotReview.EvidenceFindings)
	}
	if !reflect.DeepEqual(snapshotReview.EvidenceFindings[0].MatchedLogIndices, []int{1}) {
		t.Fatalf("snapshot matched indices = %v, want [1]", snapshotReview.EvidenceFindings[0].MatchedLogIndices)
	}
	if got := snapshotReview.EvidenceFindings[0].EvidenceLogs[0].Message; got != "snapshot evidence" {
		t.Fatalf("snapshot evidence message = %q, want snapshot evidence", got)
	}
	if snapshotReview.NoiseSelectedCount != 0 || snapshotReview.EvidencePenalty != 0 {
		t.Fatalf("snapshot review noise = %d penalty = %f, want 0/0",
			snapshotReview.NoiseSelectedCount, snapshotReview.EvidencePenalty)
	}

	if liveReview.EvidenceFindings[0].Found {
		t.Fatal("live row would credit index 0, so using it would diverge from the scored snapshot")
	}
	if liveReview.NoiseSelectedCount != 1 || liveReview.EvidencePenalty != 2 {
		t.Fatalf("live review noise = %d penalty = %f, want 1/2",
			liveReview.NoiseSelectedCount, liveReview.EvidencePenalty)
	}
}

type fakeSpecialistRegStore struct {
	auth          *fakeAuthStore
	forceCreateTx error
	applications  map[string]*models.SpecialistApplication
}

func newFakeSpecialistRegStore() *fakeSpecialistRegStore {
	return &fakeSpecialistRegStore{
		auth:         newFakeAuthStore(),
		applications: make(map[string]*models.SpecialistApplication),
	}
}

func (f *fakeSpecialistRegStore) FindUserByEmail(email string) (*models.User, error) {
	return f.auth.FindUserByEmail(email)
}

func (f *fakeSpecialistRegStore) FindUserByNickname(nickname string) (*models.User, error) {
	return f.auth.FindUserByNickname(nickname)
}

func (f *fakeSpecialistRegStore) CreateUserAndApplication(user *models.User, app *models.SpecialistApplication) error {
	if f.forceCreateTx != nil {
		return f.forceCreateTx
	}
	if err := f.auth.CreateUser(user); err != nil {
		return err
	}
	app.UserID = user.UserID
	copyApp := *app
	f.applications[app.ApplicationID] = &copyApp
	return nil
}

func buildSpecialistMultipartRequest(t *testing.T, fields map[string]string, resumeBytes []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	if resumeBytes != nil {
		fw, err := mw.CreateFormFile("resume", "resume.pdf")
		if err != nil {
			t.Fatalf("CreateFormFile failed: %v", err)
		}
		_, _ = fw.Write(resumeBytes)
	}
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/register-specialist", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestSpecialistController_AtomicRegisterSpecialistAndRollback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("REQUIRE_EMAIL_VERIFICATION", "true")
	t.Setenv("APP_BASE_URL", "https://soc.example.com")

	t.Run("creates_unverified_user_and_application_atomically_and_dispatches_email", func(t *testing.T) {
		tmpUploadDir := t.TempDir()
		store := newFakeSpecialistRegStore()
		mailer := newFakeMailSender()
		sc := NewSpecialistControllerWithDeps(nil, store, mailer, tmpUploadDir)

		r := gin.New()
		r.POST("/api/register-specialist", sc.RegisterSpecialist)

		req := buildSpecialistMultipartRequest(t, map[string]string{
			"email":              "spec@soc.local",
			"password":           "StrongPass123!",
			"nickname":           "spec_one",
			"preferred_language": "th",
			"bio":                "Ten years defending SOC networks and incident response.",
			"linkedin_url":       "https://linkedin.com/in/specone",
		}, []byte("%PDF-1.4 fake resume bytes"))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var body map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		if body["requires_verification"] != true || body["status"] != "pending" {
			t.Fatalf("expected requires_verification=true and status=pending, got %+v", body)
		}

		savedUser, err := store.FindUserByEmail("spec@soc.local")
		if err != nil || savedUser.EmailVerified || savedUser.VerificationTokenHash == nil || *savedUser.VerificationTokenHash == "" {
			t.Fatalf("expected unverified user with hashed token, got %+v err=%v", savedUser, err)
		}
		if len(store.applications) != 1 {
			t.Fatalf("expected 1 persisted application, got %d", len(store.applications))
		}
		entries, _ := os.ReadDir(tmpUploadDir)
		if len(entries) != 1 {
			t.Fatalf("expected 1 uploaded application directory, got %d", len(entries))
		}

		sentMail := mailer.waitForMail(t)
		if sentMail.Recipient != "spec@soc.local" || !strings.Contains(sentMail.URL, "/verify-email?token=") {
			t.Fatalf("unexpected verification email dispatched: %+v", sentMail)
		}
	})

	t.Run("db_transaction_error_cleans_up_uploaded_files_leaving_zero_orphans", func(t *testing.T) {
		tmpUploadDir := t.TempDir()
		store := newFakeSpecialistRegStore()
		store.forceCreateTx = errors.New("simulated database failure")
		mailer := newFakeMailSender()
		sc := NewSpecialistControllerWithDeps(nil, store, mailer, tmpUploadDir)

		r := gin.New()
		r.POST("/api/register-specialist", sc.RegisterSpecialist)

		req := buildSpecialistMultipartRequest(t, map[string]string{
			"email":    "failtx@soc.local",
			"password": "StrongPass123!",
			"nickname": "fail_tx",
			"bio":      "SOC threat hunting analyst.",
		}, []byte("%PDF-1.4 fake resume bytes"))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 on DB tx failure, got %d: %s", w.Code, w.Body.String())
		}

		entries, err := os.ReadDir(tmpUploadDir)
		if err != nil {
			t.Fatalf("ReadDir failed: %v", err)
		}
		if len(entries) != 0 {
			t.Fatalf("expected 0 orphan upload directories after DB rollback, got %d", len(entries))
		}
	})
}

type fakeAdminStore struct {
	apps  map[string]*models.SpecialistApplication
	users map[string]*models.User
}

func (f *fakeAdminStore) GetApplicationByID(appID string) (*models.SpecialistApplication, error) {
	a, ok := f.apps[appID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	c := *a
	return &c, nil
}

func (f *fakeAdminStore) GetUserByID(userID string) (*models.User, error) {
	u, ok := f.users[userID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	c := *u
	return &c, nil
}

func (f *fakeAdminStore) ApproveApplication(app *models.SpecialistApplication, adminID string, now time.Time) error {
	app.Status = "approved"
	app.ReviewedBy = &adminID
	app.ReviewedAt = &now
	app.UpdatedAt = now
	copyApp := *app
	f.apps[app.ApplicationID] = &copyApp
	if u, ok := f.users[app.UserID]; ok {
		u.Role = "specialist"
	}
	return nil
}

func TestAdminController_VerifiedApplicantGate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("list_applications_sql_filters_verified_users_and_qualifies_columns", func(t *testing.T) {
		db, err := gorm.Open(postgres.New(postgres.Config{
			DSN:                  "host=localhost user=test dbname=test port=5432 sslmode=disable",
			PreferSimpleProtocol: true,
		}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
		if err != nil {
			t.Fatalf("gorm.Open DryRun failed: %v", err)
		}
		sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			var apps []models.SpecialistApplication
			return buildVerifiedApplicationsQuery(tx, "pending").Find(&apps)
		})
		if !strings.Contains(sql, "users.email_verified") || !strings.Contains(sql, "specialist_applications.status") || !strings.Contains(sql, "specialist_applications.created_at") {
			t.Fatalf("expected qualified email_verified filter and disambiguated status/created_at, got SQL: %s", sql)
		}
	})

	t.Run("approve_blocks_unverified_applicant_and_allows_verified_applicant", func(t *testing.T) {
		store := &fakeAdminStore{
			apps: map[string]*models.SpecialistApplication{
				"app-unverified": {ApplicationID: "app-unverified", UserID: "u-unver", Status: "pending"},
				"app-verified":   {ApplicationID: "app-verified", UserID: "u-ver", Status: "pending"},
			},
			users: map[string]*models.User{
				"u-unver": {UserID: "u-unver", Email: "unver@soc.local", Role: "learner", EmailVerified: false},
				"u-ver":   {UserID: "u-ver", Email: "ver@soc.local", Role: "learner", EmailVerified: true},
			},
		}
		ac := NewAdminControllerWithStore(nil, store)
		r := gin.New()
		r.POST("/api/admin/applications/:id/approve", func(c *gin.Context) {
			c.Set("user_id", "admin-1")
			ac.Approve(c)
		})

		wBlocked := httptest.NewRecorder()
		reqBlocked := httptest.NewRequest(http.MethodPost, "/api/admin/applications/app-unverified/approve", nil)
		r.ServeHTTP(wBlocked, reqBlocked)
		var blockedBody map[string]any
		_ = json.Unmarshal(wBlocked.Body.Bytes(), &blockedBody)
		if wBlocked.Code != http.StatusBadRequest || blockedBody["code"] != "ADMIN_APPLICANT_EMAIL_UNVERIFIED" {
			t.Fatalf("expected 400 ADMIN_APPLICANT_EMAIL_UNVERIFIED, got %d %+v", wBlocked.Code, blockedBody)
		}

		wOK := httptest.NewRecorder()
		reqOK := httptest.NewRequest(http.MethodPost, "/api/admin/applications/app-verified/approve", nil)
		r.ServeHTTP(wOK, reqOK)
		if wOK.Code != http.StatusOK || store.users["u-ver"].Role != "specialist" {
			t.Fatalf("expected 200 approval and specialist promotion, got %d role=%q", wOK.Code, store.users["u-ver"].Role)
		}
	})
}

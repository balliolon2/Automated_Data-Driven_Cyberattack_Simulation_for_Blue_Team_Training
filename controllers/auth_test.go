package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"cybersim/middlewares"
	"cybersim/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type fakeAuthStore struct {
	usersByID    map[string]*models.User
	appsByUserID map[string]*models.SpecialistApplication
	nextID       int
	forceDBErr   error
}

func newFakeAuthStore() *fakeAuthStore {
	return &fakeAuthStore{
		usersByID:    make(map[string]*models.User),
		appsByUserID: make(map[string]*models.SpecialistApplication),
	}
}

func (f *fakeAuthStore) seedUser(u models.User) *models.User {
	if u.UserID == "" {
		f.nextID++
		u.UserID = "user-" + string(rune('0'+f.nextID))
	}
	copyUser := u
	f.usersByID[u.UserID] = &copyUser
	return &copyUser
}

func (f *fakeAuthStore) FindUserByEmail(email string) (*models.User, error) {
	if f.forceDBErr != nil {
		return nil, f.forceDBErr
	}
	target := strings.TrimSpace(email)
	for _, u := range f.usersByID {
		if strings.EqualFold(u.Email, target) {
			c := *u
			return &c, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeAuthStore) FindUserByNickname(nickname string) (*models.User, error) {
	if f.forceDBErr != nil {
		return nil, f.forceDBErr
	}
	for _, u := range f.usersByID {
		if u.Nickname == nickname {
			c := *u
			return &c, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeAuthStore) FindUserByID(userID string) (*models.User, error) {
	if f.forceDBErr != nil {
		return nil, f.forceDBErr
	}
	u, ok := f.usersByID[userID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	c := *u
	return &c, nil
}

func (f *fakeAuthStore) CreateUser(user *models.User) error {
	if f.forceDBErr != nil {
		return f.forceDBErr
	}
	if user.UserID == "" {
		f.nextID++
		user.UserID = "user-" + time.Now().Format("150405") + "-" + string(rune('a'+f.nextID))
	}
	if user.PreferredLanguage == "" {
		user.PreferredLanguage = "en"
	}
	copyUser := *user
	f.usersByID[user.UserID] = &copyUser
	return nil
}

func (f *fakeAuthStore) UpdateLastLogin(userID string, t time.Time) error {
	if u, ok := f.usersByID[userID]; ok {
		u.LastLogin = &t
	}
	return nil
}

func (f *fakeAuthStore) GetLatestApplicationByUserID(userID string) (*models.SpecialistApplication, error) {
	app, ok := f.appsByUserID[userID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	c := *app
	return &c, nil
}

func (f *fakeAuthStore) IsNicknameTakenByOther(nickname, excludeUserID string) (bool, error) {
	if f.forceDBErr != nil {
		return false, f.forceDBErr
	}
	for id, u := range f.usersByID {
		if id != excludeUserID && u.Nickname == nickname {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeAuthStore) UpdateUserNickname(userID, nickname string) error {
	if f.forceDBErr != nil {
		return f.forceDBErr
	}
	if u, ok := f.usersByID[userID]; ok {
		u.Nickname = nickname
	}
	return nil
}

func (f *fakeAuthStore) UpdateUserLanguage(userID, lang string) (int64, error) {
	if f.forceDBErr != nil {
		return 0, f.forceDBErr
	}
	u, ok := f.usersByID[userID]
	if !ok {
		return 0, nil
	}
	u.PreferredLanguage = lang
	return 1, nil
}

func (f *fakeAuthStore) FindUserByVerificationTokenHash(tokenHash string) (*models.User, error) {
	if f.forceDBErr != nil {
		return nil, f.forceDBErr
	}
	for _, u := range f.usersByID {
		if u.VerificationTokenHash != nil && *u.VerificationTokenHash == tokenHash {
			c := *u
			return &c, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeAuthStore) MarkUserEmailVerified(userID string, verifiedAt time.Time) error {
	if f.forceDBErr != nil {
		return f.forceDBErr
	}
	if u, ok := f.usersByID[userID]; ok {
		u.EmailVerified = true
		u.EmailVerifiedAt = &verifiedAt
		u.VerificationTokenHash = nil
		u.VerificationSentAt = nil
		u.VerificationExpiresAt = nil
	}
	return nil
}

func (f *fakeAuthStore) UpdateVerificationToken(userID, tokenHash string, sentAt, expiresAt time.Time) error {
	if f.forceDBErr != nil {
		return f.forceDBErr
	}
	if u, ok := f.usersByID[userID]; ok {
		hashCopy := tokenHash
		u.VerificationTokenHash = &hashCopy
		u.VerificationSentAt = &sentAt
		u.VerificationExpiresAt = &expiresAt
	}
	return nil
}

func makeAuthRouter(ac *AuthController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	api.POST("/register", ac.Register)
	api.POST("/login", ac.Login)
	api.POST("/verify-email", ac.VerifyEmail)
	api.POST("/resend-verification", ac.ResendVerification)
	userGroup := api.Group("/user", middlewares.AuthMiddleware())
	userGroup.GET("/profile", ac.GetProfile)
	userGroup.PUT("/nickname", ac.UpdateNickname)
	userGroup.PUT("/language", ac.UpdateLanguage)
	return r
}

func doJSONReq(t *testing.T, r http.Handler, method, path string, payload any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var bodyBytes []byte
	if payload != nil {
		switch v := payload.(type) {
		case string:
			bodyBytes = []byte(v)
		default:
			var err error
			bodyBytes, err = json.Marshal(v)
			if err != nil {
				t.Fatalf("marshal payload failed: %v", err)
			}
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAuthController_RegisterLanguageAndErrorCodes(t *testing.T) {
	store := newFakeAuthStore()
	ac := NewAuthControllerWithStore(store)
	r := makeAuthRouter(ac)

	t.Run("defaults preferred_language to en when omitted", func(t *testing.T) {
		w := doJSONReq(t, r, http.MethodPost, "/api/register", map[string]any{
			"email":    "alice@soc.local",
			"password": "Password123!",
		}, nil)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var body map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		if body["preferred_language"] != "en" {
			t.Fatalf("expected preferred_language=en, got %v", body["preferred_language"])
		}
		saved, err := store.FindUserByEmail("alice@soc.local")
		if err != nil || saved.PreferredLanguage != "en" {
			t.Fatalf("expected saved user preferred_language=en, got %+v, err=%v", saved, err)
		}
	})

	t.Run("accepts preferred_language th", func(t *testing.T) {
		w := doJSONReq(t, r, http.MethodPost, "/api/register", map[string]any{
			"email":              "somchai@soc.local",
			"password":           "Password123!",
			"nickname":           "somchai_soc",
			"preferred_language": "th",
		}, nil)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var body map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		if body["preferred_language"] != "th" {
			t.Fatalf("expected preferred_language=th, got %v", body["preferred_language"])
		}
	})

	t.Run("rejects invalid preferred_language with 400 INVALID_LANGUAGE", func(t *testing.T) {
		w := doJSONReq(t, r, http.MethodPost, "/api/register", map[string]any{
			"email":              "badlang@soc.local",
			"password":           "Password123!",
			"preferred_language": "fr",
		}, nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
		var body map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		if body["code"] != "INVALID_LANGUAGE" || body["error"] == "" {
			t.Fatalf("expected INVALID_LANGUAGE, got %+v", body)
		}
	})

	t.Run("duplicate email emits AUTH_EMAIL_EXISTS and duplicate nickname emits AUTH_NICKNAME_TAKEN", func(t *testing.T) {
		wEmail := doJSONReq(t, r, http.MethodPost, "/api/register", map[string]any{
			"email":    "alice@soc.local",
			"password": "Password123!",
			"nickname": "unique_nick",
		}, nil)
		if wEmail.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", wEmail.Code)
		}
		var emailBody map[string]any
		_ = json.Unmarshal(wEmail.Body.Bytes(), &emailBody)
		if emailBody["code"] != "AUTH_EMAIL_EXISTS" {
			t.Fatalf("expected AUTH_EMAIL_EXISTS, got %+v", emailBody)
		}

		wNick := doJSONReq(t, r, http.MethodPost, "/api/register", map[string]any{
			"email":    "other@soc.local",
			"password": "Password123!",
			"nickname": "somchai_soc",
		}, nil)
		if wNick.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", wNick.Code)
		}
		var nickBody map[string]any
		_ = json.Unmarshal(wNick.Body.Bytes(), &nickBody)
		if nickBody["code"] != "AUTH_NICKNAME_TAKEN" {
			t.Fatalf("expected AUTH_NICKNAME_TAKEN, got %+v", nickBody)
		}
	})

	t.Run("invalid payload emits AUTH_INVALID_PAYLOAD", func(t *testing.T) {
		w := doJSONReq(t, r, http.MethodPost, "/api/register", "{bad-json", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
		var body map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		if body["code"] != "AUTH_INVALID_PAYLOAD" {
			t.Fatalf("expected AUTH_INVALID_PAYLOAD, got %+v", body)
		}
	})
}

func TestAuthController_LoginProfileNicknameAndUpdateLanguage(t *testing.T) {
	t.Setenv("JWT_SECRET", "auth-test-secret")
	store := newFakeAuthStore()
	pwHash, _ := bcrypt.GenerateFromPassword([]byte("Secret123!"), bcrypt.MinCost)
	store.seedUser(models.User{
		UserID:            "u-th-1",
		Email:             "thai@soc.local",
		Nickname:          "thai_analyst",
		PasswordHash:      string(pwHash),
		Role:              "learner",
		CurrentTier:       2,
		PreferredLanguage: "th",
		EmailVerified:     true,
		IsActive:          true,
	})
	store.seedUser(models.User{
		UserID:            "u-en-2",
		Email:             "eng@soc.local",
		Nickname:          "eng_analyst",
		PasswordHash:      string(pwHash),
		Role:              "learner",
		CurrentTier:       1,
		PreferredLanguage: "",
		EmailVerified:     true,
		IsActive:          true,
	})

	ac := NewAuthControllerWithStore(store)
	r := makeAuthRouter(ac)

	// Login tests
	wBadCreds := doJSONReq(t, r, http.MethodPost, "/api/login", map[string]any{
		"email":    "thai@soc.local",
		"password": "WrongPassword",
	}, nil)
	if wBadCreds.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", wBadCreds.Code)
	}
	var badBody map[string]any
	_ = json.Unmarshal(wBadCreds.Body.Bytes(), &badBody)
	if badBody["code"] != "AUTH_INVALID_CREDENTIALS" {
		t.Fatalf("expected AUTH_INVALID_CREDENTIALS, got %+v", badBody)
	}

	wLogin := doJSONReq(t, r, http.MethodPost, "/api/login", map[string]any{
		"email":    "thai@soc.local",
		"password": "Secret123!",
	}, nil)
	if wLogin.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", wLogin.Code, wLogin.Body.String())
	}
	var loginBody map[string]any
	_ = json.Unmarshal(wLogin.Body.Bytes(), &loginBody)
	if loginBody["preferred_language"] != "th" {
		t.Fatalf("expected preferred_language=th on login, got %+v", loginBody)
	}
	tokenStr, _ := loginBody["token"].(string)
	if tokenStr == "" {
		t.Fatalf("expected JWT token on login")
	}

	// GetProfile fallback "" -> "en"
	tokEn := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "u-en-2",
		"role":    "learner",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	signedEn, _ := tokEn.SignedString([]byte("auth-test-secret"))
	wProfEn := doJSONReq(t, r, http.MethodGet, "/api/user/profile", nil, map[string]string{
		"Authorization": "Bearer " + signedEn,
	})
	if wProfEn.Code != http.StatusOK {
		t.Fatalf("expected 200 profile, got %d", wProfEn.Code)
	}
	var profEn map[string]any
	_ = json.Unmarshal(wProfEn.Body.Bytes(), &profEn)
	if profEn["preferred_language"] != "en" {
		t.Fatalf("expected blank user preferred_language to default to en, got %+v", profEn)
	}

	// Missing user -> 404 USER_NOT_FOUND
	tokMissing := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "u-missing",
		"role":    "learner",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	signedMissing, _ := tokMissing.SignedString([]byte("auth-test-secret"))
	wMissing := doJSONReq(t, r, http.MethodGet, "/api/user/profile", nil, map[string]string{
		"Authorization": "Bearer " + signedMissing,
	})
	if wMissing.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", wMissing.Code)
	}
	var missingBody map[string]any
	_ = json.Unmarshal(wMissing.Body.Bytes(), &missingBody)
	if missingBody["code"] != "USER_NOT_FOUND" {
		t.Fatalf("expected USER_NOT_FOUND, got %+v", missingBody)
	}

	// UpdateNickname validation codes (post-trim length, empty, taken)
	authHeaders := map[string]string{"Authorization": "Bearer " + tokenStr}

	wNickEmpty := doJSONReq(t, r, http.MethodPut, "/api/user/nickname", map[string]any{"nickname": ""}, authHeaders)
	var nickEmptyBody map[string]any
	_ = json.Unmarshal(wNickEmpty.Body.Bytes(), &nickEmptyBody)
	if wNickEmpty.Code != http.StatusBadRequest || nickEmptyBody["code"] != "AUTH_NICKNAME_REQUIRED" {
		t.Fatalf("expected 400 AUTH_NICKNAME_REQUIRED, got %d %+v", wNickEmpty.Code, nickEmptyBody)
	}

	wNickTrimShort := doJSONReq(t, r, http.MethodPut, "/api/user/nickname", map[string]any{"nickname": " a "}, authHeaders)
	var nickTrimBody map[string]any
	_ = json.Unmarshal(wNickTrimShort.Body.Bytes(), &nickTrimBody)
	if wNickTrimShort.Code != http.StatusBadRequest || nickTrimBody["code"] != "AUTH_NICKNAME_INVALID" {
		t.Fatalf("expected 400 AUTH_NICKNAME_INVALID for ' a ', got %d %+v", wNickTrimShort.Code, nickTrimBody)
	}

	wNickTaken := doJSONReq(t, r, http.MethodPut, "/api/user/nickname", map[string]any{"nickname": "eng_analyst"}, authHeaders)
	var nickTakenBody map[string]any
	_ = json.Unmarshal(wNickTaken.Body.Bytes(), &nickTakenBody)
	if wNickTaken.Code != http.StatusBadRequest || nickTakenBody["code"] != "AUTH_NICKNAME_TAKEN" {
		t.Fatalf("expected 400 AUTH_NICKNAME_TAKEN, got %d %+v", wNickTaken.Code, nickTakenBody)
	}

	// UpdateLanguage endpoint
	wLangUnauth := doJSONReq(t, r, http.MethodPut, "/api/user/language", map[string]any{"language": "th"}, nil)
	var langUnauthBody map[string]any
	_ = json.Unmarshal(wLangUnauth.Body.Bytes(), &langUnauthBody)
	if wLangUnauth.Code != http.StatusUnauthorized || langUnauthBody["code"] != "AUTH_UNAUTHORIZED" {
		t.Fatalf("expected 401 AUTH_UNAUTHORIZED, got %d %+v", wLangUnauth.Code, langUnauthBody)
	}

	wLangBad := doJSONReq(t, r, http.MethodPut, "/api/user/language", map[string]any{"language": "jp"}, authHeaders)
	var langBadBody map[string]any
	_ = json.Unmarshal(wLangBad.Body.Bytes(), &langBadBody)
	if wLangBad.Code != http.StatusBadRequest || langBadBody["code"] != "INVALID_LANGUAGE" {
		t.Fatalf("expected 400 INVALID_LANGUAGE, got %d %+v", wLangBad.Code, langBadBody)
	}

	wLangEmpty := doJSONReq(t, r, http.MethodPut, "/api/user/language", map[string]any{"language": ""}, authHeaders)
	var langEmptyBody map[string]any
	_ = json.Unmarshal(wLangEmpty.Body.Bytes(), &langEmptyBody)
	if wLangEmpty.Code != http.StatusBadRequest || langEmptyBody["code"] != "INVALID_LANGUAGE" {
		t.Fatalf("expected 400 INVALID_LANGUAGE for empty language, got %d %+v", wLangEmpty.Code, langEmptyBody)
	}

	wLangMalformed := doJSONReq(t, r, http.MethodPut, "/api/user/language", "{bad", authHeaders)
	var langMalBody map[string]any
	_ = json.Unmarshal(wLangMalformed.Body.Bytes(), &langMalBody)
	if wLangMalformed.Code != http.StatusBadRequest || langMalBody["code"] != "AUTH_INVALID_PAYLOAD" {
		t.Fatalf("expected 400 AUTH_INVALID_PAYLOAD for malformed json, got %d %+v", wLangMalformed.Code, langMalBody)
	}

	wLangMissingUser := doJSONReq(t, r, http.MethodPut, "/api/user/language", map[string]any{"language": "th"}, map[string]string{
		"Authorization": "Bearer " + signedMissing,
	})
	var langMissingBody map[string]any
	_ = json.Unmarshal(wLangMissingUser.Body.Bytes(), &langMissingBody)
	if wLangMissingUser.Code != http.StatusNotFound || langMissingBody["code"] != "USER_NOT_FOUND" {
		t.Fatalf("expected 404 USER_NOT_FOUND, got %d %+v", wLangMissingUser.Code, langMissingBody)
	}

	wLangOK := doJSONReq(t, r, http.MethodPut, "/api/user/language", map[string]any{"language": "en"}, authHeaders)
	var langOKBody map[string]any
	_ = json.Unmarshal(wLangOK.Body.Bytes(), &langOKBody)
	if wLangOK.Code != http.StatusOK || langOKBody["message"] != "Language preference updated" || langOKBody["preferred_language"] != "en" {
		t.Fatalf("expected 200 updated to en, got %d %+v", wLangOK.Code, langOKBody)
	}

	// Verify persisted via GetProfile
	wProfUpdated := doJSONReq(t, r, http.MethodGet, "/api/user/profile", nil, authHeaders)
	var profUpdated map[string]any
	_ = json.Unmarshal(wProfUpdated.Body.Bytes(), &profUpdated)
	if profUpdated["preferred_language"] != "en" {
		t.Fatalf("expected profile preferred_language=en after update, got %+v", profUpdated)
	}
}

func TestGormAuthStore_UpdateUserLanguageSQL(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "host=localhost user=test dbname=test port=5432 sslmode=disable",
		PreferSimpleProtocol: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("gorm.Open DryRun failed: %v", err)
	}

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&models.User{}).Where("user_id = ?", "u-1").Update("preferred_language", "th")
	})

	if !strings.Contains(sql, `UPDATE "users"`) || !strings.Contains(sql, `"preferred_language"`) {
		t.Fatalf("expected UPDATE users SET preferred_language in generated SQL, got: %s", sql)
	}
}

type fakeMailRecord struct {
	Recipient string
	Nickname  string
	URL       string
}

type fakeMailSender struct {
	mu   sync.Mutex
	sent []fakeMailRecord
	ch   chan fakeMailRecord
}

func newFakeMailSender() *fakeMailSender {
	return &fakeMailSender{
		ch: make(chan fakeMailRecord, 16),
	}
}

func (f *fakeMailSender) SendVerificationEmail(_ context.Context, recipientEmail, nickname, verificationURL string) error {
	rec := fakeMailRecord{Recipient: recipientEmail, Nickname: nickname, URL: verificationURL}
	f.mu.Lock()
	f.sent = append(f.sent, rec)
	f.mu.Unlock()
	f.ch <- rec
	return nil
}

func (f *fakeMailSender) waitForMail(t *testing.T) fakeMailRecord {
	t.Helper()
	select {
	case rec := <-f.ch:
		return rec
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for async verification email dispatch")
		return fakeMailRecord{}
	}
}

func extractTokenFromVerifyURL(verifyURL string) string {
	const prefix = "token="
	idx := strings.Index(verifyURL, prefix)
	if idx == -1 {
		return ""
	}
	return verifyURL[idx+len(prefix):]
}

func TestAuthController_EmailVerificationLifecycle(t *testing.T) {
	t.Setenv("JWT_SECRET", "auth-verify-secret")
	t.Setenv("APP_BASE_URL", "https://soc.example.com")
	t.Setenv("REQUIRE_EMAIL_VERIFICATION", "true")

	store := newFakeAuthStore()
	mailer := newFakeMailSender()
	ac := NewAuthControllerWithDeps(store, mailer)
	r := makeAuthRouter(ac)

	// 1. Register creates unverified account (normalized lowercase) and dispatches dual-language verification email async
	wReg := doJSONReq(t, r, http.MethodPost, "/api/register", map[string]any{
		"email":    "NewLearner@SOC.Local",
		"password": "Password123!",
		"nickname": "new_learner",
	}, nil)
	if wReg.Code != http.StatusCreated {
		t.Fatalf("expected 201 register, got %d: %s", wReg.Code, wReg.Body.String())
	}
	var regBody map[string]any
	_ = json.Unmarshal(wReg.Body.Bytes(), &regBody)
	if regBody["requires_verification"] != true || regBody["email"] != "newlearner@soc.local" {
		t.Fatalf("expected requires_verification=true and canonical lowercase email, got %+v", regBody)
	}

	firstMail := mailer.waitForMail(t)
	firstToken := extractTokenFromVerifyURL(firstMail.URL)
	if !strings.HasPrefix(firstMail.URL, "https://soc.example.com/verify-email?token=") || firstToken == "" {
		t.Fatalf("unexpected verification link: %q", firstMail.URL)
	}

	// 2. Wrong password returns 401 AUTH_INVALID_CREDENTIALS even when unverified (no status leak)
	wWrongPw := doJSONReq(t, r, http.MethodPost, "/api/login", map[string]any{
		"email":    "NEWLEARNER@SOC.LOCAL",
		"password": "WrongPassword!",
	}, nil)
	if wWrongPw.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on wrong password for unverified user, got %d: %s", wWrongPw.Code, wWrongPw.Body.String())
	}

	// 3. Valid password while unverified returns 403 AUTH_EMAIL_NOT_VERIFIED (case-insensitive)
	wUnverifiedLogin := doJSONReq(t, r, http.MethodPost, "/api/login", map[string]any{
		"email":    "NEWLEARNER@SOC.LOCAL",
		"password": "Password123!",
	}, nil)
	var unverifiedBody map[string]any
	_ = json.Unmarshal(wUnverifiedLogin.Body.Bytes(), &unverifiedBody)
	if wUnverifiedLogin.Code != http.StatusForbidden || unverifiedBody["code"] != "AUTH_EMAIL_NOT_VERIFIED" {
		t.Fatalf("expected 403 AUTH_EMAIL_NOT_VERIFIED, got %d %+v", wUnverifiedLogin.Code, unverifiedBody)
	}

	// 4. Resend within 60s returns 429 AUTH_VERIFICATION_COOLDOWN
	wCooldown := doJSONReq(t, r, http.MethodPost, "/api/resend-verification", map[string]any{
		"email": "NewLearner@soc.local",
	}, nil)
	var cooldownBody map[string]any
	_ = json.Unmarshal(wCooldown.Body.Bytes(), &cooldownBody)
	if wCooldown.Code != http.StatusTooManyRequests || cooldownBody["code"] != "AUTH_VERIFICATION_COOLDOWN" {
		t.Fatalf("expected 429 AUTH_VERIFICATION_COOLDOWN, got %d %+v", wCooldown.Code, cooldownBody)
	}

	// 5. After 60s elapsed, resend rotates token hash and invalidates first token
	savedUser, _ := store.FindUserByEmail("newlearner@soc.local")
	pastSent := time.Now().UTC().Add(-65 * time.Second)
	store.usersByID[savedUser.UserID].VerificationSentAt = &pastSent

	wResend := doJSONReq(t, r, http.MethodPost, "/api/resend-verification", map[string]any{
		"email": "newlearner@soc.local",
	}, nil)
	if wResend.Code != http.StatusOK {
		t.Fatalf("expected 200 on resend after cooldown, got %d: %s", wResend.Code, wResend.Body.String())
	}
	secondMail := mailer.waitForMail(t)
	secondToken := extractTokenFromVerifyURL(secondMail.URL)
	if secondToken == "" || secondToken == firstToken {
		t.Fatalf("expected fresh rotated verification token")
	}

	// Old token must fail
	wOldTok := doJSONReq(t, r, http.MethodPost, "/api/verify-email", map[string]any{"token": firstToken}, nil)
	if wOldTok.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalidated prior token, got %d", wOldTok.Code)
	}

	// 6. Unknown email on resend returns neutral 200 (no enumeration)
	wUnknownResend := doJSONReq(t, r, http.MethodPost, "/api/resend-verification", map[string]any{
		"email": "nobody@soc.local",
	}, nil)
	if wUnknownResend.Code != http.StatusOK {
		t.Fatalf("expected neutral 200 on unknown resend email, got %d", wUnknownResend.Code)
	}

	// 7. Verify latest valid token -> 200, single-use replay -> 400, login -> 200 JWT
	wVerify := doJSONReq(t, r, http.MethodPost, "/api/verify-email", map[string]any{"token": secondToken}, nil)
	if wVerify.Code != http.StatusOK {
		t.Fatalf("expected 200 verify-email, got %d: %s", wVerify.Code, wVerify.Body.String())
	}
	wReplay := doJSONReq(t, r, http.MethodPost, "/api/verify-email", map[string]any{"token": secondToken}, nil)
	if wReplay.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on single-use replay, got %d", wReplay.Code)
	}

	wLoginOK := doJSONReq(t, r, http.MethodPost, "/api/login", map[string]any{
		"email":    "newlearner@soc.local",
		"password": "Password123!",
	}, nil)
	if wLoginOK.Code != http.StatusOK {
		t.Fatalf("expected 200 login after verification, got %d: %s", wLoginOK.Code, wLoginOK.Body.String())
	}
}

func TestAuthController_ExpiredTokenAndEnvBypass(t *testing.T) {
	store := newFakeAuthStore()
	mailer := newFakeMailSender()
	ac := NewAuthControllerWithDeps(store, mailer)
	r := makeAuthRouter(ac)

	t.Run("expired_token_returns_400_expired", func(t *testing.T) {
		t.Setenv("REQUIRE_EMAIL_VERIFICATION", "true")
		_ = doJSONReq(t, r, http.MethodPost, "/api/register", map[string]any{
			"email":    "expired@soc.local",
			"password": "Password123!",
			"nickname": "expired_user",
		}, nil)
		mail := mailer.waitForMail(t)
		tok := extractTokenFromVerifyURL(mail.URL)
		u, _ := store.FindUserByEmail("expired@soc.local")
		expiredAt := time.Now().UTC().Add(-time.Minute)
		store.usersByID[u.UserID].VerificationExpiresAt = &expiredAt

		wExp := doJSONReq(t, r, http.MethodPost, "/api/verify-email", map[string]any{"token": tok}, nil)
		var expBody map[string]any
		_ = json.Unmarshal(wExp.Body.Bytes(), &expBody)
		if wExp.Code != http.StatusBadRequest || expBody["code"] != "AUTH_VERIFICATION_TOKEN_EXPIRED" {
			t.Fatalf("expected 400 AUTH_VERIFICATION_TOKEN_EXPIRED, got %d %+v", wExp.Code, expBody)
		}
	})

	t.Run("bypass_flag_auto_verifies_on_register", func(t *testing.T) {
		t.Setenv("REQUIRE_EMAIL_VERIFICATION", "false")
		wReg := doJSONReq(t, r, http.MethodPost, "/api/register", map[string]any{
			"email":    "bypass@soc.local",
			"password": "Password123!",
			"nickname": "bypass_user",
		}, nil)
		var body map[string]any
		_ = json.Unmarshal(wReg.Body.Bytes(), &body)
		if wReg.Code != http.StatusCreated || body["requires_verification"] != false {
			t.Fatalf("expected 201 requires_verification=false, got %d %+v", wReg.Code, body)
		}
		wLogin := doJSONReq(t, r, http.MethodPost, "/api/login", map[string]any{
			"email":    "bypass@soc.local",
			"password": "Password123!",
		}, nil)
		if wLogin.Code != http.StatusOK {
			t.Fatalf("expected immediate 200 login when REQUIRE_EMAIL_VERIFICATION=false, got %d", wLogin.Code)
		}
	})

	t.Run("nil_auth_store_returns_500_cleanly", func(t *testing.T) {
		acNil := &AuthController{}
		rNil := makeAuthRouter(acNil)
		w := doJSONReq(t, rNil, http.MethodPost, "/api/login", map[string]any{
			"email":    "any@soc.local",
			"password": "Password123!",
		}, nil)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 for nil auth store, got %d", w.Code)
		}
	})
}

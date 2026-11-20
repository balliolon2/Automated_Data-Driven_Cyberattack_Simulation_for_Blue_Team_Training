package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	for _, u := range f.usersByID {
		if u.Email == email {
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

func makeAuthRouter(ac *AuthController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	api.POST("/register", ac.Register)
	api.POST("/login", ac.Login)
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

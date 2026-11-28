package controllers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("supersecretkey") // Default for development
	}
	return []byte(secret)
}

func isSupportedLanguage(lang string) bool {
	return lang == "en" || lang == "th"
}

func normalizeUserLanguage(lang string) string {
	if lang == "th" {
		return "th"
	}
	return "en"
}

// AuthStore abstracts persistence for AuthController handlers so the HTTP
// layer can be tested cleanly without a live PostgreSQL instance.
type AuthStore interface {
	FindUserByEmail(email string) (*models.User, error)
	FindUserByNickname(nickname string) (*models.User, error)
	FindUserByID(userID string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateLastLogin(userID string, t time.Time) error
	GetLatestApplicationByUserID(userID string) (*models.SpecialistApplication, error)
	IsNicknameTakenByOther(nickname, excludeUserID string) (bool, error)
	UpdateUserNickname(userID, nickname string) error
	UpdateUserLanguage(userID, lang string) (int64, error)
	FindUserByVerificationTokenHash(tokenHash string) (*models.User, error)
	MarkUserEmailVerified(userID string, verifiedAt time.Time) error
	UpdateVerificationToken(userID, tokenHash string, sentAt, expiresAt time.Time) error
}

type gormAuthStore struct {
	DB *gorm.DB
}

func (s *gormAuthStore) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.DB.Where("LOWER(email) = LOWER(?)", strings.TrimSpace(email)).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *gormAuthStore) FindUserByNickname(nickname string) (*models.User, error) {
	var user models.User
	if err := s.DB.Where("nickname = ?", nickname).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *gormAuthStore) FindUserByID(userID string) (*models.User, error) {
	var user models.User
	if err := s.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *gormAuthStore) CreateUser(user *models.User) error {
	return s.DB.Create(user).Error
}

func (s *gormAuthStore) UpdateLastLogin(userID string, t time.Time) error {
	return s.DB.Model(&models.User{}).Where("user_id = ?", userID).Update("last_login", t).Error
}

func (s *gormAuthStore) GetLatestApplicationByUserID(userID string) (*models.SpecialistApplication, error) {
	var app models.SpecialistApplication
	if err := s.DB.Where("user_id = ?", userID).Order("created_at desc").First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (s *gormAuthStore) IsNicknameTakenByOther(nickname, excludeUserID string) (bool, error) {
	var count int64
	if err := s.DB.Model(&models.User{}).Where("nickname = ? AND user_id != ?", nickname, excludeUserID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *gormAuthStore) UpdateUserNickname(userID, nickname string) error {
	return s.DB.Model(&models.User{}).Where("user_id = ?", userID).Update("nickname", nickname).Error
}

func (s *gormAuthStore) UpdateUserLanguage(userID, lang string) (int64, error) {
	res := s.DB.Model(&models.User{}).Where("user_id = ?", userID).Update("preferred_language", lang)
	return res.RowsAffected, res.Error
}

func (s *gormAuthStore) FindUserByVerificationTokenHash(tokenHash string) (*models.User, error) {
	var user models.User
	if err := s.DB.Where("verification_token_hash = ?", tokenHash).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *gormAuthStore) MarkUserEmailVerified(userID string, verifiedAt time.Time) error {
	return s.DB.Model(&models.User{}).Where("user_id = ?", userID).Updates(map[string]any{
		"email_verified":          true,
		"email_verified_at":       verifiedAt,
		"verification_token_hash": nil,
		"verification_sent_at":    nil,
		"verification_expires_at": nil,
	}).Error
}

func (s *gormAuthStore) UpdateVerificationToken(userID, tokenHash string, sentAt, expiresAt time.Time) error {
	return s.DB.Model(&models.User{}).Where("user_id = ?", userID).Updates(map[string]any{
		"verification_token_hash": tokenHash,
		"verification_sent_at":    sentAt,
		"verification_expires_at": expiresAt,
	}).Error
}

type AuthController struct {
	DB         *gorm.DB
	store      AuthStore
	mailSender services.MailSender
}

func (ac *AuthController) effectiveStore() AuthStore {
	if ac.store != nil {
		return ac.store
	}
	if ac.DB != nil {
		return &gormAuthStore{DB: ac.DB}
	}
	return nil
}

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{
		DB:         db,
		store:      &gormAuthStore{DB: db},
		mailSender: services.NewSMTPMailSenderFromEnv(),
	}
}

func NewAuthControllerWithStore(store AuthStore) *AuthController {
	return &AuthController{store: store}
}

func NewAuthControllerWithDeps(store AuthStore, mailSender services.MailSender) *AuthController {
	return &AuthController{
		store:      store,
		mailSender: mailSender,
	}
}

// initUserVerification assigns verification state and issues a fresh CSPRNG token when required.
func initUserVerification(user *models.User, requireVerification bool) (rawToken string, err error) {
	if !requireVerification {
		now := time.Now().UTC()
		user.EmailVerified = true
		user.EmailVerifiedAt = &now
		user.VerificationTokenHash = nil
		return "", nil
	}
	raw, hash, expiresAt, err := services.GenerateVerificationToken()
	if err != nil {
		return "", err
	}
	sentAt := time.Now().UTC()
	user.EmailVerified = false
	user.EmailVerifiedAt = nil
	user.VerificationTokenHash = &hash
	user.VerificationSentAt = &sentAt
	user.VerificationExpiresAt = &expiresAt
	return raw, nil
}

type userAvailabilityLookup interface {
	FindUserByEmail(email string) (*models.User, error)
	FindUserByNickname(nickname string) (*models.User, error)
}

type preparedRegistration struct {
	User              models.User
	RawToken          string
	RequireVerify     bool
	NormalizedLang    string
	ValidationErrCode string
	ValidationErrMsg  string
	ValidationStatus  int
}

func validateAndPrepareNewUser(
	lookup userAvailabilityLookup,
	rawEmail, password, rawNickname, rawPreferredLang string,
) (preparedRegistration, error) {
	preferredLang := strings.TrimSpace(rawPreferredLang)
	if preferredLang == "" {
		preferredLang = "en"
	} else if !isSupportedLanguage(preferredLang) {
		return preparedRegistration{
			ValidationStatus:  http.StatusBadRequest,
			ValidationErrCode: "INVALID_LANGUAGE",
			ValidationErrMsg:  "Unsupported language. Supported values: en, th",
		}, nil
	}

	email := strings.ToLower(strings.TrimSpace(rawEmail))
	if email == "" || len(password) < 6 {
		return preparedRegistration{
			ValidationStatus:  http.StatusBadRequest,
			ValidationErrCode: "AUTH_INVALID_PAYLOAD",
			ValidationErrMsg:  "Valid email and password (minimum 6 characters) are required",
		}, nil
	}
	emailValidate := validator.New()
	if err := emailValidate.Var(email, "required,email"); err != nil {
		return preparedRegistration{
			ValidationStatus:  http.StatusBadRequest,
			ValidationErrCode: "AUTH_INVALID_PAYLOAD",
			ValidationErrMsg:  "Invalid email address format",
		}, nil
	}
	if parsed, err := mail.ParseAddress(email); err != nil || parsed == nil || parsed.Address != email {
		return preparedRegistration{
			ValidationStatus:  http.StatusBadRequest,
			ValidationErrCode: "AUTH_INVALID_PAYLOAD",
			ValidationErrMsg:  "Invalid email address format",
		}, nil
	}

	nickname := strings.TrimSpace(rawNickname)
	if nickname == "" {
		parts := strings.Split(email, "@")
		if len(parts) > 0 && parts[0] != "" {
			nickname = parts[0]
		} else {
			return preparedRegistration{
				ValidationStatus:  http.StatusBadRequest,
				ValidationErrCode: "AUTH_NICKNAME_REQUIRED",
				ValidationErrMsg:  "Nickname is required",
			}, nil
		}
	}
	if nickLen := len([]rune(nickname)); strings.TrimSpace(rawNickname) != "" && (nickLen < 2 || nickLen > 30) {
		return preparedRegistration{
			ValidationStatus:  http.StatusBadRequest,
			ValidationErrCode: "AUTH_NICKNAME_INVALID",
			ValidationErrMsg:  "Nickname must be between 2 and 30 characters",
		}, nil
	}

	if _, err := lookup.FindUserByEmail(email); err == nil {
		return preparedRegistration{
			ValidationStatus:  http.StatusBadRequest,
			ValidationErrCode: "AUTH_EMAIL_EXISTS",
			ValidationErrMsg:  "Email is already registered",
		}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return preparedRegistration{}, fmt.Errorf("could not verify email availability: %w", err)
	}

	if _, err := lookup.FindUserByNickname(nickname); err == nil {
		return preparedRegistration{
			ValidationStatus:  http.StatusBadRequest,
			ValidationErrCode: "AUTH_NICKNAME_TAKEN",
			ValidationErrMsg:  "Nickname is already taken",
		}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return preparedRegistration{}, fmt.Errorf("could not verify nickname availability: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return preparedRegistration{}, fmt.Errorf("could not hash password: %w", err)
	}

	requireVerify := services.RequireEmailVerification()
	user := models.User{
		Email:             email,
		Nickname:          nickname,
		PasswordHash:      string(hashedPassword),
		Role:              "learner",
		PreferredLanguage: preferredLang,
	}
	rawToken, err := initUserVerification(&user, requireVerify)
	if err != nil {
		return preparedRegistration{}, fmt.Errorf("could not initialize email verification: %w", err)
	}

	return preparedRegistration{
		User:           user,
		RawToken:       rawToken,
		RequireVerify:  requireVerify,
		NormalizedLang: normalizeUserLanguage(user.PreferredLanguage),
	}, nil
}

type RegisterInput = dto.RegisterRequest

// Register godoc
// @Summary Register a new learner
// @Description Creates a new learner account with email, password, nickname, and optional preferred_language
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User Registration Info"
// @Success 201 {object} dto.RegisterResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /register [post]
func (ac *AuthController) Register(c *gin.Context) {
	store := ac.effectiveStore()
	if store == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth store is unavailable", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "AUTH_INVALID_PAYLOAD"})
		return
	}

	prep, err := validateAndPrepareNewUser(store, input.Email, input.Password, input.Nickname, input.PreferredLanguage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": "INTERNAL_SERVER_ERROR"})
		return
	}
	if prep.ValidationErrCode != "" {
		c.JSON(prep.ValidationStatus, gin.H{"error": prep.ValidationErrMsg, "code": prep.ValidationErrCode})
		return
	}

	user := prep.User
	if err := store.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user: " + err.Error(), "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	if prep.RequireVerify && prep.RawToken != "" {
		services.DispatchVerificationEmailAsync(services.EffectiveMailSender(ac.mailSender), user.Email, user.Nickname, prep.RawToken)
	}

	msg := "User registered successfully"
	if prep.RequireVerify {
		msg = "User registered. Please check your email to verify your account."
	}
	c.JSON(http.StatusCreated, dto.RegisterResponse{
		Message:              msg,
		Nickname:             user.Nickname,
		Email:                user.Email,
		PreferredLanguage:    prep.NormalizedLang,
		RequiresVerification: prep.RequireVerify,
	})
}

type LoginInput = dto.LoginRequest

// Login godoc
// @Summary Login user
// @Description Authenticates a user and returns a signed JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Credentials"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /login [post]
func (ac *AuthController) Login(c *gin.Context) {
	store := ac.effectiveStore()
	if store == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth store is unavailable", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "AUTH_INVALID_PAYLOAD"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := store.FindUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password", "code": "AUTH_INVALID_CREDENTIALS"})
		return
	}

	// Always validate password before disclosing pending verification status.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password", "code": "AUTH_INVALID_CREDENTIALS"})
		return
	}

	if services.RequireEmailVerification() && !user.EmailVerified {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Your account is pending email verification. Please verify your email before logging in.",
			"code":  "AUTH_EMAIL_NOT_VERIFIED",
			"email": user.Email,
		})
		return
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	_ = store.UpdateLastLogin(user.UserID, now)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.UserID,
		"role":     user.Role,
		"nickname": user.Nickname,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		Token:             tokenString,
		Email:             user.Email,
		Nickname:          user.Nickname,
		Role:              user.Role,
		UserID:            user.UserID,
		PreferredLanguage: normalizeUserLanguage(user.PreferredLanguage),
	})
}

// VerifyEmail godoc
// @Summary Verify user email address
// @Description Confirms single-use SHA-256 hashed verification token issued on registration
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.VerifyEmailRequest true "Verification Token Payload"
// @Success 200 {object} dto.VerifyEmailResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /verify-email [post]
func (ac *AuthController) VerifyEmail(c *gin.Context) {
	store := ac.effectiveStore()
	if store == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth store is unavailable", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	var input dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.Token) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification token is required", "code": "AUTH_INVALID_PAYLOAD"})
		return
	}

	tokenHash := services.HashVerificationToken(input.Token)
	user, err := store.FindUserByVerificationTokenHash(tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or already used verification token", "code": "AUTH_VERIFICATION_TOKEN_INVALID"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate verification token", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	if user.VerificationExpiresAt == nil || time.Now().UTC().After(*user.VerificationExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification link has expired. Please request a new verification email.", "code": "AUTH_VERIFICATION_TOKEN_EXPIRED"})
		return
	}

	if err := store.MarkUserEmailVerified(user.UserID, time.Now().UTC()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not mark email as verified", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, dto.VerifyEmailResponse{
		Message: "Email verified successfully",
		Email:   user.Email,
	})
}

// ResendVerification godoc
// @Summary Resend account verification email
// @Description Rotates verification token and sends a new dual-language verification email (60s cooldown)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ResendVerificationRequest true "Target Email Payload"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 429 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /resend-verification [post]
func (ac *AuthController) ResendVerification(c *gin.Context) {
	store := ac.effectiveStore()
	if store == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth store is unavailable", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	var input dto.ResendVerificationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "AUTH_INVALID_PAYLOAD"})
		return
	}

	const neutralMsg = "If the account exists and requires verification, a verification email has been sent."
	if !services.RequireEmailVerification() {
		c.JSON(http.StatusOK, gin.H{"message": neutralMsg})
		return
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := store.FindUserByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"message": neutralMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process verification request", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	if user.EmailVerified {
		c.JSON(http.StatusOK, gin.H{"message": neutralMsg})
		return
	}

	if user.VerificationSentAt != nil {
		elapsed := time.Since(*user.VerificationSentAt)
		if elapsed < services.VerificationResendCooldown {
			remaining := int(math.Ceil((services.VerificationResendCooldown - elapsed).Seconds()))
			if remaining < 1 {
				remaining = 1
			}
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":               fmt.Sprintf("Please wait %d seconds before requesting another verification email.", remaining),
				"code":                "AUTH_VERIFICATION_COOLDOWN",
				"retry_after_seconds": remaining,
				"params":              gin.H{"seconds": remaining},
			})
			return
		}
	}

	rawToken, tokenHash, expiresAt, err := services.GenerateVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate verification token", "code": "INTERNAL_SERVER_ERROR"})
		return
	}
	sentAt := time.Now().UTC()
	if err := store.UpdateVerificationToken(user.UserID, tokenHash, sentAt, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update verification token", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	services.DispatchVerificationEmailAsync(services.EffectiveMailSender(ac.mailSender), user.Email, user.Nickname, rawToken)
	c.JSON(http.StatusOK, gin.H{"message": neutralMsg})
}

// GetProfile returns the authenticated user's profile and latest specialist application status
func (ac *AuthController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}

	store := ac.effectiveStore()
	if store == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth store is unavailable", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	user, err := store.FindUserByID(userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "code": "USER_NOT_FOUND"})
		return
	}

	// Find latest application if any
	var appDTO *dto.SpecialistApplicationDTO
	if app, err := store.GetLatestApplicationByUserID(userIDStr); err == nil && app != nil {
		dtoObj := app.ToDTO(user.Nickname, user.Email)
		appDTO = &dtoObj
	}

	c.JSON(http.StatusOK, dto.UserProfileResponse{
		UserID:            user.UserID,
		Email:             user.Email,
		Nickname:          user.Nickname,
		Role:              user.Role,
		CurrentTier:       user.CurrentTier,
		PreferredLanguage: normalizeUserLanguage(user.PreferredLanguage),
		EmailVerified:     user.EmailVerified,
		LatestApplication: appDTO,
	})
}

// UpdateNickname updates the current user's nickname
func (ac *AuthController) UpdateNickname(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}

	store := ac.effectiveStore()
	if store == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth store is unavailable", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	var input dto.UpdateNicknameRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		code := "AUTH_INVALID_PAYLOAD"
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			code = "AUTH_NICKNAME_INVALID"
			for _, fe := range ve {
				if fe.Tag() == "required" {
					code = "AUTH_NICKNAME_REQUIRED"
					break
				}
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": code})
		return
	}

	newNickname := strings.TrimSpace(input.Nickname)
	if newNickname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname cannot be empty", "code": "AUTH_NICKNAME_REQUIRED"})
		return
	}
	trimmedLen := len([]rune(newNickname))
	if trimmedLen < 2 || trimmedLen > 30 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname must be between 2 and 30 characters", "code": "AUTH_NICKNAME_INVALID"})
		return
	}

	// Check if nickname taken by someone else
	taken, err := store.IsNicknameTakenByOther(newNickname, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify nickname availability", "code": "INTERNAL_SERVER_ERROR"})
		return
	}
	if taken {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname is already taken", "code": "AUTH_NICKNAME_TAKEN"})
		return
	}

	if err := store.UpdateUserNickname(userIDStr, newNickname); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update nickname", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Nickname updated successfully",
		"nickname": newNickname,
	})
}

// UpdateLanguage godoc
// @Summary Update user language preference
// @Description Updates the authenticated user's preferred UI language ('en' or 'th')
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateLanguageRequest true "Language Preference"
// @Success 200 {object} dto.UpdateLanguageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/language [put]
func (ac *AuthController) UpdateLanguage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "code": "AUTH_UNAUTHORIZED"})
		return
	}

	store := ac.effectiveStore()
	if store == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth store is unavailable", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	var input dto.UpdateLanguageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported language. Supported values: en, th", "code": "INVALID_LANGUAGE"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "AUTH_INVALID_PAYLOAD"})
		return
	}

	lang := strings.TrimSpace(input.Language)
	if !isSupportedLanguage(lang) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported language. Supported values: en, th", "code": "INVALID_LANGUAGE"})
		return
	}

	rowsAffected, err := store.UpdateUserLanguage(userIDStr, lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update language preference", "code": "INTERNAL_SERVER_ERROR"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "code": "USER_NOT_FOUND"})
		return
	}

	c.JSON(http.StatusOK, dto.UpdateLanguageResponse{
		Message:           "Language preference updated",
		PreferredLanguage: lang,
	})
}

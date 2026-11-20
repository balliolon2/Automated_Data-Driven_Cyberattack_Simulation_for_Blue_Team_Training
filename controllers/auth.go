package controllers

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"

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
}

type gormAuthStore struct {
	DB *gorm.DB
}

func (s *gormAuthStore) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
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

type AuthController struct {
	DB    *gorm.DB
	store AuthStore
}

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{
		DB:    db,
		store: &gormAuthStore{DB: db},
	}
}

func NewAuthControllerWithStore(store AuthStore) *AuthController {
	return &AuthController{store: store}
}

type RegisterInput = dto.RegisterRequest

// Register godoc
// @Summary Register a new learner
// @Description Creates a new learner account with email, password, nickname, and optional preferred_language
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User Registration Info"
// @Success 201 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "AUTH_INVALID_PAYLOAD"})
		return
	}

	preferredLang := strings.TrimSpace(input.PreferredLanguage)
	if preferredLang == "" {
		preferredLang = "en"
	} else if !isSupportedLanguage(preferredLang) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported language. Supported values: en, th", "code": "INVALID_LANGUAGE"})
		return
	}

	nickname := strings.TrimSpace(input.Nickname)
	if nickname == "" {
		parts := strings.Split(input.Email, "@")
		if len(parts) > 0 && parts[0] != "" {
			nickname = parts[0]
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname is required", "code": "AUTH_NICKNAME_REQUIRED"})
			return
		}
	}

	// Check if email already exists
	if _, err := ac.store.FindUserByEmail(input.Email); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already registered", "code": "AUTH_EMAIL_EXISTS"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify email availability", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	// Check if nickname already exists
	if _, err := ac.store.FindUserByNickname(nickname); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname is already taken", "code": "AUTH_NICKNAME_TAKEN"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify nickname availability", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password", "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	user := models.User{
		Email:             input.Email,
		Nickname:          nickname,
		PasswordHash:      string(hashedPassword),
		Role:              "learner",
		PreferredLanguage: preferredLang,
	}

	if err := ac.store.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user: " + err.Error(), "code": "INTERNAL_SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":            "User registered successfully",
		"nickname":           user.Nickname,
		"email":              user.Email,
		"preferred_language": normalizeUserLanguage(user.PreferredLanguage),
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
// @Failure 500 {object} dto.ErrorResponse
// @Router /login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "AUTH_INVALID_PAYLOAD"})
		return
	}

	user, err := ac.store.FindUserByEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password", "code": "AUTH_INVALID_CREDENTIALS"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password", "code": "AUTH_INVALID_CREDENTIALS"})
		return
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	_ = ac.store.UpdateLastLogin(user.UserID, now)

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

	user, err := ac.store.FindUserByID(userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "code": "USER_NOT_FOUND"})
		return
	}

	// Find latest application if any
	var appDTO *dto.SpecialistApplicationDTO
	if app, err := ac.store.GetLatestApplicationByUserID(userIDStr); err == nil && app != nil {
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
	taken, err := ac.store.IsNicknameTakenByOther(newNickname, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify nickname availability", "code": "INTERNAL_SERVER_ERROR"})
		return
	}
	if taken {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname is already taken", "code": "AUTH_NICKNAME_TAKEN"})
		return
	}

	if err := ac.store.UpdateUserNickname(userIDStr, newNickname); err != nil {
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

	rowsAffected, err := ac.store.UpdateUserLanguage(userIDStr, lang)
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

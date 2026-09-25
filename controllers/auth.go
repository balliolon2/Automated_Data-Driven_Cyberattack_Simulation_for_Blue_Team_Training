package controllers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"

	"github.com/gin-gonic/gin"
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

type AuthController struct {
	DB *gorm.DB
}

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{DB: db}
}

type RegisterInput = dto.RegisterRequest

// Register godoc
// @Summary Register a new learner
// @Description Creates a new learner account with email, password, and nickname
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nickname := strings.TrimSpace(input.Nickname)
	if nickname == "" {
		parts := strings.Split(input.Email, "@")
		if len(parts) > 0 && parts[0] != "" {
			nickname = parts[0]
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname is required"})
			return
		}
	}

	// Check if email already exists
	var existingUser models.User
	if err := ac.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already registered"})
		return
	}

	// Check if nickname already exists
	if err := ac.DB.Where("nickname = ?", nickname).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname is already taken"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	user := models.User{
		Email:        input.Email,
		Nickname:     nickname,
		PasswordHash: string(hashedPassword),
		Role:         "learner",
	}

	if err := ac.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "User registered successfully",
		"nickname": user.Nickname,
		"email":    user.Email,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := ac.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	ac.DB.Model(&user).Update("last_login", now)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.UserID,
		"role":     user.Role,
		"nickname": user.Nickname,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		Token:    tokenString,
		Email:    user.Email,
		Nickname: user.Nickname,
		Role:     user.Role,
		UserID:   user.UserID,
	})
}

// GetProfile returns the authenticated user's profile and latest specialist application status
func (ac *AuthController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user models.User
	if err := ac.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Find latest application if any
	var app models.SpecialistApplication
	var appDTO *dto.SpecialistApplicationDTO
	if err := ac.DB.Where("user_id = ?", userID).Order("created_at desc").First(&app).Error; err == nil {
		dtoObj := app.ToDTO(user.Nickname, user.Email)
		appDTO = &dtoObj
	}

	c.JSON(http.StatusOK, dto.UserProfileResponse{
		UserID:            user.UserID,
		Email:             user.Email,
		Nickname:          user.Nickname,
		Role:              user.Role,
		CurrentTier:       user.CurrentTier,
		LatestApplication: appDTO,
	})
}

// UpdateNickname updates the current user's nickname
func (ac *AuthController) UpdateNickname(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input dto.UpdateNicknameRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newNickname := strings.TrimSpace(input.Nickname)
	if newNickname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname cannot be empty"})
		return
	}

	// Check if nickname taken by someone else
	var count int64
	ac.DB.Model(&models.User{}).Where("nickname = ? AND user_id != ?", newNickname, userID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nickname is already taken"})
		return
	}

	if err := ac.DB.Model(&models.User{}).Where("user_id = ?", userID).Update("nickname", newNickname).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update nickname"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Nickname updated successfully",
		"nickname": newNickname,
	})
}

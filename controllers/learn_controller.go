package controllers

import (
	"math"
	"net/http"
	"time"

	"cybersim/dto"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LearnStore abstracts persistence for the Learning Theory Hub handlers so the
// HTTP layer can be tested cleanly without a live PostgreSQL instance.
type LearnStore interface {
	GetUserSkillProfiles(userID string) ([]models.UserSkillProfile, error)
	GetTheoryProgress(userID string) ([]models.UserTheoryProgress, error)
	MarkTheoryCompleted(userID, domainID string) error
	LoadModules() ([]services.TheoryModuleContent, error)
}

type gormLearnStore struct {
	DB *gorm.DB
}

func (s *gormLearnStore) GetUserSkillProfiles(userID string) ([]models.UserSkillProfile, error) {
	var profiles []models.UserSkillProfile
	if err := s.DB.Where("user_id = ?", userID).Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

func (s *gormLearnStore) GetTheoryProgress(userID string) ([]models.UserTheoryProgress, error) {
	var progress []models.UserTheoryProgress
	if err := s.DB.Where("user_id = ?", userID).Find(&progress).Error; err != nil {
		return nil, err
	}
	return progress, nil
}

func (s *gormLearnStore) MarkTheoryCompleted(userID, domainID string) error {
	row := models.UserTheoryProgress{
		UserID:      userID,
		DomainID:    domainID,
		CompletedAt: time.Now(),
	}
	return s.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "domain_id"}},
		DoNothing: true,
	}).Create(&row).Error
}

func (s *gormLearnStore) LoadModules() ([]services.TheoryModuleContent, error) {
	return services.LoadTheoryModules(s.DB)
}

type LearnController struct {
	Store LearnStore
}

func NewLearnController(db *gorm.DB) *LearnController {
	return &LearnController{
		Store: &gormLearnStore{DB: db},
	}
}

func NewLearnControllerWithStore(store LearnStore) *LearnController {
	return &LearnController{Store: store}
}

func buildOrderedCompletedDomains(progress []models.UserTheoryProgress) ([]string, map[string]models.UserTheoryProgress) {
	byDomain := make(map[string]models.UserTheoryProgress, len(progress))
	for _, item := range progress {
		byDomain[item.DomainID] = item
	}
	ordered := make([]string, 0, len(services.OrderedTheoryDomains))
	for _, spec := range services.OrderedTheoryDomains {
		if _, ok := byDomain[spec.DomainID]; ok {
			ordered = append(ordered, spec.DomainID)
		}
	}
	return ordered, byDomain
}

func (lc *LearnController) requireStage1Access(c *gin.Context) (string, []models.UserSkillProfile, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return "", nil, false
	}
	userID, _ := userIDVal.(string)
	role := c.GetString("role")

	profiles, err := lc.Store.GetUserSkillProfiles(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify pre-test status"})
		return "", nil, false
	}

	if !services.CanAccessTheoryHub(role, len(profiles)) {
		unlocked := false
		c.JSON(http.StatusForbidden, dto.TheoryGateErrorResponse{
			Error:    "Please complete the Pre-Test first to unlock Learning Theory.",
			Reason:   "pre_test_required",
			Unlocked: &unlocked,
		})
		return "", nil, false
	}

	return userID, profiles, true
}

// GetDomains godoc
// @Summary List all five Security+ Theory Modules
// @Description Returns the 5 CompTIA Security+ domain Theory Modules, learner proficiency context, and completion state. Learners must finish the Pre-Test first; Specialists and Admins always have access.
// @Tags Learn
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.LearnDomainsResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.TheoryGateErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /learn/domains [get]
func (lc *LearnController) GetDomains(c *gin.Context) {
	userID, profiles, ok := lc.requireStage1Access(c)
	if !ok {
		return
	}

	modules, err := lc.Store.LoadModules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load theory modules"})
		return
	}

	progress, err := lc.Store.GetTheoryProgress(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load theory progress"})
		return
	}

	completedDomains, completedByDomain := buildOrderedCompletedDomains(progress)

	scoreMap := make(map[string]float64, len(profiles))
	for _, p := range profiles {
		scoreMap[p.DomainID] = math.Round(p.ProficiencyScore*100) / 100
	}
	recommended := services.DetermineRecommendedFocusDomains(scoreMap, ProficiencyThreshold)

	domainDTOs := make([]dto.TheoryDomainDTO, 0, len(modules))
	for _, mod := range modules {
		score, hasScore := scoreMap[mod.DomainID]
		passedThreshold := hasScore && score >= ProficiencyThreshold

		prog, isCompleted := completedByDomain[mod.DomainID]
		var completedAt *time.Time
		if isCompleted {
			ts := prog.CompletedAt
			completedAt = &ts
		}

		domainDTOs = append(domainDTOs, dto.TheoryDomainDTO{
			DomainID:         mod.DomainID,
			DomainName:       mod.DomainName,
			Title:            mod.Title,
			Content:          mod.Content,
			SourceURL:        mod.SourceURL,
			ProficiencyScore: score,
			Threshold:        ProficiencyThreshold,
			PassedThreshold:  passedThreshold,
			RecommendedFocus: recommended[mod.DomainID],
			IsCompleted:      isCompleted,
			CompletedAt:      completedAt,
		})
	}

	totalDomains := len(services.OrderedTheoryDomains)
	completedCount := len(completedDomains)

	c.JSON(http.StatusOK, dto.LearnDomainsResponse{
		Unlocked:         true,
		TheoryCompleted:  services.IsTheoryGateSatisfied(completedDomains),
		CompletedCount:   completedCount,
		TotalDomains:     totalDomains,
		CompletedDomains: completedDomains,
		Domains:          domainDTOs,
	})
}

// CompleteDomain godoc
// @Summary Mark a domain Theory Module as completed
// @Description Idempotently records that the authenticated user finished studying the specified domain Theory Module without altering Proficiency Score.
// @Tags Learn
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param domainId path string true "Domain ID (domain1..domain5)"
// @Success 200 {object} dto.CompleteTheoryDomainResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.TheoryGateErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /learn/domains/{domainId}/complete [post]
func (lc *LearnController) CompleteDomain(c *gin.Context) {
	domainID := c.Param("domainId")
	if !services.IsValidTheoryDomain(domainID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain_id; expected domain1 through domain5"})
		return
	}

	userID, _, ok := lc.requireStage1Access(c)
	if !ok {
		return
	}

	if err := lc.Store.MarkTheoryCompleted(userID, domainID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record domain completion"})
		return
	}

	progress, err := lc.Store.GetTheoryProgress(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated theory progress"})
		return
	}

	completedDomains, _ := buildOrderedCompletedDomains(progress)

	c.JSON(http.StatusOK, dto.CompleteTheoryDomainResponse{
		DomainID:         domainID,
		IsCompleted:      true,
		CompletedDomains: completedDomains,
		CompletedCount:   len(completedDomains),
		TotalDomains:     len(services.OrderedTheoryDomains),
		TheoryCompleted:  services.IsTheoryGateSatisfied(completedDomains),
	})
}

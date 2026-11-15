package controllers

import (
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	errSampleNotFound   = errors.New("sample question not found for this sub-topic")
	errSampleIncomplete = errors.New("all assigned sample questions must be attempted before completing this sub-topic")
)

// LearnStore abstracts persistence for the Learning Theory Hub handlers so the
// HTTP layer can be tested cleanly without a live PostgreSQL instance.
type LearnStore interface {
	GetUserSkillProfiles(userID string) ([]models.UserSkillProfile, error)
	GetTheoryProgress(userID string) ([]models.UserTheoryProgress, error)
	GetSubtopicProgress(userID string) ([]models.UserTheorySubtopicProgress, error)
	GetSubtopicSampleProgress(userID string) (map[string]int, map[string]int, error)
	MarkTheoryCompleted(userID, domainID string) error
	CompleteSubtopic(userID, role, subTopicID, domainID string) (dto.CompleteSubtopicResponse, error)
	LoadModules() ([]services.TheoryModuleContent, error)
	LoadSubtopics() ([]services.TheorySubtopicContent, error)
	GetOrAssignSampleQuestions(userID, role, subTopicID, domainID string) ([]models.UserTheorySampleQuestionDetail, error)
	SubmitSampleQuestionAttempt(userID, role, subTopicID, questionID, answer string) (dto.AttemptSampleQuestionResponse, error)
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

func (s *gormLearnStore) LoadSubtopics() ([]services.TheorySubtopicContent, error) {
	return services.LoadTheorySubtopics(s.DB)
}

func (s *gormLearnStore) GetSubtopicProgress(userID string) ([]models.UserTheorySubtopicProgress, error) {
	var progress []models.UserTheorySubtopicProgress
	if err := s.DB.Where("user_id = ?", userID).Find(&progress).Error; err != nil {
		return nil, err
	}
	return progress, nil
}

func isSampleAnswered(submitted *string) bool {
	return submitted != nil && *submitted != ""
}

func isStaffRole(role string) bool {
	return role != "learner"
}

func countAttemptedSamples(rows []models.UserTheorySampleQuestion) int {
	attempted := 0
	for _, r := range rows {
		if isSampleAnswered(r.SubmittedAnswer) {
			attempted++
		}
	}
	return attempted
}

func (s *gormLearnStore) GetSubtopicSampleProgress(userID string) (map[string]int, map[string]int, error) {
	var rows []models.UserTheorySampleQuestion
	if err := s.DB.Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	attempted := make(map[string]int)
	total := make(map[string]int)
	for _, r := range rows {
		total[r.SubTopicID]++
		if isSampleAnswered(r.SubmittedAnswer) {
			attempted[r.SubTopicID]++
		}
	}
	return attempted, total, nil
}

func (s *gormLearnStore) loadAssignedSampleDetails(tx *gorm.DB, userID, subTopicID string) ([]models.UserTheorySampleQuestionDetail, error) {
	var rows []models.UserTheorySampleQuestion
	if err := tx.Where("user_id = ? AND sub_topic_id = ?", userID, subTopicID).
		Order("order_index ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	qIDs := make([]string, 0, len(rows))
	for _, r := range rows {
		qIDs = append(qIDs, r.QuestionID)
	}
	var questions []models.Question
	if err := tx.Where("question_id IN ?", qIDs).Find(&questions).Error; err != nil {
		return nil, err
	}
	qMap := make(map[string]models.Question, len(questions))
	for _, q := range questions {
		qMap[q.QuestionID] = q
	}
	details := make([]models.UserTheorySampleQuestionDetail, 0, len(rows))
	for _, r := range rows {
		if q, ok := qMap[r.QuestionID]; ok {
			details = append(details, models.UserTheorySampleQuestionDetail{
				Assignment: r,
				Question:   q,
			})
		}
	}
	return details, nil
}

func (s *gormLearnStore) GetOrAssignSampleQuestions(userID, role, subTopicID, domainID string) ([]models.UserTheorySampleQuestionDetail, error) {
	if role != "learner" {
		return s.loadStaffPreviewSampleQuestions(subTopicID, domainID)
	}

	existing, err := s.loadAssignedSampleDetails(s.DB, userID, subTopicID)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return existing, nil
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		_ = tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", "sample_draw:"+userID).Error

		current, checkErr := s.loadAssignedSampleDetails(tx, userID, subTopicID)
		if checkErr != nil {
			return checkErr
		}
		if len(current) > 0 {
			existing = current
			return nil
		}

		exclusionQuery := BuildPostTestExclusionQuery(tx, userID)

		var primary []models.Question
		if qErr := tx.Where(
			"domain_id = ? AND sub_topic_id = ? AND is_active = true AND question_id NOT IN (?)",
			domainID, subTopicID, exclusionQuery,
		).Order("RANDOM()").Limit(2).Find(&primary).Error; qErr != nil {
			return qErr
		}

		var fallback []models.Question
		if len(primary) < 2 {
			needed := 2 - len(primary)
			fbQuery := tx.Where(
				"domain_id = ? AND is_active = true AND question_id NOT IN (?)",
				domainID, exclusionQuery,
			)
			if len(primary) > 0 {
				primaryIDs := make([]string, 0, len(primary))
				for _, pq := range primary {
					primaryIDs = append(primaryIDs, pq.QuestionID)
				}
				fbQuery = fbQuery.Where("question_id NOT IN ?", primaryIDs)
			}
			if fbErr := fbQuery.Order("RANDOM()").Limit(needed).Find(&fallback).Error; fbErr != nil {
				return fbErr
			}
		}

		now := time.Now()
		orderIdx := 1
		for _, q := range primary {
			row := models.UserTheorySampleQuestion{
				UserID:     userID,
				SubTopicID: subTopicID,
				QuestionID: q.QuestionID,
				OrderIndex: orderIdx,
				IsFallback: false,
				AssignedAt: now,
			}
			if insErr := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; insErr != nil {
				return insErr
			}
			orderIdx++
		}
		for _, q := range fallback {
			row := models.UserTheorySampleQuestion{
				UserID:     userID,
				SubTopicID: subTopicID,
				QuestionID: q.QuestionID,
				OrderIndex: orderIdx,
				IsFallback: true,
				AssignedAt: now,
			}
			if insErr := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; insErr != nil {
				return insErr
			}
			orderIdx++
		}

		loaded, loadErr := s.loadAssignedSampleDetails(tx, userID, subTopicID)
		if loadErr != nil {
			return loadErr
		}
		existing = loaded
		return nil
	})
	return existing, err
}

func (s *gormLearnStore) loadStaffPreviewSampleQuestions(subTopicID, domainID string) ([]models.UserTheorySampleQuestionDetail, error) {
	var primary []models.Question
	if err := s.DB.Where("domain_id = ? AND sub_topic_id = ? AND is_active = true", domainID, subTopicID).
		Order("question_id ASC").
		Limit(2).
		Find(&primary).Error; err != nil {
		return nil, err
	}
	var fallback []models.Question
	if len(primary) < 2 {
		fbQ := s.DB.Where("domain_id = ? AND is_active = true", domainID)
		if len(primary) > 0 {
			fbQ = fbQ.Where("question_id <> ?", primary[0].QuestionID)
		}
		if err := fbQ.Order("question_id ASC").Limit(2 - len(primary)).Find(&fallback).Error; err != nil {
			return nil, err
		}
	}
	details := make([]models.UserTheorySampleQuestionDetail, 0, len(primary)+len(fallback))
	idx := 1
	for _, q := range primary {
		details = append(details, models.UserTheorySampleQuestionDetail{
			Assignment: models.UserTheorySampleQuestion{
				SubTopicID: subTopicID,
				QuestionID: q.QuestionID,
				OrderIndex: idx,
				IsFallback: false,
			},
			Question: q,
		})
		idx++
	}
	for _, q := range fallback {
		details = append(details, models.UserTheorySampleQuestionDetail{
			Assignment: models.UserTheorySampleQuestion{
				SubTopicID: subTopicID,
				QuestionID: q.QuestionID,
				OrderIndex: idx,
				IsFallback: true,
			},
			Question: q,
		})
		idx++
	}
	return details, nil
}

func (s *gormLearnStore) SubmitSampleQuestionAttempt(userID, role, subTopicID, questionID, answer string) (dto.AttemptSampleQuestionResponse, error) {
	if role != "learner" {
		var q models.Question
		if err := s.DB.Where("question_id = ? AND is_active = true", questionID).First(&q).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return dto.AttemptSampleQuestionResponse{}, errSampleNotFound
			}
			return dto.AttemptSampleQuestionResponse{}, err
		}
		return dto.AttemptSampleQuestionResponse{
			SubTopicID:       subTopicID,
			QuestionID:       questionID,
			SubmittedAnswer:  answer,
			IsCorrect:        strings.EqualFold(answer, q.CorrectAnswer),
			CorrectAnswer:    q.CorrectAnswer,
			Explanation:      q.Explanation,
			AttemptedCount:   2,
			RequiredCount:    2,
			ReadyToComplete:  true,
			AlreadySubmitted: false,
			PreviewMode:      true,
		}, nil
	}

	var resp dto.AttemptSampleQuestionResponse
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var row models.UserTheorySampleQuestion
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND sub_topic_id = ? AND question_id = ?", userID, subTopicID, questionID).
			First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errSampleNotFound
			}
			return err
		}

		var q models.Question
		if err := tx.Where("question_id = ?", questionID).First(&q).Error; err != nil {
			return err
		}

		alreadySubmitted := isSampleAnswered(row.SubmittedAnswer)
		submittedAns := answer
		var isCorrect bool
		if !alreadySubmitted {
			isCorrect = strings.EqualFold(answer, q.CorrectAnswer)
			now := time.Now()
			if err := tx.Model(&models.UserTheorySampleQuestion{}).
				Where("user_id = ? AND sub_topic_id = ? AND question_id = ?", userID, subTopicID, questionID).
				Updates(map[string]any{
					"submitted_answer": submittedAns,
					"is_correct":       isCorrect,
					"answered_at":      now,
				}).Error; err != nil {
				return err
			}
		} else {
			submittedAns = *row.SubmittedAnswer
			if row.IsCorrect != nil {
				isCorrect = *row.IsCorrect
			}
		}

		var allRows []models.UserTheorySampleQuestion
		if err := tx.Where("user_id = ? AND sub_topic_id = ?", userID, subTopicID).Find(&allRows).Error; err != nil {
			return err
		}
		attemptedCount := countAttemptedSamples(allRows)
		requiredCount := len(allRows)
		resp = dto.AttemptSampleQuestionResponse{
			SubTopicID:       subTopicID,
			QuestionID:       questionID,
			SubmittedAnswer:  submittedAns,
			IsCorrect:        isCorrect,
			CorrectAnswer:    q.CorrectAnswer,
			Explanation:      q.Explanation,
			AttemptedCount:   attemptedCount,
			RequiredCount:    requiredCount,
			ReadyToComplete:  requiredCount > 0 && attemptedCount >= requiredCount,
			AlreadySubmitted: alreadySubmitted,
			PreviewMode:      false,
		}
		return nil
	})
	return resp, err
}

func (s *gormLearnStore) CompleteSubtopic(userID, role, subTopicID, domainID string) (dto.CompleteSubtopicResponse, error) {
	domainSubs := services.SubtopicsForDomain(domainID)
	if isStaffRole(role) {
		return dto.CompleteSubtopicResponse{
			SubTopicID:                 subTopicID,
			DomainID:                   domainID,
			IsCompleted:                true,
			DomainAutoCompleted:        false,
			CompletedSubtopicsInDomain: len(domainSubs),
			TotalSubtopicsInDomain:     len(domainSubs),
			CompletedSubtopicCount:     0,
			TotalSubtopics:             len(services.OrderedTheorySubtopics),
			CompletedSubtopics:         []string{},
			CompletedDomains:           []string{},
			CompletedCount:             0,
			TotalDomains:               len(services.OrderedTheoryDomains),
			TheoryCompleted:            false,
			PreviewMode:                true,
		}, nil
	}

	var resp dto.CompleteSubtopicResponse
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var sampleRows []models.UserTheorySampleQuestion
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND sub_topic_id = ?", userID, subTopicID).
			Find(&sampleRows).Error; err != nil {
			return err
		}

		attempted := countAttemptedSamples(sampleRows)
		if len(sampleRows) == 0 || attempted < len(sampleRows) {
			return errSampleIncomplete
		}

		subRow := models.UserTheorySubtopicProgress{
			UserID:      userID,
			SubTopicID:  subTopicID,
			CompletedAt: time.Now(),
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "sub_topic_id"}},
			DoNothing: true,
		}).Create(&subRow).Error; err != nil {
			return err
		}

		var allSubProg []models.UserTheorySubtopicProgress
		if err := tx.Where("user_id = ?", userID).Find(&allSubProg).Error; err != nil {
			return err
		}
		completedSubtopics, subByID := buildOrderedCompletedSubtopics(allSubProg)

		completedInDomain := 0
		for _, sp := range domainSubs {
			if _, ok := subByID[sp.SubTopicID]; ok {
				completedInDomain++
			}
		}

		domainAutoCompleted := false
		if len(domainSubs) > 0 && completedInDomain == len(domainSubs) {
			domRes := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "domain_id"}},
				DoNothing: true,
			}).Create(&models.UserTheoryProgress{
				UserID:      userID,
				DomainID:    domainID,
				CompletedAt: time.Now(),
			})
			if domRes.Error != nil {
				return domRes.Error
			}
			domainAutoCompleted = domRes.RowsAffected > 0
		}

		var domProg []models.UserTheoryProgress
		if err := tx.Where("user_id = ?", userID).Find(&domProg).Error; err != nil {
			return err
		}
		completedDomains, _ := buildOrderedCompletedDomains(domProg)

		resp = dto.CompleteSubtopicResponse{
			SubTopicID:                 subTopicID,
			DomainID:                   domainID,
			IsCompleted:                true,
			DomainAutoCompleted:        domainAutoCompleted,
			CompletedSubtopicsInDomain: completedInDomain,
			TotalSubtopicsInDomain:     len(domainSubs),
			CompletedSubtopicCount:     len(completedSubtopics),
			TotalSubtopics:             len(services.OrderedTheorySubtopics),
			CompletedSubtopics:         completedSubtopics,
			CompletedDomains:           completedDomains,
			CompletedCount:             len(completedDomains),
			TotalDomains:               len(services.OrderedTheoryDomains),
			TheoryCompleted:            services.IsTheorySubtopicGateSatisfied(completedSubtopics),
			PreviewMode:                false,
		}
		return nil
	})
	return resp, err
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

func buildOrderedCompletedSubtopics(progress []models.UserTheorySubtopicProgress) ([]string, map[string]models.UserTheorySubtopicProgress) {
	byID := make(map[string]models.UserTheorySubtopicProgress, len(progress))
	for _, item := range progress {
		byID[item.SubTopicID] = item
	}
	ordered := make([]string, 0, len(services.OrderedTheorySubtopics))
	for _, spec := range services.OrderedTheorySubtopics {
		if _, ok := byID[spec.SubTopicID]; ok {
			ordered = append(ordered, spec.SubTopicID)
		}
	}
	return ordered, byID
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

	subtopics, err := lc.Store.LoadSubtopics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load theory sub-topics"})
		return
	}

	progress, err := lc.Store.GetTheoryProgress(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load theory progress"})
		return
	}

	subProgress, err := lc.Store.GetSubtopicProgress(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load sub-topic progress"})
		return
	}

	attemptedBySub, totalBySub, err := lc.Store.GetSubtopicSampleProgress(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load sample question progress"})
		return
	}

	completedDomains, completedByDomain := buildOrderedCompletedDomains(progress)
	completedSubtopics, subCompletedByID := buildOrderedCompletedSubtopics(subProgress)

	subDTOsByDomain := make(map[string][]dto.TheorySubtopicSummaryDTO, len(services.OrderedTheoryDomains))
	completedSubCountByDomain := make(map[string]int, len(services.OrderedTheoryDomains))
	for _, sub := range subtopics {
		sp, isSubDone := subCompletedByID[sub.SubTopicID]
		var subDoneAt *time.Time
		if isSubDone {
			ts := sp.CompletedAt
			subDoneAt = &ts
			completedSubCountByDomain[sub.DomainID]++
		}
		sampleTotal := totalBySub[sub.SubTopicID]
		if sampleTotal == 0 {
			sampleTotal = 2
		}
		subDTOsByDomain[sub.DomainID] = append(subDTOsByDomain[sub.DomainID], dto.TheorySubtopicSummaryDTO{
			SubTopicID:           sub.SubTopicID,
			DomainID:             sub.DomainID,
			OrderIndex:           sub.OrderIndex,
			Title:                sub.Title,
			IsCompleted:          isSubDone,
			CompletedAt:          subDoneAt,
			SampleAttemptedCount: attemptedBySub[sub.SubTopicID],
			SampleTotalCount:     sampleTotal,
		})
	}

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
		domainSubs := subDTOsByDomain[mod.DomainID]

		domainDTOs = append(domainDTOs, dto.TheoryDomainDTO{
			DomainID:                   mod.DomainID,
			DomainName:                 mod.DomainName,
			Title:                      mod.Title,
			Content:                    mod.Content,
			SourceURL:                  mod.SourceURL,
			ProficiencyScore:           score,
			Threshold:                  ProficiencyThreshold,
			PassedThreshold:            passedThreshold,
			RecommendedFocus:           recommended[mod.DomainID],
			IsCompleted:                isCompleted,
			CompletedAt:                completedAt,
			CompletedSubtopicsInDomain: completedSubCountByDomain[mod.DomainID],
			TotalSubtopicsInDomain:     len(domainSubs),
			Subtopics:                  domainSubs,
		})
	}

	totalDomains := len(services.OrderedTheoryDomains)
	completedCount := len(completedDomains)

	c.JSON(http.StatusOK, dto.LearnDomainsResponse{
		Unlocked:               true,
		TheoryCompleted:        services.IsTheorySubtopicGateSatisfied(completedSubtopics),
		CompletedCount:         completedCount,
		TotalDomains:           totalDomains,
		CompletedSubtopicCount: len(completedSubtopics),
		TotalSubtopics:         len(services.OrderedTheorySubtopics),
		CompletedDomains:       completedDomains,
		CompletedSubtopics:     completedSubtopics,
		Domains:                domainDTOs,
	})
}

// CompleteDomain godoc
// @Summary Mark a domain Theory Module as completed
// @Description Idempotently records that the authenticated user finished studying the specified domain Theory Module without altering Proficiency Score. For Learners, rejects with 403 subtopics_incomplete unless all child Sub-topics are completed.
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
	role := c.GetString("role")

	subProgress, err := lc.Store.GetSubtopicProgress(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify sub-topic progress"})
		return
	}
	completedSubtopics, _ := buildOrderedCompletedSubtopics(subProgress)

	if !isStaffRole(role) {
		if missing := services.MissingSubtopicsForDomain(domainID, completedSubtopics); len(missing) > 0 {
			doneSubCount := len(completedSubtopics)
			totalSub := len(services.OrderedTheorySubtopics)
			c.JSON(http.StatusForbidden, dto.TheoryGateErrorResponse{
				Error:                  "Complete all Sub-topics in this domain before marking the domain complete.",
				Reason:                 "subtopics_incomplete",
				CompletedSubtopicCount: &doneSubCount,
				TotalSubtopics:         &totalSub,
				CompletedSubtopics:     completedSubtopics,
				MissingSubtopics:       missing,
			})
			return
		}
		if err := lc.Store.MarkTheoryCompleted(userID, domainID); err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to record domain completion"})
			return
		}
	} else {
		c.JSON(http.StatusOK, dto.CompleteTheoryDomainResponse{
			DomainID:         domainID,
			IsCompleted:      true,
			CompletedDomains: []string{domainID},
			CompletedCount:   1,
			TotalDomains:     len(services.OrderedTheoryDomains),
			TheoryCompleted:  false,
		})
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
		TheoryCompleted:  services.IsTheorySubtopicGateSatisfied(completedSubtopics),
	})
}

// GetSubTopic godoc
// @Summary Retrieve a Sub-topic lesson and its 2 locked sample questions
// @Description Returns the web-grounded Markdown lesson for a Security+ Sub-topic alongside 2 permanently assigned formative sample questions (drawn cleanly without overlap with Pre-Test or future Post-Test). Correct answers and explanations stay hidden until submitted.
// @Tags Learn
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subTopicId path string true "Sub-topic ID (e.g. 1.1..5.6)"
// @Success 200 {object} dto.LearnSubtopicDetailResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.TheoryGateErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /learn/subtopics/{subTopicId} [get]
func (lc *LearnController) GetSubTopic(c *gin.Context) {
	subTopicID := c.Param("subTopicId")
	spec, okSpec := services.GetSubtopicSpec(subTopicID)
	if !okSpec {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Unknown sub_topic_id"})
		return
	}

	userID, _, ok := lc.requireStage1Access(c)
	if !ok {
		return
	}
	role := c.GetString("role")

	subtopics, err := lc.Store.LoadSubtopics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to load theory sub-topics"})
		return
	}

	var matchedSub *services.TheorySubtopicContent
	for i := range subtopics {
		if subtopics[i].SubTopicID == subTopicID {
			matchedSub = &subtopics[i]
			break
		}
	}
	if matchedSub == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Sub-topic content not found"})
		return
	}

	subProgress, err := lc.Store.GetSubtopicProgress(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to load sub-topic progress"})
		return
	}
	var isCompleted bool
	var completedAt *time.Time
	for _, sp := range subProgress {
		if sp.SubTopicID == subTopicID {
			isCompleted = true
			ts := sp.CompletedAt
			completedAt = &ts
			break
		}
	}

	samples, err := lc.Store.GetOrAssignSampleQuestions(userID, role, subTopicID, spec.DomainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to assign sample questions"})
		return
	}

	sampleDTOs := make([]dto.LearnSubtopicSampleQuestionDTO, 0, len(samples))
	attemptedCount := 0
	for _, s := range samples {
		item := dto.LearnSubtopicSampleQuestionDTO{
			QuestionID:      s.Assignment.QuestionID,
			OrderIndex:      s.Assignment.OrderIndex,
			QuestionText:    s.Question.QuestionText,
			Options:         s.Question.Options,
			IsFallback:      s.Assignment.IsFallback,
			SubmittedAnswer: s.Assignment.SubmittedAnswer,
			IsCorrect:       s.Assignment.IsCorrect,
			AnsweredAt:      s.Assignment.AnsweredAt,
		}
		if isSampleAnswered(s.Assignment.SubmittedAnswer) {
			attemptedCount++
			ans := s.Question.CorrectAnswer
			exp := s.Question.Explanation
			item.CorrectAnswer = &ans
			item.Explanation = &exp
		}
		sampleDTOs = append(sampleDTOs, item)
	}

	prevID, nextID := services.GetSubtopicNeighbors(subTopicID)
	requiredCount := len(sampleDTOs)
	previewMode := isStaffRole(role)

	c.JSON(http.StatusOK, dto.LearnSubtopicDetailResponse{
		SubTopicID:         matchedSub.SubTopicID,
		DomainID:           matchedSub.DomainID,
		OrderIndex:         matchedSub.OrderIndex,
		Title:              matchedSub.Title,
		Content:            matchedSub.Content,
		SourceURL:          matchedSub.SourceURL,
		IsCompleted:        isCompleted,
		CompletedAt:        completedAt,
		PreviewMode:        previewMode,
		PreviousSubTopicID: prevID,
		NextSubTopicID:     nextID,
		AttemptedCount:     attemptedCount,
		RequiredCount:      requiredCount,
		ReadyToComplete:    previewMode || (requiredCount > 0 && attemptedCount >= requiredCount),
		SampleQuestions:    sampleDTOs,
	})
}

// AttemptSampleQuestion godoc
// @Summary Submit an answer for a Sub-topic sample question
// @Description Immediately evaluates a Learner's answer choice (A-D) for an assigned sample question, reveals the explanation, and locks the first submitted answer. Does not modify Proficiency Score.
// @Tags Learn
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subTopicId path string true "Sub-topic ID (e.g. 1.1..5.6)"
// @Param questionId path string true "Assigned Sample Question UUID"
// @Param body body dto.AttemptSampleQuestionRequest true "Answer choice A-D"
// @Success 200 {object} dto.AttemptSampleQuestionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.TheoryGateErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /learn/subtopics/{subTopicId}/sample-questions/{questionId}/attempt [post]
func (lc *LearnController) AttemptSampleQuestion(c *gin.Context) {
	subTopicID := c.Param("subTopicId")
	if _, okSpec := services.GetSubtopicSpec(subTopicID); !okSpec {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Unknown sub_topic_id"})
		return
	}

	userID, _, ok := lc.requireStage1Access(c)
	if !ok {
		return
	}
	role := c.GetString("role")

	var req dto.AttemptSampleQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Answer is required"})
		return
	}

	ans := strings.ToUpper(strings.TrimSpace(req.Answer))
	switch ans {
	case "A", "B", "C", "D":
	default:
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Answer must be one of A, B, C, or D"})
		return
	}

	questionID := c.Param("questionId")
	resp, err := lc.Store.SubmitSampleQuestionAttempt(userID, role, subTopicID, questionID, ans)
	if err != nil {
		if errors.Is(err, errSampleNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Sample question not assigned to this user and sub-topic"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to record sample question attempt"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CompleteSubTopic godoc
// @Summary Mark a Sub-topic as completed (attempt-gated)
// @Description Verifies that all assigned sample questions for the Learner in this Sub-topic have been attempted (400 sample_questions_incomplete otherwise), idempotently marks the Sub-topic complete, and automatically completes the parent domain when all child Sub-topics are done. Specialists/Admins receive preview_mode=true without database writes.
// @Tags Learn
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subTopicId path string true "Sub-topic ID (e.g. 1.1..5.6)"
// @Success 200 {object} dto.CompleteSubtopicResponse
// @Failure 400 {object} dto.TheoryGateErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.TheoryGateErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /learn/subtopics/{subTopicId}/complete [post]
func (lc *LearnController) CompleteSubTopic(c *gin.Context) {
	subTopicID := c.Param("subTopicId")
	spec, okSpec := services.GetSubtopicSpec(subTopicID)
	if !okSpec {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Unknown sub_topic_id"})
		return
	}

	userID, _, ok := lc.requireStage1Access(c)
	if !ok {
		return
	}
	role := c.GetString("role")

	resp, err := lc.Store.CompleteSubtopic(userID, role, subTopicID, spec.DomainID)
	if err != nil {
		if errors.Is(err, errSampleIncomplete) {
			c.JSON(http.StatusBadRequest, dto.TheoryGateErrorResponse{
				Error:  "Attempt all assigned sample questions for this Sub-topic before marking it completed.",
				Reason: "sample_questions_incomplete",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to record sub-topic completion"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

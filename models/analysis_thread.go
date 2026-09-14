package models

import (
	"time"

	"cybersim/dto"
)

type AnalysisThread struct {
	ThreadID    string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"thread_id"`
	AuthorID    string     `gorm:"type:uuid;not null;index" json:"author_id"`
	Author      *User      `gorm:"foreignKey:AuthorID;references:UserID;constraint:-" json:"author,omitempty"`
	ScenarioID  *string    `gorm:"type:uuid;index" json:"scenario_id,omitempty"`
	Scenario    *Scenario  `gorm:"foreignKey:ScenarioID;references:ScenarioID;constraint:-" json:"scenario,omitempty"`
	Title       string     `gorm:"type:varchar(255);not null" json:"title"`
	Content     string     `gorm:"type:text;not null" json:"content"`
	Tags        any        `gorm:"type:jsonb;serializer:json" json:"tags"`
	UpvoteCount int        `gorm:"default:0" json:"upvote_count"`
	ViewCount   int        `gorm:"default:0" json:"view_count"`
	IsPinned    bool       `gorm:"default:false" json:"is_pinned"`
	IsLocked    bool       `gorm:"default:false" json:"is_locked"`
	CreatedAt   time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"default:now()" json:"updated_at"`
}

// ToResponseDTO converts AnalysisThread into ThreadResponse
func (t *AnalysisThread) ToResponseDTO() dto.ThreadResponse {
	authorDTO := dto.ThreadAuthorDTO{
		UserID: t.AuthorID,
	}
	if t.Author != nil {
		authorDTO.Nickname = t.Author.Nickname
		authorDTO.Role = t.Author.Role
	}

	var scenarioDTO *dto.ThreadScenarioDTO
	if t.Scenario != nil {
		mitre := ""
		if t.Scenario.MitreTechnique != nil {
			mitre = *t.Scenario.MitreTechnique
		}
		scenarioDTO = &dto.ThreadScenarioDTO{
			ScenarioID:     t.Scenario.ScenarioID,
			Title:          t.Scenario.Title,
			DomainID:       t.Scenario.DomainID,
			MitreTechnique: mitre,
		}
	}

	tagsList := make([]string, 0)
	if t.Tags != nil {
		if rawList, ok := t.Tags.([]interface{}); ok {
			for _, item := range rawList {
				if s, ok := item.(string); ok {
					tagsList = append(tagsList, s)
				}
			}
		} else if strList, ok := t.Tags.([]string); ok {
			tagsList = strList
		}
	}

	return dto.ThreadResponse{
		ThreadID:    t.ThreadID,
		Author:      authorDTO,
		Scenario:    scenarioDTO,
		Title:       t.Title,
		Content:     t.Content,
		Tags:        tagsList,
		UpvoteCount: t.UpvoteCount,
		ViewCount:   t.ViewCount,
		IsPinned:    t.IsPinned,
		IsLocked:    t.IsLocked,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
	}
}

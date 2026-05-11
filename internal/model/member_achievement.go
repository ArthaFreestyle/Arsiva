package model

// ==================== Requests ====================

type MemberAchievementCreateRequest struct {
	AchievementId string `json:"achievement_id" validate:"required,numeric"`
}

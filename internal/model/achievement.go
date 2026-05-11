package model

// ==================== Requests ====================

type AchievementCreateRequest struct {
	Nama       string `json:"nama"        validate:"required,min=1,max=50"`
	Deskripsi  string `json:"deskripsi"   validate:"max=2000"`
	BadgeIcon  string `json:"badge_icon"  validate:"required,url,max=255"`
	XPRequired int    `json:"xp_required" validate:"gte=0"`
	Tier       string `json:"tier"        validate:"required,oneof=bronze silver gold platinum"`
}

type AchievementUpdateRequest struct {
	Nama       string `json:"nama"        validate:"required,min=1,max=50"`
	Deskripsi  string `json:"deskripsi"   validate:"max=2000"`
	BadgeIcon  string `json:"badge_icon"  validate:"required,url,max=255"`
	XPRequired int    `json:"xp_required" validate:"gte=0"`
	Tier       string `json:"tier"        validate:"required,oneof=bronze silver gold platinum"`
}

// ==================== Responses ====================

type AchievementResponse struct {
	AchievementId string `json:"achievement_id"`
	Nama          string `json:"nama"`
	Deskripsi     string `json:"deskripsi"`
	BadgeIcon     string `json:"badge_icon"`
	XPRequired    int    `json:"xp_required"`
	Tier          string `json:"tier"`
}

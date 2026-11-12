package model

import "time"

type AchievementPayload struct {
	AchievementID string `json:"achievement_id"`
	Nama          string `json:"nama"`
	BadgeIcon     string `json:"badge_icon"`
	Tier          string `json:"tier"`
}
type UnlockedAchievement struct {
	AchievementPayload
	UnlockedAt     time.Time `json:"unlocked_at"`
	NotificationID string    `json:"notification_id"`
}
type Notification struct {
	NotificationID string             `json:"notification_id"`
	Type           string             `json:"type"`
	Title          string             `json:"title"`
	Message        string             `json:"message"`
	Payload        AchievementPayload `json:"payload"`
	CreatedAt      time.Time          `json:"created_at"`
	ReadAt         *time.Time         `json:"read_at"`
}

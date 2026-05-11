package entity

type TierAchievement string

const (
	TierBronze   TierAchievement = "bronze"
	TierSilver   TierAchievement = "silver"
	TierGold     TierAchievement = "gold"
	TierPlatinum TierAchievement = "platinum"
)

type Achievement struct {
	AchievementId string          `db:"achievement_id"`
	Nama          string          `db:"nama"`
	Deskripsi     string          `db:"deskripsi"`
	BadgeIcon     string          `db:"badge_icon"`
	XPRequired    int             `db:"xp_required"`
	Tier          TierAchievement `db:"tier"`
}

package entity

type MemberAchievement struct {
	AchievementId string `db:"achievement_id"`
	Nama          string `db:"nama"`
	Deskripsi     string `db:"deskripsi"`
	BadgeIcon     string `db:"badge_icon"`
	XPRequired    int    `db:"xp_required"`
	Tier          string `db:"tier"`
	UnlockedAt    string `db:"unlocked_at"`
}

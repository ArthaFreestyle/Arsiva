package model

type MemberAchievementResponse struct {
	AchievementId string `json:"achievement_id"`
	Nama          string `json:"nama"`
	Deskripsi     string `json:"deskripsi"`
	BadgeIcon     string `json:"badge_icon"`
	XPRequired    int    `json:"xp_required"`
	Tier          string `json:"tier"`
	UnlockedAt    string `json:"unlocked_at"`
}

type MemberProfileResponse struct {
	MemberId     string                       `json:"member_id"`
	Username     string                       `json:"username"`
	Email        string                       `json:"email"`
	NIS          string                       `json:"nis"`
	FotoProfil   string                       `json:"foto_profil"`
	Bio          string                       `json:"bio"`
	TanggalLahir string                       `json:"tanggal_lahir"`
	JenisKelamin string                       `json:"jenis_kelamin"`
	Minat        string                       `json:"minat"`
	LastActive   string                       `json:"last_active"`
	TotalXP      int                          `json:"total_xp"`
	Level        int                          `json:"level"`
	Sekolah      *SekolahResponse             `json:"sekolah,omitempty"`
	Achievements []*MemberAchievementResponse `json:"achievements"`
	SocialLinks  []*MemberSocialLinkResponse  `json:"social_links"`
}

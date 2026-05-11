package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToMemberAchievementResponse(ma *entity.MemberAchievement) *model.MemberAchievementResponse {
	if ma == nil {
		return nil
	}
	return &model.MemberAchievementResponse{
		AchievementId: ma.AchievementId,
		Nama:          ma.Nama,
		Deskripsi:     ma.Deskripsi,
		BadgeIcon:     ma.BadgeIcon,
		XPRequired:    ma.XPRequired,
		Tier:          ma.Tier,
		UnlockedAt:    ma.UnlockedAt,
	}
}

func ToMemberAchievementResponses(achievements []*entity.MemberAchievement) []*model.MemberAchievementResponse {
	responses := make([]*model.MemberAchievementResponse, len(achievements))
	for i, a := range achievements {
		responses[i] = ToMemberAchievementResponse(a)
	}
	return responses
}

package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToAchievementResponse(achievement *entity.Achievement) *model.AchievementResponse {
	if achievement == nil {
		return nil
	}
	return &model.AchievementResponse{
		AchievementId: achievement.AchievementId,
		Nama:          achievement.Nama,
		Deskripsi:     achievement.Deskripsi,
		BadgeIcon:     achievement.BadgeIcon,
		XPRequired:    achievement.XPRequired,
		Tier:          string(achievement.Tier),
	}
}

func ToAchievementResponses(achievements []*entity.Achievement) []*model.AchievementResponse {
	responses := make([]*model.AchievementResponse, len(achievements))
	for i, a := range achievements {
		responses[i] = ToAchievementResponse(a)
	}
	return responses
}

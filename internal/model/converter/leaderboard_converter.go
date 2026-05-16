package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToPublicLeaderboardItem(e *entity.LeaderboardEntry) model.PublicLeaderboardItem {
	item := model.PublicLeaderboardItem{
		Rank:       e.Rank,
		MemberId:   e.MemberId,
		Username:   e.Username,
		FotoProfil: e.FotoProfil,
		Level:      e.Level,
		TotalXP:    e.TotalXP,
		MonthlyXP:  e.MonthlyXP,
	}
	if e.SekolahId != nil && e.SekolahNama != nil {
		item.Sekolah = &model.LeaderboardSekolah{
			SekolahId:   *e.SekolahId,
			NamaSekolah: *e.SekolahNama,
		}
	}
	return item
}

func ToPublicLeaderboardItems(entries []entity.LeaderboardEntry) []model.PublicLeaderboardItem {
	items := make([]model.PublicLeaderboardItem, len(entries))
	for i := range entries {
		items[i] = ToPublicLeaderboardItem(&entries[i])
	}
	return items
}

func ToGroupLeaderboardItem(e *entity.LeaderboardEntry) model.GroupLeaderboardItem {
	return model.GroupLeaderboardItem{
		Rank:           e.Rank,
		MemberId:       e.MemberId,
		Username:       e.Username,
		FotoProfil:     e.FotoProfil,
		Level:          e.Level,
		GroupXP:        e.GroupXP,
		CompletedCount: e.CompletedCount,
		TotalXP:        e.TotalXP,
	}
}

func ToGroupLeaderboardItems(entries []entity.LeaderboardEntry) []model.GroupLeaderboardItem {
	items := make([]model.GroupLeaderboardItem, len(entries))
	for i := range entries {
		items[i] = ToGroupLeaderboardItem(&entries[i])
	}
	return items
}

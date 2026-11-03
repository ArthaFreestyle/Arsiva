package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToUserRoleCountResponse(c entity.UserRoleCount) model.UserRoleCountResponse {
	return model.UserRoleCountResponse{
		SuperAdmin: c.SuperAdmin,
		Guru:       c.Guru,
		Member:     c.Member,
	}
}

// ToUserSummaryResponse derives total from the per-role counts so it can never
// drift from by_role.
func ToUserSummaryResponse(c entity.UserRoleCount) *model.UserSummaryResponse {
	return &model.UserSummaryResponse{
		Total:  c.SuperAdmin + c.Guru + c.Member,
		ByRole: ToUserRoleCountResponse(c),
	}
}

func toPublishCount(total, published int) model.PublishCountResponse {
	return model.PublishCountResponse{
		Total:     total,
		Published: published,
		Draft:     total - published,
	}
}

// ToContentSummaryResponse derives draft and game totals from the raw counts.
// Articles are not games and stay out of game.*.
func ToContentSummaryResponse(c entity.ContentCount) *model.ContentSummaryResponse {
	kuis := toPublishCount(c.KuisTotal, c.KuisPublished)
	cerita := toPublishCount(c.CeritaTotal, c.CeritaPublished)
	puzzle := toPublishCount(c.PuzzleTotal, c.PuzzlePublished)

	return &model.ContentSummaryResponse{
		Game: model.GameSummaryResponse{
			Total:     kuis.Total + cerita.Total + puzzle.Total,
			Published: kuis.Published + cerita.Published + puzzle.Published,
			Draft:     kuis.Draft + cerita.Draft + puzzle.Draft,
			Kuis:      kuis,
			Cerita:    cerita,
			Puzzle:    puzzle,
		},
		Artikel: toPublishCount(c.ArtikelTotal, c.ArtikelPublished),
	}
}

func ToContentTypeCountResponse(c entity.ContentTypeCount) model.ContentTypeCountResponse {
	return model.ContentTypeCountResponse{
		Kuis:    c.Kuis,
		Cerita:  c.Cerita,
		Puzzle:  c.Puzzle,
		Artikel: c.Artikel,
		Total:   c.Kuis + c.Cerita + c.Puzzle + c.Artikel,
	}
}

package entity

import "time"

type Quiz struct {
	QuizId      int        `db:"kuis_id"`
	Judul       string     `db:"judul"`
	ThumbnailAssetId *int         `db:"thumbnail_asset_id"`
	GambarAssetId    *int         `db:"gambar_asset_id"`
	Gambar      string     `db:"gambar"`
	Thumbnail   string     `db:"thumbnail"`
	KategoriId  string     `db:"kategori_id"`
	Kategori    string     `db:"kategori"`
	XpReward    int        `db:"xp_reward"`
	CreatedAt   *time.Time `db:"created_at"`
	CreatedBy   User       `db:"user"`
	IsPublished bool       `db:"is_published"`
	Soal        []*Question
}
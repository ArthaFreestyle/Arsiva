package entity

import "time"

type Quiz struct {
	QuizId int `db:"kuis_id"`
	Judul string `db:"judul"`
	Gambar string `db:"gambar"`
	Thumbnail string `db:"thumbnail"`
	KategoriId int `db:"kategori_id"`
	XpReward int `db:"xp_reward"`
	CreatedAt time.Time `db:"created_at"`
	CreatedBy int `db:"created_by"`
	IsPublished bool `db:"is_published"`
	Soal []*Question `db:"soal"`
}
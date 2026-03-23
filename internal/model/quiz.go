package model

import "time"

type Quiz struct {
	PuzzleId int `json:"puzzle_id"`
	Judul string `json:"judul"`
	Gambar string `json:"gambar"`
	Thumbnail string `json:"thumbnail"`
	Kategori string `json:"kategori"`
	XpReward int `json:"xp_reward"`
	Soal []*Question `json:"soal"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy UserResponse `json:"created_by"`
	IsPublished bool `json:"is_published"`
}
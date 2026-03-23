package entity

import "time"

type Puzzle struct {
	PuzzleId	string		`db:"puzzle_id"`
	Judul		string		`db:"judul"`
	Thumbnail	string		`db:"thumbnail"`
	Gambar		string		`db:"gambar"`
	Kategori	string		`db:"kategori"`
	XpReward	int			`db:"xp_reward"`
	CreatedBy	User		`db:"user"`
	CreatedAt	*time.Time	`db:"created_at"`
	IsPublished	bool		`db:"is_published"`
}

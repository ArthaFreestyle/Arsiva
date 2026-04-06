package model

import "time"

type PuzzleRequest struct {
	Judul		string	`json:"judul"`
	GambarAssetId    *int	`json:"gambar_asset_id"`
	ThumbnailAssetId *int	`json:"thumbnail_asset_id"`
	Kategori	string	`json:"kategori"`
	XpReward	int		`json:"xp_reward"`
	IsPublished	bool	`json:"is_published"`
}

type PuzzleResponse struct {
	PuzzleId	string	`json:"puzzle_id"`
	Judul		string	`json:"judul"`
	Thumbnail	string	`json:"thumbnail"`
	Gambar		string	`json:"gambar"`
	Kategori	string	`json:"kategori"`
	XpReward	int		`json:"xp_reward"`
	CreatedBy	UserResponse	`json:"user"`
	CreatedAt	*time.Time	`json:"created_at"`
	IsPublished	bool		`json:"is_published"`
}
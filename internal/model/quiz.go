package model

import "time"

type QuizRequest struct {
	Judul       string             `json:"judul" validate:"required"`
	GambarAssetId    *int               `json:"gambar_asset_id"`
	ThumbnailAssetId *int               `json:"thumbnail_asset_id"`
	KategoriId       string             `json:"kategori_id" validate:"required"`
	XpReward    int                `json:"xp_reward"`
	IsPublished bool               `json:"is_published"`
	Soal        []*QuestionRequest `json:"soal" validate:"required,dive"`
}

type QuizResponse struct {
	QuizId      int                 `json:"quiz_id"`
	Judul       string              `json:"judul"`
	Gambar      *AssetResponse      `json:"gambar"`
	Thumbnail   *AssetResponse      `json:"thumbnail"`
	Kategori    string              `json:"kategori"`
	XpReward    int                 `json:"xp_reward"`
	Soal        []*QuestionResponse `json:"soal,omitempty"`
	CreatedAt   *time.Time          `json:"created_at"`
	CreatedBy   UserResponse        `json:"created_by"`
	IsPublished bool                `json:"is_published"`
}
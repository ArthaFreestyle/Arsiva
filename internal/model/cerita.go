package model

import "time"

type SceneChoice struct {
	Text string `json:"text"`
	Next string `json:"next"`
}

type SceneRequest struct {
	SceneKey          string        `json:"scene_key" validate:"required"`
	SceneImageAssetId *int          `json:"scene_image_asset_id"`
	SceneText         string        `json:"scene_text" validate:"required"`
	SceneChoices      []SceneChoice `json:"scene_choices"`
	IsEnding          bool          `json:"is_ending"`
	EndingPoint       int           `json:"ending_point"`
	EndingType        string        `json:"ending_type"`
	Urutan            int           `json:"urutan"`
}

type SceneResponse struct {
	SceneId      int           `json:"scene_id"`
	CeritaId     int           `json:"cerita_id"`
	SceneKey     string        `json:"scene_key"`
	SceneImage   *AssetResponse `json:"scene_image"`
	SceneText    string        `json:"scene_text"`
	SceneChoices []SceneChoice `json:"scene_choices"`
	IsEnding     bool          `json:"is_ending"`
	EndingPoint  int           `json:"ending_point"`
	EndingType   string        `json:"ending_type"`
	Urutan       int           `json:"urutan"`
}

// PublicSceneResponse is the member-facing scene shape. It deliberately omits
// EndingPoint (per-ending score / answer key), EndingType (reveals the
// desirable ending), and Urutan (authoring order), so a player cannot compute
// the optimal path without playing. Scenes are still returned in authored
// order by the repository's ORDER BY urutan.
type PublicSceneResponse struct {
	SceneId      int            `json:"scene_id"`
	CeritaId     int            `json:"cerita_id"`
	SceneKey     string         `json:"scene_key"`
	SceneImage   *AssetResponse `json:"scene_image"`
	SceneText    string         `json:"scene_text"`
	SceneChoices []SceneChoice  `json:"scene_choices"`
	IsEnding     bool           `json:"is_ending"`
}

type CeritaRequest struct {
	Judul            string `json:"judul" validate:"required"`
	ThumbnailAssetId *int   `json:"thumbnail_asset_id"`
	Deskripsi        string `json:"deskripsi"`
	KategoriId       int    `json:"kategori_id" validate:"required"`
	XpReward         int    `json:"xp_reward"`
	IsPublished      bool   `json:"is_published"`
}

type CeritaResponse struct {
	CeritaId    int              `json:"cerita_id"`
	Judul       string           `json:"judul"`
	Thumbnail   *AssetResponse   `json:"thumbnail"`
	Deskripsi   string           `json:"deskripsi"`
	KategoriId  int              `json:"kategori_id"`
	XpReward    int              `json:"xp_reward"`
	Scenes      []*SceneResponse `json:"scenes,omitempty"`
	CreatedAt   *time.Time       `json:"created_at"`
	CreatedBy   UserResponse     `json:"created_by"`
	IsPublished bool             `json:"is_published"`
}

// PublicCeritaResponse is the member-facing story shape returned by
// GET /v1/stories/:id. It carries the same story-level fields as
// CeritaResponse but embeds PublicSceneResponse so per-scene scoring/ordering
// metadata is never serialized to members.
type PublicCeritaResponse struct {
	CeritaId    int                    `json:"cerita_id"`
	Judul       string                 `json:"judul"`
	Thumbnail   *AssetResponse         `json:"thumbnail"`
	Deskripsi   string                 `json:"deskripsi"`
	KategoriId  int                    `json:"kategori_id"`
	XpReward    int                    `json:"xp_reward"`
	Scenes      []*PublicSceneResponse `json:"scenes,omitempty"`
	CreatedAt   *time.Time             `json:"created_at"`
	CreatedBy   UserResponse           `json:"created_by"`
	IsPublished bool                   `json:"is_published"`
}

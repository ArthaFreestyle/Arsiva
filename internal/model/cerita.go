package model

import "time"

type SceneChoice struct {
	Text string `json:"text"`
	Next string `json:"next"`
}

type SceneRequest struct {
	SceneKey     string        `json:"scene_key" validate:"required"`
	SceneImage   string        `json:"scene_image"`
	SceneText    string        `json:"scene_text" validate:"required"`
	SceneChoices []SceneChoice `json:"scene_choices"`
	IsEnding     bool          `json:"is_ending"`
	EndingPoint  int           `json:"ending_point"`
	EndingType   string        `json:"ending_type"`
	Urutan       int           `json:"urutan"`
}

type SceneResponse struct {
	SceneId      int           `json:"scene_id"`
	CeritaId     int           `json:"cerita_id"`
	SceneKey     string        `json:"scene_key"`
	SceneImage   string        `json:"scene_image"`
	SceneText    string        `json:"scene_text"`
	SceneChoices []SceneChoice `json:"scene_choices"`
	IsEnding     bool          `json:"is_ending"`
	EndingPoint  int           `json:"ending_point"`
	EndingType   string        `json:"ending_type"`
	Urutan       int           `json:"urutan"`
}

type CeritaRequest struct {
	Judul       string          `json:"judul" validate:"required"`
	Thumbnail   string          `json:"thumbnail"`
	Deskripsi   string          `json:"deskripsi"`
	KategoriId  int             `json:"kategori_id" validate:"required"`
	XpReward    int             `json:"xp_reward"`
	IsPublished bool            `json:"is_published"`
	Scenes      []*SceneRequest `json:"scenes"`
}

type CeritaResponse struct {
	CeritaId    int              `json:"cerita_id"`
	Judul       string           `json:"judul"`
	Thumbnail   string           `json:"thumbnail"`
	Deskripsi   string           `json:"deskripsi"`
	KategoriId  int              `json:"kategori_id"`
	XpReward    int              `json:"xp_reward"`
	Scenes      []*SceneResponse `json:"scenes,omitempty"`
	CreatedAt   *time.Time       `json:"created_at"`
	CreatedBy   UserResponse     `json:"created_by"`
	IsPublished bool             `json:"is_published"`
}

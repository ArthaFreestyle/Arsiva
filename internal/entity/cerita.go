package entity

import "time"

type CeritaInteraktif struct {
	CeritaId    int        `db:"cerita_id"`
	Judul       string     `db:"judul"`
	Thumbnail   string     `db:"thumbnail"`
	Deskripsi   string     `db:"deskripsi"`
	KategoriId  int        `db:"kategori_id"`
	XpReward    int        `db:"xp_reward"`
	CreatedBy   User       `db:"user"`
	CreatedAt   *time.Time `db:"created_at"`
	IsPublished bool       `db:"is_published"`
	Scenes      []*Scene
}

type Scene struct {
	SceneId      int                      `db:"scene_id"`
	CeritaId     int                      `db:"cerita_id"`
	SceneKey     string                   `db:"scene_key"`
	SceneImage   string                   `db:"scene_image"`
	SceneText    string                   `db:"scene_text"`
	SceneChoices []map[string]interface{} `db:"scene_choices"`
	IsEnding     bool                     `db:"is_ending"`
	EndingPoint  int                      `db:"ending_point"`
	EndingType   string                   `db:"ending_type"`
	Urutan       int                      `db:"urutan"`
}

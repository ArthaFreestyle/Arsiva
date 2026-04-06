package entity

import "time"

type Asset struct {
	AssetId   int        `db:"asset_id"`
	Url       string     `db:"url"`
	IsUsed    bool       `db:"is_used"`
	CreatedAt time.Time  `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

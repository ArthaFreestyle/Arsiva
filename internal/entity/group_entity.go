package entity

import (
	"time"
)

type Group struct {
	GroupId               string     `db:"group_id"`
	GroupName             string     `db:"group_name"`
	GroupThumbnailAssetId *int       `db:"group_thumbnail_asset_id"`
	GroupThumbnail        *string    `db:"group_thumbnail"`
	CreatedBy             int        `db:"created_by"`
	CreatedAt             *time.Time `db:"created_at"`
	UpdatedAt             *time.Time `db:"updated_at"`
	Guru                  Guru       `db:"guru"`
	MemberCount           int        `db:"member_count"`
}
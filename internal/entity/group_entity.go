package entity

import (
	"time"
)

type Group struct {
	GroupId			string 		`db:"group_id"`
	GroupName		string 		`db:"group_name"`
	GroupThumbnail	string 		`db:"group_thumbnail"`
	CreatedBy		User 		`db:"created_by"`
	CreatedAt		*time.Time 	`db:"created_at"`
	UpdatedAt		*time.Time	`db:"updated_at"`
	Guru			[]Guru		`db:"guru"`
	
	
}
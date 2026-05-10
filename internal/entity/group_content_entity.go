package entity

type GroupContent struct {
	GroupContentId int    `db:"group_content_id"`
	GroupId        string `db:"group_id"`
	ContentType    string `db:"content_type"`
	ContentId      int    `db:"content_id"`
	Judul          string `db:"judul"`
	Thumbnail      string `db:"thumbnail"`
}

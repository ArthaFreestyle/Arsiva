package entity

import "time"

type Article struct {
	ArticleId 		string 			`db:"artikel_id"`
	Judul			string			`db:"judul"`
	Slug			string			`db:"slug"`
	Excerpt			string			`db:"excerpt"`
	Konten			string			`db:"konten"`
	KategoriId		ArticleCategory	`db:"kategori"`
	Status			string			`db:"status"`
	CreatedBy		User			`db:"user"`
	CreatedAt		*time.Time			`db:"created_at"`
	Thumbnail		string			`db:"thumbnail"`
}
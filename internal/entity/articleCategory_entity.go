package entity

type ArticleCategory struct {
	ArticleCategoryId 	string 		`db:"kategori_artikel_id"`
	NamaKategori 		string 		`db:"nama_kategori"`
}
package model

type ArticleCategoryResponse struct {
	ArticleCategoryId 		string `json:"article_category_id"`
	NamaKategori 			string `json:"nama_kategori,omitempty"`
}

type ArticleCategoryRequest struct {
	NamaKategori 			string `json:"nama_kategori" validate:"required"`
}
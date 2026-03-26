package model

type StoryCategoryResponse struct {
	StoryCategoryId string `json:"story_category_id"`
	NamaKategori    string `json:"nama_kategori,omitempty"`
}

type StoryCategoryRequest struct {
	NamaKategori string `json:"nama_kategori" validate:"required"`
}

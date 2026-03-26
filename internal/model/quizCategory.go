package model

import "time"

type QuizCategoryResponse struct {
	QuizCategoryId string     `json:"quiz_category_id"`
	NamaKategori   string     `json:"nama_kategori,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	CreatedBy      string     `json:"created_by,omitempty"`
	Deskripsi      *string    `json:"deskripsi,omitempty"`
}

type QuizCategoryRequest struct {
	NamaKategori string `json:"nama_kategori" validate:"required"`
	Deskripsi    string `json:"deskripsi"`
}

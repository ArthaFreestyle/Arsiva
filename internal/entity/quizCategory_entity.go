package entity

import "time"

type QuizCategory struct {
	QuizCategoryId string     `db:"kategori_id"`
	NamaKategori   string     `db:"nama_kategori"`
	CreatedAt      *time.Time `db:"created_at"`
	CreatedBy      string     `db:"created_by"`
	Deskripsi      *string    `db:"deskripsi"`
}

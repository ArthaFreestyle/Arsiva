package model

type Question struct{
	PertanyaanId int `json:"pertanyaan_id"`
	KuisId int `json:"kuis_id"`
	TeksPertanyaan string `json:"teks_pertanyaan"`
	Image string `json:"image"`
	Tipe string `json:"tipe"`
	Poin int `json:"poin"`
	Urutan int `json:"urutan"`
}
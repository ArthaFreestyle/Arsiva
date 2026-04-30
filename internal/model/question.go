package model

type QuestionRequest struct {
	TeksPertanyaan string           `json:"teks_pertanyaan" validate:"required"`
	ImageAssetId   *int             `json:"image_asset_id"`
	Tipe           string           `json:"tipe" validate:"required"`
	Poin           int              `json:"poin" validate:"required,min=1"`
	Urutan         int              `json:"urutan" validate:"required,min=1"`
	Pilihan        []*OptionRequest `json:"pilihan" validate:"required,dive"`
}

type QuestionResponse struct {
	PertanyaanId   int               `json:"pertanyaan_id"`
	KuisId         int               `json:"kuis_id"`
	TeksPertanyaan string            `json:"teks_pertanyaan"`
	Image          *AssetResponse    `json:"image"`
	Tipe           string            `json:"tipe"`
	Poin           int               `json:"poin"`
	Urutan         int               `json:"urutan"`
	Pilihan        []*OptionResponse `json:"pilihan"`
}
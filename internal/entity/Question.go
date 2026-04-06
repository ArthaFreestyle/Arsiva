package entity

type Question struct {
	PertanyaanId   int       `db:"pertanyaan_id"`
	KuisId         int       `db:"kuis_id"`
	TeksPertanyaan string    `db:"teks_pertanyaan"`
	ImageAssetId   *int      `db:"image_asset_id"`
	Image          string    `db:"image"`
	Tipe           string    `db:"tipe"`
	Poin           int       `db:"poin"`
	Urutan         int       `db:"urutan"`
	Pilihan        []*Option
}
package entity

type Option struct {
	JawabanId    int    `db:"jawaban_id"`
	PertanyaanId int    `db:"pertanyaan_id"`
	TeksJawaban  string `db:"teks_jawaban"`
	Score        int    `db:"score"`
}

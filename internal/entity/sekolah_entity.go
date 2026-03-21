package entity

type Sekolah struct {
	SekolahId		string `db:"sekolah_id"`
	NamaSekolah		string `db:"nama_sekolah"`
	AlamatSekolah	string `db:"alamat_sekolah"`
	Gurus			[]Guru	`db:"guru"`
	

}

package entity

type Guru struct {
	GuruId		string `db:"guru_id"`
	NIP			string `db:"nip"`
	BidangAjar	string `db:"bidang_ajar"`
	Sekolah		Sekolah `db:"sekolah"`
	Groups		[]Group	`db:"groups"`
}
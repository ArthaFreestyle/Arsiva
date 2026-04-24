package entity

type Guru struct {
	GuruId		string `db:"guru_id" json:"GuruId"`
	NIP			string `db:"nip" json:"NIP"`
	BidangAjar	string `db:"bidang_ajar" json:"BidangAjar"`
	Username    string `json:"Username"`
	Sekolah		Sekolah `db:"sekolah"`
	Groups		[]Group	`db:"groups"`
}
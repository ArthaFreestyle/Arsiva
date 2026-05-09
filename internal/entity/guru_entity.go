package entity

type Guru struct {
	GuruId     string  `db:"guru_id"`
	UserId     string  `db:"user_id"`
	SekolahId  string  `db:"sekolah_id"`
	NIP        string  `db:"nip"`
	BidangAjar string  `db:"bidang_ajar"`
	Username   string  `db:"username"`
	Email      string  `db:"email"`
	Sekolah    Sekolah
	Groups     []Group
}
package entity

type JenisKelamin string

const (
	JenisKelaminL       JenisKelamin = "L"
	JenisKelaminP       JenisKelamin = "P"
	JenisKelaminLainnya JenisKelamin = "Lainnya"
)

type Member struct {
	MemberId     string       `db:"member_id"`
	UserId       string       `db:"user_id"`
	SekolahId    string       `db:"sekolah_id"`
	NIS          string       `db:"nis"`
	TotalXP      int          `db:"total_xp"`
	Level        int          `db:"level"`
	FotoProfil   string       `db:"foto_profil"`
	Bio          string       `db:"bio"`
	TanggalLahir string       `db:"tanggal_lahir"`
	JenisKelamin JenisKelamin `db:"jenis_kelamin"`
	Minat        string       `db:"minat"`
	LastActive   string       `db:"last_active"`
	Username     string       `db:"username"`
	Email        string       `db:"email"`
	Sekolah      Sekolah
}

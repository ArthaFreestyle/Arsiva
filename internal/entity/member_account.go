package entity

// MemberAccount represents an active member account, with an optional profile.
type MemberAccount struct {
	UserId       string  `db:"user_id"`
	HasProfile   bool    `db:"has_profile"`
	MemberId     *string `db:"member_id"`
	SekolahId    *string `db:"sekolah_id"`
	NIS          *string `db:"nis"`
	TotalXP      *int    `db:"total_xp"`
	Level        *int    `db:"level"`
	FotoProfil   *string `db:"foto_profil"`
	Bio          *string `db:"bio"`
	TanggalLahir *string `db:"tanggal_lahir"`
	JenisKelamin *string `db:"jenis_kelamin"`
	Minat        *string `db:"minat"`
	Username     string  `db:"username"`
	Email        string  `db:"email"`
}

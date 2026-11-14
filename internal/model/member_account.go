package model

// MemberAccountResponse represents an active member account, with an optional profile.
type MemberAccountResponse struct {
	UserId       string  `json:"user_id"`
	HasProfile   bool    `json:"has_profile"`
	MemberId     *string `json:"member_id"`
	SekolahId    *string `json:"sekolah_id"`
	NIS          *string `json:"nis"`
	TotalXP      *int    `json:"total_xp"`
	Level        *int    `json:"level"`
	FotoProfil   *string `json:"foto_profil"`
	Bio          *string `json:"bio"`
	TanggalLahir *string `json:"tanggal_lahir"`
	JenisKelamin *string `json:"jenis_kelamin"`
	Minat        *string `json:"minat"`
	Username     string  `json:"username"`
	Email        string  `json:"email"`
}

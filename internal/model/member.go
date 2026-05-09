package model

// ==================== Requests ====================

type MemberCreateRequest struct {
	UserId    string `json:"user_id" validate:"required"`
	SekolahId string `json:"sekolah_id"`
	NIS       string `json:"nis"`
}

type MemberUpdateProfileRequest struct {
	SekolahId    string `json:"sekolah_id"`
	NIS          string `json:"nis"`
	FotoProfil   string `json:"foto_profil"`
	Bio          string `json:"bio"`
	TanggalLahir string `json:"tanggal_lahir"`
	JenisKelamin string `json:"jenis_kelamin"`
	Minat        string `json:"minat"`
}

// ==================== Responses ====================

type MemberResponse struct {
	MemberId     string `json:"member_id"`
	SekolahId    string `json:"sekolah_id"`
	NIS          string `json:"nis"`
	TotalXP      int    `json:"total_xp"`
	Level        int    `json:"level"`
	FotoProfil   string `json:"foto_profil"`
	Bio          string `json:"bio"`
	TanggalLahir string `json:"tanggal_lahir"`
	JenisKelamin string `json:"jenis_kelamin"`
	Minat        string `json:"minat"`
	Username     string `json:"username"`
	Email        string `json:"email"`
}

type MemberDetailResponse struct {
	MemberId     string           `json:"member_id"`
	NIS          string           `json:"nis"`
	TotalXP      int              `json:"total_xp"`
	Level        int              `json:"level"`
	FotoProfil   string           `json:"foto_profil"`
	Bio          string           `json:"bio"`
	TanggalLahir string           `json:"tanggal_lahir"`
	JenisKelamin string           `json:"jenis_kelamin"`
	Minat        string           `json:"minat"`
	Username     string           `json:"username"`
	Email        string           `json:"email"`
	Sekolah      *SekolahResponse `json:"sekolah,omitempty"`
}

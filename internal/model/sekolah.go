package model

// ==================== Requests ====================

type SekolahCreateRequest struct {
	NamaSekolah   string `json:"nama_sekolah" validate:"required"`
	AlamatSekolah string `json:"alamat_sekolah" validate:"required"`
}

type SekolahUpdateRequest struct {
	NamaSekolah   string `json:"nama_sekolah" validate:"required"`
	AlamatSekolah string `json:"alamat_sekolah" validate:"required"`
}

// ==================== Responses ====================

type SekolahResponse struct {
	SekolahId     string `json:"sekolah_id"`
	NamaSekolah   string `json:"nama_sekolah"`
	AlamatSekolah string `json:"alamat_sekolah"`
}

type GuruSummary struct {
	GuruId     string `json:"guru_id"`
	NIP        string `json:"nip,omitempty"`
	BidangAjar string `json:"bidang_ajar,omitempty"`
	Username   string `json:"username,omitempty"`
}

type SekolahDetailResponse struct {
	SekolahId     string       `json:"sekolah_id"`
	NamaSekolah   string       `json:"nama_sekolah"`
	AlamatSekolah string       `json:"alamat_sekolah"`
	Gurus         []GuruSummary `json:"gurus"`
}

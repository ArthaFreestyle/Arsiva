package model

// ==================== Requests ====================

type GuruCreateRequest struct {
	UserId     string `json:"user_id" validate:"required"`
	SekolahId  string `json:"sekolah_id"`
	NIP        string `json:"nip" validate:"required"`
	BidangAjar string `json:"bidang_ajar" validate:"required"`
}

type GuruUpdateRequest struct {
	SekolahId  string `json:"sekolah_id"`
	NIP        string `json:"nip"`
	BidangAjar string `json:"bidang_ajar"`
}

// ==================== Responses ====================

type GuruResponse struct {
	GuruId     string `json:"guru_id"`
	NIP        string `json:"nip"`
	BidangAjar string `json:"bidang_ajar"`
	SekolahId  string `json:"sekolah_id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
}

type GroupSummary struct {
	GroupId   string `json:"group_id"`
	GroupName string `json:"group_name"`
}

type GuruDetailResponse struct {
	GuruId     string           `json:"guru_id"`
	NIP        string           `json:"nip"`
	BidangAjar string           `json:"bidang_ajar"`
	Username   string           `json:"username"`
	Email      string           `json:"email"`
	Sekolah    *SekolahResponse `json:"sekolah,omitempty"`
	Groups     []GroupSummary   `json:"groups"`
}

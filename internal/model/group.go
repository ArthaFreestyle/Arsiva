package model

// ==================== Requests ====================

type GroupCreateRequest struct {
	GroupName string `json:"group_name" validate:"required,max=191"`
}

type GroupUpdateRequest struct {
	GroupName        string `json:"group_name" validate:"required,max=191"`
	ThumbnailAssetId *int   `json:"thumbnail_asset_id"`
}

type GroupInviteEmailRequest struct {
	Emails []string `json:"emails" validate:"required,min=1,dive,email"`
}

type GroupJoinRequest struct {
	InviteToken string `json:"invite_token" validate:"required"`
}

// ==================== Responses ====================

type GroupResponse struct {
	GroupId     string            `json:"group_id"`
	GroupName   string            `json:"group_name"`
	Thumbnail   string            `json:"thumbnail,omitempty"`
	CreatedBy   GroupGuruResponse `json:"created_by"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at,omitempty"`
	MemberCount int               `json:"member_count"`
}

type GroupGuruResponse struct {
	GuruId     int    `json:"guru_id"`
	NIP        string `json:"nip,omitempty"`
	BidangAjar string `json:"bidang_ajar,omitempty"`
	Username   string `json:"username,omitempty"`
}

type GroupDetailResponse struct {
	GroupId     string                `json:"group_id"`
	GroupName   string                `json:"group_name"`
	Thumbnail   string                `json:"thumbnail,omitempty"`
	CreatedBy   GroupGuruResponse     `json:"created_by"`
	CreatedAt   string                `json:"created_at"`
	UpdatedAt   string                `json:"updated_at,omitempty"`
	MemberCount int                   `json:"member_count"`
	Members     []GroupMemberResponse `json:"members"`
}

type GroupMemberResponse struct {
	MemberId         int    `json:"member_id"`
	Username         string `json:"username"`
	Email            string `json:"email"`
	NIS              string `json:"nis,omitempty"`
	FotoProfil       string `json:"foto_profil,omitempty"`
	TanggalBergabung string `json:"tanggal_bergabung"`
}

type GroupInviteResponse struct {
	InviteToken string `json:"invite_token"`
	InviteLink  string `json:"invite_link"`
	QRCodeData  string `json:"qr_code_data"`
	ExpiresAt   string `json:"expires_at"`
}

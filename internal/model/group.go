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
	// Emails is the list of student addresses to invite. The FE sends this both when
	// a guru types addresses and when they pick students from a list (resolved to
	// their emails client-side).
	Emails []string `json:"emails" validate:"required,min=1,max=50,dive,email"`
	// Message is an optional personal note from the guru, shown in the invite email.
	Message string `json:"message" validate:"omitempty,max=500"`
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

// GroupInviteEmailResult is the per-address outcome of an email invitation.
type GroupInviteEmailResult struct {
	Email  string `json:"email"`
	Status string `json:"status"` // "sent" | "failed"
}

// GroupInviteEmailResponse is the confirmation returned to the guru after sending
// invitations: aggregate counts plus a per-email breakdown so the FE can show
// which addresses failed.
type GroupInviteEmailResponse struct {
	Total   int                      `json:"total"`
	Sent    int                      `json:"sent"`
	Failed  int                      `json:"failed"`
	Results []GroupInviteEmailResult `json:"results"`
}

// ==================== Group Contents ====================

type GroupContentCreateRequest struct {
	ContentType string `json:"content_type" validate:"required,oneof=kuis cerita puzzle"`
	ContentId   int    `json:"content_id" validate:"required,min=1"`
}

type GroupContentResponse struct {
	GroupContentId int    `json:"group_content_id"`
	ContentType    string `json:"content_type"`
	ContentId      int    `json:"content_id"`
	Judul          string `json:"judul"`
	Thumbnail      string `json:"thumbnail,omitempty"`
}

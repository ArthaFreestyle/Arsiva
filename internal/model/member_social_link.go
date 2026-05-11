package model

// ==================== Requests ====================

type MemberSocialLinkCreateRequest struct {
	Platform string `json:"platform" validate:"required,oneof=Instagram X TikTok Facebook YouTube"`
	URL      string `json:"url"      validate:"required,url,max=255"`
}

type MemberSocialLinkUpdateRequest struct {
	Platform string `json:"platform" validate:"required,oneof=Instagram X TikTok Facebook YouTube"`
	URL      string `json:"url"      validate:"required,url,max=255"`
}

// ==================== Response ====================

type MemberSocialLinkResponse struct {
	SocialId  string `json:"social_id"`
	Platform  string `json:"platform"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

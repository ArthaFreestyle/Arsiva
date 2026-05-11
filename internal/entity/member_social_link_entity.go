package entity

type Platform string

const (
	PlatformInstagram Platform = "Instagram"
	PlatformX         Platform = "X"
	PlatformTikTok    Platform = "TikTok"
	PlatformFacebook  Platform = "Facebook"
	PlatformYouTube   Platform = "YouTube"
)

type MemberSocialLink struct {
	SocialId  string   `db:"social_id"`
	MemberId  string   `db:"member_id"`
	Platform  Platform `db:"platform"`
	URL       string   `db:"url"`
	CreatedAt string   `db:"created_at"`
}

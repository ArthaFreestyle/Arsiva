package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToMemberSocialLinkResponse(link *entity.MemberSocialLink) *model.MemberSocialLinkResponse {
	if link == nil {
		return nil
	}
	return &model.MemberSocialLinkResponse{
		SocialId:  link.SocialId,
		Platform:  string(link.Platform),
		URL:       link.URL,
		CreatedAt: link.CreatedAt,
	}
}

func ToMemberSocialLinkResponses(links []*entity.MemberSocialLink) []*model.MemberSocialLinkResponse {
	responses := make([]*model.MemberSocialLinkResponse, len(links))
	for i, l := range links {
		responses[i] = ToMemberSocialLinkResponse(l)
	}
	return responses
}

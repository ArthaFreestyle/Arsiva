package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToMemberProgressFinalizeResponse(p *entity.MemberProgress) *model.ProgressFinalizeResponse {
	if p == nil {
		return nil
	}
	return &model.ProgressFinalizeResponse{
		ProgresId: p.ProgresId,
	}
}

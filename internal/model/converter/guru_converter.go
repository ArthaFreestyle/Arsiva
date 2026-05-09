package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToGuruResponse(guru *entity.Guru) *model.GuruResponse {
	if guru == nil {
		return nil
	}
	return &model.GuruResponse{
		GuruId:     guru.GuruId,
		NIP:        guru.NIP,
		BidangAjar: guru.BidangAjar,
		SekolahId:  guru.SekolahId,
		Username:   guru.Username,
		Email:      guru.Email,
	}
}

func ToGuruResponses(gurus []*entity.Guru) []*model.GuruResponse {
	var responses []*model.GuruResponse
	for _, g := range gurus {
		responses = append(responses, ToGuruResponse(g))
	}
	return responses
}

func ToGuruDetailResponse(guru *entity.Guru, sekolah *entity.Sekolah, groups []*entity.Group) *model.GuruDetailResponse {
	if guru == nil {
		return nil
	}

	groupSummaries := []model.GroupSummary{}
	for _, g := range groups {
		groupSummaries = append(groupSummaries, model.GroupSummary{
			GroupId:   g.GroupId,
			GroupName: g.GroupName,
		})
	}

	return &model.GuruDetailResponse{
		GuruId:     guru.GuruId,
		NIP:        guru.NIP,
		BidangAjar: guru.BidangAjar,
		Username:   guru.Username,
		Email:      guru.Email,
		Sekolah:    ToSekolahResponse(sekolah),
		Groups:     groupSummaries,
	}
}

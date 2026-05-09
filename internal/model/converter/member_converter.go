package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToMemberResponse(member *entity.Member) *model.MemberResponse {
	if member == nil {
		return nil
	}
	return &model.MemberResponse{
		MemberId:     member.MemberId,
		SekolahId:    member.SekolahId,
		NIS:          member.NIS,
		TotalXP:      member.TotalXP,
		Level:        member.Level,
		FotoProfil:   member.FotoProfil,
		Bio:          member.Bio,
		TanggalLahir: member.TanggalLahir,
		JenisKelamin: string(member.JenisKelamin),
		Minat:        member.Minat,
		Username:     member.Username,
		Email:        member.Email,
	}
}

func ToMemberResponses(members []*entity.Member) []*model.MemberResponse {
	var responses []*model.MemberResponse
	for _, m := range members {
		responses = append(responses, ToMemberResponse(m))
	}
	return responses
}

func ToMemberDetailResponse(member *entity.Member, sekolah *entity.Sekolah) *model.MemberDetailResponse {
	if member == nil {
		return nil
	}
	return &model.MemberDetailResponse{
		MemberId:     member.MemberId,
		NIS:          member.NIS,
		TotalXP:      member.TotalXP,
		Level:        member.Level,
		FotoProfil:   member.FotoProfil,
		Bio:          member.Bio,
		TanggalLahir: member.TanggalLahir,
		JenisKelamin: string(member.JenisKelamin),
		Minat:        member.Minat,
		Username:     member.Username,
		Email:        member.Email,
		Sekolah:      ToSekolahResponse(sekolah),
	}
}

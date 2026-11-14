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

func ToMemberProfileResponse(
	member *entity.Member,
	sekolah *entity.Sekolah,
	achievements []*entity.MemberAchievement,
	socialLinks []*entity.MemberSocialLink,
) *model.MemberProfileResponse {
	if member == nil {
		return nil
	}

	achievementResponses := make([]*model.MemberAchievementResponse, len(achievements))
	for i, a := range achievements {
		achievementResponses[i] = &model.MemberAchievementResponse{
			AchievementId: a.AchievementId,
			Nama:          a.Nama,
			Deskripsi:     a.Deskripsi,
			BadgeIcon:     a.BadgeIcon,
			XPRequired:    a.XPRequired,
			Tier:          a.Tier,
			UnlockedAt:    a.UnlockedAt,
		}
	}

	socialLinkResponses := make([]*model.MemberSocialLinkResponse, len(socialLinks))
	for i, l := range socialLinks {
		socialLinkResponses[i] = &model.MemberSocialLinkResponse{
			SocialId:  l.SocialId,
			Platform:  string(l.Platform),
			URL:       l.URL,
			CreatedAt: l.CreatedAt,
		}
	}

	return &model.MemberProfileResponse{
		MemberId:     member.MemberId,
		Username:     member.Username,
		Email:        member.Email,
		NIS:          member.NIS,
		FotoProfil:   member.FotoProfil,
		Bio:          member.Bio,
		TanggalLahir: member.TanggalLahir,
		JenisKelamin: string(member.JenisKelamin),
		Minat:        member.Minat,
		LastActive:   member.LastActive,
		TotalXP:      member.TotalXP,
		Level:        member.Level,
		Sekolah:      ToSekolahResponse(sekolah),
		Achievements: achievementResponses,
		SocialLinks:  socialLinkResponses,
	}
}

func ToMemberAccountResponses(accounts []*entity.MemberAccount) []*model.MemberAccountResponse {
	result := make([]*model.MemberAccountResponse, len(accounts))
	for i, account := range accounts {
		result[i] = &model.MemberAccountResponse{
			UserId:       account.UserId,
			HasProfile:   account.HasProfile,
			MemberId:     account.MemberId,
			SekolahId:    account.SekolahId,
			NIS:          account.NIS,
			TotalXP:      account.TotalXP,
			Level:        account.Level,
			FotoProfil:   account.FotoProfil,
			Bio:          account.Bio,
			TanggalLahir: account.TanggalLahir,
			JenisKelamin: account.JenisKelamin,
			Minat:        account.Minat,
			Username:     account.Username,
			Email:        account.Email,
		}
	}
	return result
}

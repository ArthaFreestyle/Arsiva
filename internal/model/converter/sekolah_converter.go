package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToSekolahResponse(sekolah *entity.Sekolah) *model.SekolahResponse {
	if sekolah == nil {
		return nil
	}
	return &model.SekolahResponse{
		SekolahId:     sekolah.SekolahId,
		NamaSekolah:   sekolah.NamaSekolah,
		AlamatSekolah: sekolah.AlamatSekolah,
	}
}

func ToSekolahResponses(sekolahs []*entity.Sekolah) []*model.SekolahResponse {
	var responses []*model.SekolahResponse
	for _, s := range sekolahs {
		responses = append(responses, ToSekolahResponse(s))
	}
	return responses
}

func ToGuruSummary(guru *entity.Guru) model.GuruSummary {
	if guru == nil {
		return model.GuruSummary{}
	}
	return model.GuruSummary{
		GuruId:     guru.GuruId,
		NIP:        guru.NIP,
		BidangAjar: guru.BidangAjar,
		Username:   guru.Username,
	}
}

func ToSekolahDetailResponse(sekolah *entity.Sekolah, gurus []*entity.Guru) *model.SekolahDetailResponse {
	if sekolah == nil {
		return nil
	}

	guruSummaries := []model.GuruSummary{}
	for _, g := range gurus {
		guruSummaries = append(guruSummaries, ToGuruSummary(g))
	}

	return &model.SekolahDetailResponse{
		SekolahId:     sekolah.SekolahId,
		NamaSekolah:   sekolah.NamaSekolah,
		AlamatSekolah: sekolah.AlamatSekolah,
		Gurus:         guruSummaries,
	}
}

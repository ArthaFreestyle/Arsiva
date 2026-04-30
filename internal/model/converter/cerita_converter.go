package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToSceneResponse(scene *entity.Scene) *model.SceneResponse {
	choices := make([]model.SceneChoice, 0, len(scene.SceneChoices))
	for _, c := range scene.SceneChoices {
		text, _ := c["text"].(string)
		next, _ := c["next"].(string)
		choices = append(choices, model.SceneChoice{
			Text: text,
			Next: next,
		})
	}

	return &model.SceneResponse{
		SceneId:      scene.SceneId,
		CeritaId:     scene.CeritaId,
		SceneKey:     scene.SceneKey,
		SceneImage:   toAsset(scene.SceneImageAssetId, scene.SceneImage),
		SceneText:    scene.SceneText,
		SceneChoices: choices,
		IsEnding:     scene.IsEnding,
		EndingPoint:  scene.EndingPoint,
		EndingType:   scene.EndingType,
		Urutan:       scene.Urutan,
	}
}

func ToCeritaResponse(cerita *entity.CeritaInteraktif) *model.CeritaResponse {
	var scenes []*model.SceneResponse
	for _, s := range cerita.Scenes {
		scenes = append(scenes, ToSceneResponse(s))
	}

	return &model.CeritaResponse{
		CeritaId:    cerita.CeritaId,
		Judul:       cerita.Judul,
		Thumbnail:   toAsset(cerita.ThumbnailAssetId, cerita.Thumbnail),
		Deskripsi:   cerita.Deskripsi,
		KategoriId:  cerita.KategoriId,
		XpReward:    cerita.XpReward,
		Scenes:      scenes,
		CreatedAt:   cerita.CreatedAt,
		CreatedBy:   *ToUserResponse(&cerita.CreatedBy),
		IsPublished: cerita.IsPublished,
	}
}

func ToCeritaResponses(ceritas []*entity.CeritaInteraktif) []*model.CeritaResponse {
	responses := make([]*model.CeritaResponse, len(ceritas))
	for i, cerita := range ceritas {
		responses[i] = ToCeritaResponse(cerita)
	}
	return responses
}

package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToStoryCategoryResponse(storyCategory *entity.StoryCategory) *model.StoryCategoryResponse {
	return &model.StoryCategoryResponse{
		StoryCategoryId: storyCategory.StoryCategoryId,
		NamaKategori:    storyCategory.NamaKategori,
	}
}

func ToStoryCategoriesResponse(storyCategories []*entity.StoryCategory) []*model.StoryCategoryResponse {
	storyCategoriesResponse := make([]*model.StoryCategoryResponse, len(storyCategories))
	for i, storyCategory := range storyCategories {
		storyCategoriesResponse[i] = ToStoryCategoryResponse(storyCategory)
	}
	return storyCategoriesResponse
}

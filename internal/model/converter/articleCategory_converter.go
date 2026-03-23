package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToArticleCategoryResponse(articleCategory *entity.ArticleCategory) *model.ArticleCategoryResponse {
	return &model.ArticleCategoryResponse{
		ArticleCategoryId: articleCategory.ArticleCategoryId,
		NamaKategori: articleCategory.NamaKategori,
	}
}

func ToArticleCategoriesResponse(articleCategories []*entity.ArticleCategory) []*model.ArticleCategoryResponse {
	articleCategoriesResponse := make([]*model.ArticleCategoryResponse, len(articleCategories))
	for i, articleCategory := range articleCategories {
		articleCategoriesResponse[i] = ToArticleCategoryResponse(articleCategory)
	}
	return articleCategoriesResponse
}
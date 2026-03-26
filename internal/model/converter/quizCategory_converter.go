package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToQuizCategoryResponse(quizCategory *entity.QuizCategory) *model.QuizCategoryResponse {
	return &model.QuizCategoryResponse{
		QuizCategoryId: quizCategory.QuizCategoryId,
		NamaKategori:   quizCategory.NamaKategori,
		CreatedAt:      quizCategory.CreatedAt,
		CreatedBy:      quizCategory.CreatedBy,
		Deskripsi:      quizCategory.Deskripsi,
	}
}

func ToQuizCategoriesResponse(quizCategories []*entity.QuizCategory) []*model.QuizCategoryResponse {
	quizCategoriesResponse := make([]*model.QuizCategoryResponse, len(quizCategories))
	for i, quizCategory := range quizCategories {
		quizCategoriesResponse[i] = ToQuizCategoryResponse(quizCategory)
	}
	return quizCategoriesResponse
}

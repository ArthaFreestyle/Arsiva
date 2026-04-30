package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToQuizResponse(quiz *entity.Quiz) *model.QuizResponse {
	var soal []*model.QuestionResponse
	for _, q := range quiz.Soal {
		soal = append(soal, ToQuestionResponse(q))
	}

	return &model.QuizResponse{
		QuizId:      quiz.QuizId,
		Judul:       quiz.Judul,
		Gambar:      toAsset(quiz.GambarAssetId, quiz.Gambar),
		Thumbnail:   toAsset(quiz.ThumbnailAssetId, quiz.Thumbnail),
		Kategori:    quiz.Kategori,
		XpReward:    quiz.XpReward,
		Soal:        soal,
		CreatedAt:   quiz.CreatedAt,
		CreatedBy:   *ToUserResponse(&quiz.CreatedBy),
		IsPublished: quiz.IsPublished,
	}
}

func ToQuizResponses(quizzes []*entity.Quiz) []*model.QuizResponse {
	responses := make([]*model.QuizResponse, len(quizzes))
	for i, quiz := range quizzes {
		responses[i] = ToQuizResponse(quiz)
	}
	return responses
}

func ToQuestionResponse(q *entity.Question) *model.QuestionResponse {
	var pilihan []*model.OptionResponse
	for _, o := range q.Pilihan {
		pilihan = append(pilihan, ToOptionResponse(o))
	}
	return &model.QuestionResponse{
		PertanyaanId:   q.PertanyaanId,
		KuisId:         q.KuisId,
		TeksPertanyaan: q.TeksPertanyaan,
		Image:          toAsset(q.ImageAssetId, q.Image),
		Tipe:           q.Tipe,
		Poin:           q.Poin,
		Urutan:         q.Urutan,
		Pilihan:        pilihan,
	}
}

func ToOptionResponse(o *entity.Option) *model.OptionResponse {
	return &model.OptionResponse{
		JawabanId:   o.JawabanId,
		TeksJawaban: o.TeksJawaban,
		Score:       o.Score,
	}
}
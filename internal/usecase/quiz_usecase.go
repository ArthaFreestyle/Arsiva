package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type QuizUseCase interface {
	GetAll(ctx context.Context, page int, size int, search string) ([]*model.QuizResponse, int, error)
	GetByID(ctx context.Context, id int) (*model.QuizResponse, error)
	Create(ctx context.Context, quiz *model.QuizRequest, userId string) (*model.QuizResponse, error)
	Update(ctx context.Context, quiz *model.QuizRequest, id int) (*model.QuizResponse, error)
	Delete(ctx context.Context, id int) error
}

type quizUseCaseImpl struct {
	QuizRepository repository.QuizRepository
	Log            *logrus.Logger
	Validator      *validator.Validate
}

func NewQuizUseCase(quizRepository repository.QuizRepository, log *logrus.Logger, validator *validator.Validate) QuizUseCase {
	return &quizUseCaseImpl{
		QuizRepository: quizRepository,
		Log:            log,
		Validator:      validator,
	}
}

func (u *quizUseCaseImpl) GetAll(ctx context.Context, page int, size int, search string) ([]*model.QuizResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	quizzes, total, err := u.QuizRepository.GetAll(ctx, page, size, search)
	if err != nil {
		u.Log.Warnf("error when get all quiz: %v", err)
		return nil, 0, fiber.ErrInternalServerError
	}

	return converter.ToQuizResponses(quizzes), total, nil
}

func (u *quizUseCaseImpl) GetByID(ctx context.Context, id int) (*model.QuizResponse, error) {
	quiz, err := u.QuizRepository.GetByID(ctx, id)
	if err != nil {
		u.Log.Warnf("error when get quiz by id: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToQuizResponse(quiz), nil
}

func (u *quizUseCaseImpl) Create(ctx context.Context, quiz *model.QuizRequest, userId string) (*model.QuizResponse, error) {
	err := u.Validator.Struct(quiz)
	if err != nil {
		u.Log.Warnf("error when validate quiz: %v", err)
		return nil, fiber.ErrBadRequest
	}

	// Map request to entity
	quizEntity := &entity.Quiz{
		Judul:       quiz.Judul,
		Gambar:      quiz.Gambar,
		Thumbnail:   quiz.Thumbnail,
		KategoriId:  quiz.KategoriId,
		XpReward:    quiz.XpReward,
		IsPublished: quiz.IsPublished,
		CreatedBy: entity.User{
			UserId: userId,
		},
	}

	// Map questions + options
	for _, q := range quiz.Soal {
		question := &entity.Question{
			TeksPertanyaan: q.TeksPertanyaan,
			Image:          q.Image,
			Tipe:           q.Tipe,
			Poin:           q.Poin,
			Urutan:         q.Urutan,
		}
		for _, o := range q.Pilihan {
			question.Pilihan = append(question.Pilihan, &entity.Option{
				TeksJawaban: o.TeksJawaban,
				Score:       o.Score,
			})
		}
		quizEntity.Soal = append(quizEntity.Soal, question)
	}

	createdQuiz, err := u.QuizRepository.Create(ctx, quizEntity)
	if err != nil {
		u.Log.Warnf("error when create quiz: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToQuizResponse(createdQuiz), nil
}

func (u *quizUseCaseImpl) Update(ctx context.Context, quiz *model.QuizRequest, id int) (*model.QuizResponse, error) {
	err := u.Validator.Struct(quiz)
	if err != nil {
		u.Log.Warnf("error when validate quiz: %v", err)
		return nil, fiber.ErrBadRequest
	}

	quizEntity := &entity.Quiz{
		QuizId:      id,
		Judul:       quiz.Judul,
		Gambar:      quiz.Gambar,
		Thumbnail:   quiz.Thumbnail,
		KategoriId:  quiz.KategoriId,
		XpReward:    quiz.XpReward,
		IsPublished: quiz.IsPublished,
	}

	for _, q := range quiz.Soal {
		question := &entity.Question{
			TeksPertanyaan: q.TeksPertanyaan,
			Image:          q.Image,
			Tipe:           q.Tipe,
			Poin:           q.Poin,
			Urutan:         q.Urutan,
		}
		for _, o := range q.Pilihan {
			question.Pilihan = append(question.Pilihan, &entity.Option{
				TeksJawaban: o.TeksJawaban,
				Score:       o.Score,
			})
		}
		quizEntity.Soal = append(quizEntity.Soal, question)
	}

	updatedQuiz, err := u.QuizRepository.Update(ctx, quizEntity)
	if err != nil {
		u.Log.Warnf("error when update quiz: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToQuizResponse(updatedQuiz), nil
}

func (u *quizUseCaseImpl) Delete(ctx context.Context, id int) error {
	err := u.QuizRepository.Delete(ctx, id)
	if err != nil {
		u.Log.Warnf("error when delete quiz: %v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}
package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type QuizCategoryUseCase interface {
	GetAllQuizCategories(ctx context.Context, page int, size int, search string) ([]*model.QuizCategoryResponse, int, error)
	GetQuizCategoryById(ctx context.Context, quizCategoryId string) (*model.QuizCategoryResponse, error)
	CreateQuizCategory(ctx context.Context, quizCategory *model.QuizCategoryRequest, userId string) (*model.QuizCategoryResponse, error)
	UpdateQuizCategory(ctx context.Context, quizCategory *model.QuizCategoryRequest, quizCategoryId string) (*model.QuizCategoryResponse, error)
	DeleteQuizCategory(ctx context.Context, quizCategoryId string) error
}

type QuizCategoryUseCaseImpl struct {
	QuizCategoryRepository repository.QuizCategoryRepository
	Log                    *logrus.Logger
	Validate               *validator.Validate
}

func NewQuizCategoryUseCase(quizCategoryRepository repository.QuizCategoryRepository, log *logrus.Logger, validate *validator.Validate) QuizCategoryUseCase {
	return &QuizCategoryUseCaseImpl{
		QuizCategoryRepository: quizCategoryRepository,
		Log:                    log,
		Validate:               validate,
	}
}

func (u *QuizCategoryUseCaseImpl) GetAllQuizCategories(ctx context.Context, page int, size int, search string) ([]*model.QuizCategoryResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	quizCategories, total, err := u.QuizCategoryRepository.GetAllQuizCategories(ctx, page, size, search)
	if err != nil {
		u.Log.Warnf("error when get all quiz categories: %v", err)
		return nil, 0, fiber.ErrInternalServerError
	}

	response := converter.ToQuizCategoriesResponse(quizCategories)
	return response, total, nil
}

func (u *QuizCategoryUseCaseImpl) GetQuizCategoryById(ctx context.Context, quizCategoryId string) (*model.QuizCategoryResponse, error) {
	quizCategory, err := u.QuizCategoryRepository.GetQuizCategoryById(ctx, quizCategoryId)
	if err != nil {
		u.Log.Warnf("error when get quiz category by id: %v", err)
		return nil, fiber.ErrInternalServerError
	}
	response := converter.ToQuizCategoryResponse(quizCategory)
	return response, nil
}

func (u *QuizCategoryUseCaseImpl) CreateQuizCategory(ctx context.Context, quizCategory *model.QuizCategoryRequest, userId string) (*model.QuizCategoryResponse, error) {
	err := u.Validate.Struct(quizCategory)
	if err != nil {
		return nil, fiber.ErrBadRequest
	}
	
	now := time.Now()
	var deskripsiPtr *string
	if quizCategory.Deskripsi != "" {
		deskripsiPtr = &quizCategory.Deskripsi
	}
	
	quizCategoryEntity := &entity.QuizCategory{
		NamaKategori: quizCategory.NamaKategori,
		Deskripsi:    deskripsiPtr,
		CreatedAt:    &now,
		CreatedBy:    userId,
	}
	category, err := u.QuizCategoryRepository.CreateQuizCategory(ctx, quizCategoryEntity)
	if err != nil {
		u.Log.Warnf("error when create quiz category: %v", err)
		return nil, fiber.ErrInternalServerError
	}
	response := converter.ToQuizCategoryResponse(category)
	return response, nil
}

func (u *QuizCategoryUseCaseImpl) UpdateQuizCategory(ctx context.Context, quizCategory *model.QuizCategoryRequest, quizCategoryId string) (*model.QuizCategoryResponse, error) {
	err := u.Validate.Struct(quizCategory)
	if err != nil {
		return nil, fiber.ErrBadRequest
	}
	
	var deskripsiPtr *string
	if quizCategory.Deskripsi != "" {
		deskripsiPtr = &quizCategory.Deskripsi
	}
	
	quizCategoryEntity := &entity.QuizCategory{
		QuizCategoryId: quizCategoryId,
		NamaKategori:   quizCategory.NamaKategori,
		Deskripsi:      deskripsiPtr,
	}
	category, err := u.QuizCategoryRepository.UpdateQuizCategory(ctx, quizCategoryEntity)
	if err != nil {
		u.Log.Warnf("error when update quiz category: %v", err)
		return nil, fiber.ErrInternalServerError
	}
	response := converter.ToQuizCategoryResponse(category)
	return response, nil
}

func (u *QuizCategoryUseCaseImpl) DeleteQuizCategory(ctx context.Context, quizCategoryId string) error {
	err := u.QuizCategoryRepository.DeleteQuizCategory(ctx, quizCategoryId)
	if err != nil {
		u.Log.Warnf("error when delete quiz category: %v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

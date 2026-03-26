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

type StoryCategoryUseCase interface {
	GetAllStoryCategories(ctx context.Context, page int, size int, search string) ([]*model.StoryCategoryResponse, int, error)
	GetStoryCategoryById(ctx context.Context, storyCategoryId string) (*model.StoryCategoryResponse, error)
	CreateStoryCategory(ctx context.Context, storyCategory *model.StoryCategoryRequest) (*model.StoryCategoryResponse, error)
	UpdateStoryCategory(ctx context.Context, storyCategory *model.StoryCategoryRequest, storyCategoryId string) (*model.StoryCategoryResponse, error)
	DeleteStoryCategory(ctx context.Context, storyCategoryId string) error
}

type StoryCategoryUseCaseImpl struct {
	StoryCategoryRepository repository.StoryCategoryRepository
	Log                     *logrus.Logger
	Validate                *validator.Validate
}

func NewStoryCategoryUseCase(storyCategoryRepository repository.StoryCategoryRepository, log *logrus.Logger, validate *validator.Validate) StoryCategoryUseCase {
	return &StoryCategoryUseCaseImpl{
		StoryCategoryRepository: storyCategoryRepository,
		Log:                     log,
		Validate:                validate,
	}
}

func (u *StoryCategoryUseCaseImpl) GetAllStoryCategories(ctx context.Context, page int, size int, search string) ([]*model.StoryCategoryResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	storyCategories, total, err := u.StoryCategoryRepository.GetAllStoryCategories(ctx, page, size, search)
	if err != nil {
		u.Log.Warnf("error when get all story categories: %v", err)
		return nil, 0, fiber.ErrInternalServerError
	}

	response := converter.ToStoryCategoriesResponse(storyCategories)
	return response, total, nil
}

func (u *StoryCategoryUseCaseImpl) GetStoryCategoryById(ctx context.Context, storyCategoryId string) (*model.StoryCategoryResponse, error) {
	storyCategory, err := u.StoryCategoryRepository.GetStoryCategoryById(ctx, storyCategoryId)
	if err != nil {
		u.Log.Warnf("error when get story category by id: %v", err)
		return nil, fiber.ErrInternalServerError
	}
	response := converter.ToStoryCategoryResponse(storyCategory)
	return response, nil
}

func (u *StoryCategoryUseCaseImpl) CreateStoryCategory(ctx context.Context, storyCategory *model.StoryCategoryRequest) (*model.StoryCategoryResponse, error) {
	err := u.Validate.Struct(storyCategory)
	if err != nil {
		return nil, fiber.ErrBadRequest
	}
	storyCategoryEntity := &entity.StoryCategory{
		NamaKategori: storyCategory.NamaKategori,
	}
	category, err := u.StoryCategoryRepository.CreateStoryCategory(ctx, storyCategoryEntity)
	if err != nil {
		u.Log.Warnf("error when create story category: %v", err)
		return nil, fiber.ErrInternalServerError
	}
	response := converter.ToStoryCategoryResponse(category)
	return response, nil
}

func (u *StoryCategoryUseCaseImpl) UpdateStoryCategory(ctx context.Context, storyCategory *model.StoryCategoryRequest, storyCategoryId string) (*model.StoryCategoryResponse, error) {
	err := u.Validate.Struct(storyCategory)
	if err != nil {
		return nil, fiber.ErrBadRequest
	}
	storyCategoryEntity := &entity.StoryCategory{
		StoryCategoryId: storyCategoryId,
		NamaKategori:    storyCategory.NamaKategori,
	}
	category, err := u.StoryCategoryRepository.UpdateStoryCategory(ctx, storyCategoryEntity)
	if err != nil {
		u.Log.Warnf("error when update story category: %v", err)
		return nil, fiber.ErrInternalServerError
	}
	response := converter.ToStoryCategoryResponse(category)
	return response, nil
}

func (u *StoryCategoryUseCaseImpl) DeleteStoryCategory(ctx context.Context, storyCategoryId string) error {
	err := u.StoryCategoryRepository.DeleteStoryCategory(ctx, storyCategoryId)
	if err != nil {
		u.Log.Warnf("error when delete story category: %v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

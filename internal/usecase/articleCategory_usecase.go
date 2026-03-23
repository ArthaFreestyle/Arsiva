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

type ArticleCategoryUseCase interface {
	GetAllArticleCategories(ctx context.Context) ([]*model.ArticleCategoryResponse, error)
	GetArticleCategoryById(ctx context.Context,ArticleCategoryId string) (*model.ArticleCategoryResponse, error)
	CreateArticleCategory(ctx context.Context,articleCategory *model.ArticleCategoryRequest) (*model.ArticleCategoryResponse, error)
	UpdateArticleCategory(ctx context.Context,articleCategory *model.ArticleCategoryRequest,ArticleCategoryId string) (*model.ArticleCategoryResponse, error)
	DeleteArticleCategory(ctx context.Context,ArticleCategoryId string) (error)
}

type ArticleCategoryUseCaseImpl struct {
	ArticleCategoryRepository repository.ArticleCategoryRepository
	Log *logrus.Logger
	Validate *validator.Validate
}

func NewArticleCategoryUseCase(articleCategoryRepository repository.ArticleCategoryRepository, log *logrus.Logger, validate *validator.Validate) ArticleCategoryUseCase {
	return &ArticleCategoryUseCaseImpl{
		ArticleCategoryRepository: articleCategoryRepository,
		Log: log,
		Validate: validate,
	}
}

func (u *ArticleCategoryUseCaseImpl) GetAllArticleCategories(ctx context.Context) ([]*model.ArticleCategoryResponse, error) {
	articleCategories,err := u.ArticleCategoryRepository.GetAllArticleCategories(ctx)
	if err != nil {
		return nil,err
	}

	response := converter.ToArticleCategoriesResponse(articleCategories)
	return response,nil
}

func (u *ArticleCategoryUseCaseImpl) GetArticleCategoryById(ctx context.Context,ArticleCategoryId string) (*model.ArticleCategoryResponse, error) {
	articleCategory,err := u.ArticleCategoryRepository.GetArticleCategoryById(ctx,ArticleCategoryId)
	if err != nil {
		return nil,err
	}
	response := converter.ToArticleCategoryResponse(articleCategory)
	return response,nil
}

func (u *ArticleCategoryUseCaseImpl) CreateArticleCategory(ctx context.Context,articleCategory *model.ArticleCategoryRequest) (*model.ArticleCategoryResponse, error) {
	err := u.Validate.Struct(articleCategory)
	if err != nil {
		return nil,fiber.ErrBadRequest
	}
	articleCategoryEntity := &entity.ArticleCategory{
		NamaKategori: articleCategory.NamaKategori,
	}
	category,err := u.ArticleCategoryRepository.CreateArticleCategory(ctx,articleCategoryEntity)
	if err != nil {
		return nil,err
	}
	response := converter.ToArticleCategoryResponse(category)
	return response,nil
}

func (u *ArticleCategoryUseCaseImpl) UpdateArticleCategory(ctx context.Context,articleCategory *model.ArticleCategoryRequest,ArticleCategoryId string) (*model.ArticleCategoryResponse, error) {
	err := u.Validate.Struct(articleCategory)
	if err != nil {
		return nil,fiber.ErrBadRequest
	}
	articleCategoryEntity := &entity.ArticleCategory{
		ArticleCategoryId: ArticleCategoryId,
		NamaKategori: articleCategory.NamaKategori,
	}
	category,err := u.ArticleCategoryRepository.UpdateArticleCategory(ctx,articleCategoryEntity)
	if err != nil {
		return nil,err
	}
	response := converter.ToArticleCategoryResponse(category)
	return response,nil
}

func (u *ArticleCategoryUseCaseImpl) DeleteArticleCategory(ctx context.Context,ArticleCategoryId string) (error) {
	err := u.ArticleCategoryRepository.DeleteArticleCategory(ctx,ArticleCategoryId)
	if err != nil {
		return err
	}
	return nil
}
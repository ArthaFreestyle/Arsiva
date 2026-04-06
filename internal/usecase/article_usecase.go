package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
	"ArthaFreestyle/Arsiva/internal/utils"
	"context"
	"strconv"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type ArticleUseCase interface {
	GetAllArticle(ctx context.Context, page int, size int, search string) ([]*model.ArticleResponse, int, error)
	GetArticleBySlug(ctx context.Context, slug string) (*model.ArticleResponse, error)
	GetArticleById(ctx context.Context, articleId string) (*model.ArticleResponse, error)
	CreateArticle(ctx context.Context, article *model.ArticleCreateRequest,UserId string) (*model.ArticleResponse, error)
	UpdateArticle(ctx context.Context, article *model.ArticleUpdateRequest, articleId string) (*model.ArticleResponse, error)
	DeleteArticle(ctx context.Context, articleId string) (error)
}

type articleUseCaseImpl struct {
	ArticleRepository repository.ArticleRepository
	AssetRepository   repository.AssetRepository
	Log *logrus.Logger
	Validator *validator.Validate
}

func NewArticleUseCase(articleRepository repository.ArticleRepository, assetRepository repository.AssetRepository, log *logrus.Logger,validator *validator.Validate) ArticleUseCase {
	return &articleUseCaseImpl{
		ArticleRepository: articleRepository,
		AssetRepository:   assetRepository,
		Log: log,
		Validator: validator,
	}
}

func (u *articleUseCaseImpl) GetAllArticle(ctx context.Context, page int, size int, search string) ([]*model.ArticleResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	articles,total,err := u.ArticleRepository.GetAllArticle(ctx,page,size,search)
	if err != nil {
		u.Log.Warnf("error when get all article: %v",err)
		return nil,0,fiber.ErrInternalServerError
	}

	res := converter.ToArticleResponses(articles)
	return res,total,nil
}

func (u *articleUseCaseImpl) GetArticleBySlug(ctx context.Context, slug string) (*model.ArticleResponse, error) {
	article,err := u.ArticleRepository.GetArticleBySlug(ctx,slug)
	if err != nil {
		u.Log.Warnf("error when get article by slug: %v",err)
		return nil,fiber.ErrInternalServerError
	}

	res := converter.ToArticleResponse(article)
	return res,nil
}

func (u *articleUseCaseImpl) GetArticleById(ctx context.Context, articleId string) (*model.ArticleResponse, error) {
	article,err := u.ArticleRepository.GetArticleById(ctx,articleId)
	if err != nil {
		u.Log.Warnf("error when get article by id: %v",err)
		return nil,fiber.ErrInternalServerError
	}

	res := converter.ToArticleResponse(article)
	return res,nil
}

func (u *articleUseCaseImpl) CreateArticle(ctx context.Context, article *model.ArticleCreateRequest,userId string) (*model.ArticleResponse, error) {
	err := u.Validator.Struct(article)
	if err != nil {
		u.Log.Warnf("error when validate article: %v",err)
		return nil,fiber.ErrBadRequest
	}

	slug := utils.GenerateSlug(article.Title)
	categoryIdInt, _ := strconv.Atoi(article.CategoryId)

	NewArticle := &entity.Article{
		Judul: article.Title,
		Slug: slug,
		KategoriId: categoryIdInt,
		Kategori: entity.ArticleCategory{
			ArticleCategoryId: article.CategoryId,
		},
		Status: "draft",
		CreatedBy: entity.User{
			UserId: userId,
		},
	}

	createdArticle,err := u.ArticleRepository.CreateArticle(ctx,NewArticle)
	if err != nil {
		u.Log.Warnf("error when create article: %v",err)
		return nil,fiber.ErrInternalServerError
	}

	res := converter.ToArticleResponse(createdArticle)
	return res,nil
}

func (u *articleUseCaseImpl) UpdateArticle(ctx context.Context, article *model.ArticleUpdateRequest, articleId string) (*model.ArticleResponse, error) {
	err := u.Validator.Struct(article)
	if err != nil {
		u.Log.Warnf("error when validate article: %v",err)
		return nil,fiber.ErrBadRequest
	}

	slug := utils.GenerateSlug(article.Title)
	excerpt := utils.GenerateExcerpt(article.Content,150)
	categoryIdInt, _ := strconv.Atoi(article.CategoryId)

	UpdatedArticle := &entity.Article{
		ArticleId: articleId,
		Judul: article.Title,
		Slug: slug,
		Konten: &article.Content,
		Excerpt: &excerpt,
		KategoriId: categoryIdInt,
		Kategori: entity.ArticleCategory{
			ArticleCategoryId: article.CategoryId,
		},
		Status: article.Status,
		ThumbnailAssetId: article.ThumbnailAssetId,
	}

	updatedArticle,err := u.ArticleRepository.UpdateArticle(ctx,UpdatedArticle)
	if err != nil {
		u.Log.Warnf("error when update article: %v",err)
		return nil,fiber.ErrInternalServerError
	}

	if article.ThumbnailAssetId != nil {
		err := u.AssetRepository.MarkAsUsed(ctx, []int{*article.ThumbnailAssetId})
		if err != nil {
			u.Log.Warnf("failed to mark asset as used: %v", err)
		}
	}

	res := converter.ToArticleResponse(updatedArticle)
	return res,nil
}

func (u *articleUseCaseImpl) DeleteArticle(ctx context.Context, articleId string) (error) {
	err := u.ArticleRepository.DeleteArticle(ctx,articleId)
	if err != nil {
		u.Log.Warnf("error when delete article: %v",err)
		return fiber.ErrInternalServerError
	}
	return nil
}
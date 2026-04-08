package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
	"context"
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const articleCategoriesListCacheKey = "arsiva:article_categories:list"

type ArticleCategoryCache interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type ArticleCategoryUseCase interface {
	GetAllArticleCategories(ctx context.Context) ([]*model.ArticleCategoryResponse, error)
	GetArticleCategoryById(ctx context.Context, ArticleCategoryId string) (*model.ArticleCategoryResponse, error)
	CreateArticleCategory(ctx context.Context, articleCategory *model.ArticleCategoryRequest) (*model.ArticleCategoryResponse, error)
	UpdateArticleCategory(ctx context.Context, articleCategory *model.ArticleCategoryRequest, ArticleCategoryId string) (*model.ArticleCategoryResponse, error)
	DeleteArticleCategory(ctx context.Context, ArticleCategoryId string) error
}

type ArticleCategoryUseCaseImpl struct {
	ArticleCategoryRepository repository.ArticleCategoryRepository
	ArticleCategoryCache      ArticleCategoryCache
	Log                       *logrus.Logger
	Validate                  *validator.Validate
}

func NewArticleCategoryUseCase(articleCategoryRepository repository.ArticleCategoryRepository, articleCategoryCache ArticleCategoryCache, log *logrus.Logger, validate *validator.Validate) ArticleCategoryUseCase {
	return &ArticleCategoryUseCaseImpl{
		ArticleCategoryRepository: articleCategoryRepository,
		ArticleCategoryCache:      articleCategoryCache,
		Log:                       log,
		Validate:                  validate,
	}
}

func (u *ArticleCategoryUseCaseImpl) GetAllArticleCategories(ctx context.Context) ([]*model.ArticleCategoryResponse, error) {
	if u.ArticleCategoryCache != nil {
		cachedCategories, err := u.ArticleCategoryCache.Get(ctx, articleCategoriesListCacheKey).Result()
		if err == nil {
			var response []*model.ArticleCategoryResponse
			if unmarshalErr := json.Unmarshal([]byte(cachedCategories), &response); unmarshalErr == nil {
				return response, nil
			} else {
				u.warnf("error when unmarshal article categories from cache: %v", unmarshalErr)
			}
		} else if err != redis.Nil {
			u.warnf("error when get article categories from cache: %v", err)
		}
	}

	articleCategories, err := u.ArticleCategoryRepository.GetAllArticleCategories(ctx)
	if err != nil {
		return nil, err
	}

	response := converter.ToArticleCategoriesResponse(articleCategories)
	if u.ArticleCategoryCache != nil {
		cachedResponse, marshalErr := json.Marshal(response)
		if marshalErr != nil {
			u.warnf("error when marshal article categories for cache: %v", marshalErr)
		} else if cacheErr := u.ArticleCategoryCache.Set(ctx, articleCategoriesListCacheKey, cachedResponse, 24*time.Hour).Err(); cacheErr != nil {
			u.warnf("error when set article categories cache: %v", cacheErr)
		}
	}
	return response, nil
}

func (u *ArticleCategoryUseCaseImpl) GetArticleCategoryById(ctx context.Context, ArticleCategoryId string) (*model.ArticleCategoryResponse, error) {
	articleCategory, err := u.ArticleCategoryRepository.GetArticleCategoryById(ctx, ArticleCategoryId)
	if err != nil {
		return nil, err
	}
	response := converter.ToArticleCategoryResponse(articleCategory)
	return response, nil
}

func (u *ArticleCategoryUseCaseImpl) CreateArticleCategory(ctx context.Context, articleCategory *model.ArticleCategoryRequest) (*model.ArticleCategoryResponse, error) {
	err := u.Validate.Struct(articleCategory)
	if err != nil {
		return nil, fiber.ErrBadRequest
	}
	articleCategoryEntity := &entity.ArticleCategory{
		NamaKategori: articleCategory.NamaKategori,
	}
	category, err := u.ArticleCategoryRepository.CreateArticleCategory(ctx, articleCategoryEntity)
	if err != nil {
		return nil, err
	}
	u.invalidateArticleCategoriesCache(ctx)
	response := converter.ToArticleCategoryResponse(category)
	return response, nil
}

func (u *ArticleCategoryUseCaseImpl) UpdateArticleCategory(ctx context.Context, articleCategory *model.ArticleCategoryRequest, ArticleCategoryId string) (*model.ArticleCategoryResponse, error) {
	err := u.Validate.Struct(articleCategory)
	if err != nil {
		return nil, fiber.ErrBadRequest
	}
	articleCategoryEntity := &entity.ArticleCategory{
		ArticleCategoryId: ArticleCategoryId,
		NamaKategori:      articleCategory.NamaKategori,
	}
	category, err := u.ArticleCategoryRepository.UpdateArticleCategory(ctx, articleCategoryEntity)
	if err != nil {
		return nil, err
	}
	u.invalidateArticleCategoriesCache(ctx)
	response := converter.ToArticleCategoryResponse(category)
	return response, nil
}

func (u *ArticleCategoryUseCaseImpl) DeleteArticleCategory(ctx context.Context, ArticleCategoryId string) error {
	err := u.ArticleCategoryRepository.DeleteArticleCategory(ctx, ArticleCategoryId)
	if err != nil {
		return err
	}
	u.invalidateArticleCategoriesCache(ctx)
	return nil
}

func (u *ArticleCategoryUseCaseImpl) invalidateArticleCategoriesCache(ctx context.Context) {
	if u.ArticleCategoryCache == nil {
		return
	}

	if err := u.ArticleCategoryCache.Del(ctx, articleCategoriesListCacheKey).Err(); err != nil {
		u.warnf("error when invalidate article categories cache: %v", err)
	}
}

func (u *ArticleCategoryUseCaseImpl) warnf(format string, args ...interface{}) {
	if u.Log != nil {
		u.Log.Warnf(format, args...)
	}
}

package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
)

const testArticleCategoriesListCacheKey = "arsiva:article_categories:list"

func TestNewArticleCategoryUseCase(t *testing.T) {
	uc := NewArticleCategoryUseCase(nil, nil, nil, validator.New())
	if uc == nil {
		t.Fatal("expected usecase instance")
	}
}

type articleCategoryRepoStub struct {
	getAllFn  func(ctx context.Context) ([]*entity.ArticleCategory, error)
	getByIDFn func(ctx context.Context, articleCategoryID string) (*entity.ArticleCategory, error)
	createFn  func(ctx context.Context, articleCategory *entity.ArticleCategory) (*entity.ArticleCategory, error)
	updateFn  func(ctx context.Context, articleCategory *entity.ArticleCategory) (*entity.ArticleCategory, error)
	deleteFn  func(ctx context.Context, articleCategoryID string) error
}

func (r *articleCategoryRepoStub) GetAllArticleCategories(ctx context.Context) ([]*entity.ArticleCategory, error) {
	return r.getAllFn(ctx)
}

func (r *articleCategoryRepoStub) GetArticleCategoryById(ctx context.Context, articleCategoryID string) (*entity.ArticleCategory, error) {
	if r.getByIDFn == nil {
		return nil, nil
	}
	return r.getByIDFn(ctx, articleCategoryID)
}

func (r *articleCategoryRepoStub) CreateArticleCategory(ctx context.Context, articleCategory *entity.ArticleCategory) (*entity.ArticleCategory, error) {
	return r.createFn(ctx, articleCategory)
}

func (r *articleCategoryRepoStub) UpdateArticleCategory(ctx context.Context, articleCategory *entity.ArticleCategory) (*entity.ArticleCategory, error) {
	return r.updateFn(ctx, articleCategory)
}

func (r *articleCategoryRepoStub) DeleteArticleCategory(ctx context.Context, articleCategoryID string) error {
	return r.deleteFn(ctx, articleCategoryID)
}

type articleCategoryCacheStub struct {
	getFn func(ctx context.Context, key string) (string, error)
	setFn func(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	delFn func(ctx context.Context, keys ...string) error
}

func (c *articleCategoryCacheStub) Get(ctx context.Context, key string) *redis.StringCmd {
	if c.getFn == nil {
		return redis.NewStringResult("", redis.Nil)
	}
	result, err := c.getFn(ctx, key)
	return redis.NewStringResult(result, err)
}

func (c *articleCategoryCacheStub) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	if c.setFn == nil {
		return redis.NewStatusResult("OK", nil)
	}
	err := c.setFn(ctx, key, value, expiration)
	return redis.NewStatusResult("OK", err)
}

func (c *articleCategoryCacheStub) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	if c.delFn == nil {
		return redis.NewIntResult(1, nil)
	}
	err := c.delFn(ctx, keys...)
	return redis.NewIntResult(1, err)
}

func TestGetAllArticleCategoriesCacheHit(t *testing.T) {
	repoCalled := false
	repo := &articleCategoryRepoStub{
		getAllFn: func(ctx context.Context) ([]*entity.ArticleCategory, error) {
			repoCalled = true
			return nil, nil
		},
	}
	cache := &articleCategoryCacheStub{
		getFn: func(ctx context.Context, key string) (string, error) {
			return `[{"article_category_id":"1","nama_kategori":"Teknologi"}]`, nil
		},
	}

	uc := NewArticleCategoryUseCase(repo, cache, nil, validator.New())
	result, err := uc.GetAllArticleCategories(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repoCalled {
		t.Fatal("expected repository not called on cache hit")
	}
	if len(result) != 1 || result[0].NamaKategori != "Teknologi" {
		t.Fatalf("unexpected cache hit result: %+v", result)
	}
}

func TestGetAllArticleCategoriesCacheMissThenCacheSet(t *testing.T) {
	repoCalled := 0
	setCalled := false
	repo := &articleCategoryRepoStub{
		getAllFn: func(ctx context.Context) ([]*entity.ArticleCategory, error) {
			repoCalled++
			return []*entity.ArticleCategory{
				{ArticleCategoryId: "1", NamaKategori: "Sains"},
			}, nil
		},
	}
	cache := &articleCategoryCacheStub{
		getFn: func(ctx context.Context, key string) (string, error) {
			return "", redis.Nil
		},
		setFn: func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
			if key != testArticleCategoriesListCacheKey {
				t.Fatalf("unexpected cache key: %s", key)
			}
			setCalled = true
			return nil
		},
	}

	uc := NewArticleCategoryUseCase(repo, cache, nil, validator.New())
	result, err := uc.GetAllArticleCategories(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repoCalled != 1 {
		t.Fatalf("expected repository called once, got %d", repoCalled)
	}
	if !setCalled {
		t.Fatal("expected cache set on cache miss")
	}
	if len(result) != 1 || result[0].NamaKategori != "Sains" {
		t.Fatalf("unexpected cache miss result: %+v", result)
	}
}

func TestGetAllArticleCategoriesCacheErrorFallbackToDatabase(t *testing.T) {
	repoCalled := false
	repo := &articleCategoryRepoStub{
		getAllFn: func(ctx context.Context) ([]*entity.ArticleCategory, error) {
			repoCalled = true
			return []*entity.ArticleCategory{
				{ArticleCategoryId: "1", NamaKategori: "Budaya"},
			}, nil
		},
	}
	cache := &articleCategoryCacheStub{
		getFn: func(ctx context.Context, key string) (string, error) {
			return "", errors.New("redis timeout")
		},
	}

	uc := NewArticleCategoryUseCase(repo, cache, nil, validator.New())
	result, err := uc.GetAllArticleCategories(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !repoCalled {
		t.Fatal("expected database fallback when cache errors")
	}
	if len(result) != 1 || result[0].NamaKategori != "Budaya" {
		t.Fatalf("unexpected fallback result: %+v", result)
	}
}

func TestArticleCategoryWriteOperationsInvalidateCache(t *testing.T) {
	delCalled := 0
	cache := &articleCategoryCacheStub{
		delFn: func(ctx context.Context, keys ...string) error {
			delCalled++
			if len(keys) != 1 || keys[0] != testArticleCategoriesListCacheKey {
				t.Fatalf("unexpected keys invalidated: %v", keys)
			}
			return nil
		},
	}
	repo := &articleCategoryRepoStub{
		getAllFn: func(ctx context.Context) ([]*entity.ArticleCategory, error) { return nil, nil },
		createFn: func(ctx context.Context, articleCategory *entity.ArticleCategory) (*entity.ArticleCategory, error) {
			return &entity.ArticleCategory{
				ArticleCategoryId: "1",
				NamaKategori:      articleCategory.NamaKategori,
			}, nil
		},
		updateFn: func(ctx context.Context, articleCategory *entity.ArticleCategory) (*entity.ArticleCategory, error) {
			return articleCategory, nil
		},
		deleteFn: func(ctx context.Context, articleCategoryId string) error {
			return nil
		},
	}

	uc := NewArticleCategoryUseCase(repo, cache, nil, validator.New())
	_, err := uc.CreateArticleCategory(context.Background(), &model.ArticleCategoryRequest{NamaKategori: "Politik"})
	if err != nil {
		t.Fatalf("create should not fail: %v", err)
	}
	_, err = uc.UpdateArticleCategory(context.Background(), &model.ArticleCategoryRequest{NamaKategori: "Olahraga"}, "1")
	if err != nil {
		t.Fatalf("update should not fail: %v", err)
	}
	err = uc.DeleteArticleCategory(context.Background(), "1")
	if err != nil {
		t.Fatalf("delete should not fail: %v", err)
	}
	if delCalled != 3 {
		t.Fatalf("expected 3 cache invalidations, got %d", delCalled)
	}
}

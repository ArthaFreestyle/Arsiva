package repository

import (
	"context"
	"ArthaFreestyle/Arsiva/internal/entity"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type ArticleCategoryRepository interface {
	GetAllArticleCategories(ctx context.Context) ([]*entity.ArticleCategory, error)
	GetArticleCategoryById(ctx context.Context,ArticleCategoryId string) (*entity.ArticleCategory, error)
	CreateArticleCategory(ctx context.Context,articleCategory *entity.ArticleCategory) (*entity.ArticleCategory, error)
	UpdateArticleCategory(ctx context.Context,articleCategory *entity.ArticleCategory) (*entity.ArticleCategory, error)
	DeleteArticleCategory(ctx context.Context,articleCategory *entity.ArticleCategory) (error)	
}

type ArticleCategoryRepositoryImpl struct {
	DB *pgxpool.Pool
	Log *logrus.Logger
}
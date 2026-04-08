package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type ArticleCategoryRepository interface {
	GetAllArticleCategories(ctx context.Context) ([]*entity.ArticleCategory, error)
	GetArticleCategoryById(ctx context.Context, ArticleCategoryId string) (*entity.ArticleCategory, error)
	CreateArticleCategory(ctx context.Context, articleCategory *entity.ArticleCategory) (*entity.ArticleCategory, error)
	UpdateArticleCategory(ctx context.Context, articleCategory *entity.ArticleCategory) (*entity.ArticleCategory, error)
	DeleteArticleCategory(ctx context.Context, articleCategoryId string) error
}

type ArticleCategoryRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewArticleCategoryRepository(db *pgxpool.Pool, log *logrus.Logger) ArticleCategoryRepository {
	return &ArticleCategoryRepositoryImpl{
		DB:  db,
		Log: log,
	}
}

func (r *ArticleCategoryRepositoryImpl) GetAllArticleCategories(ctx context.Context) ([]*entity.ArticleCategory, error) {
	SQL := `SELECT kategori_artikel_id,nama_kategori FROM kategori_artikel`
	rows, err := r.DB.Query(ctx, SQL)
	if err != nil {
		return nil, err
	}
	articleCategories, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[entity.ArticleCategory])
	if err != nil {
		return nil, err
	}
	r.Log.Info("query : ", SQL)
	return articleCategories, nil
}

func (r *ArticleCategoryRepositoryImpl) GetArticleCategoryById(ctx context.Context, ArticleCategoryId string) (*entity.ArticleCategory, error) {
	SQL := `SELECT kategori_artikel_id,nama_kategori FROM kategori_artikel WHERE kategori_artikel_id = $1`
	rows, err := r.DB.Query(ctx, SQL, ArticleCategoryId)
	if err != nil {
		return nil, err
	}
	articleCategory, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[entity.ArticleCategory])
	if err != nil {
		return nil, err
	}
	return articleCategory, nil
}

func (r *ArticleCategoryRepositoryImpl) CreateArticleCategory(ctx context.Context, articleCategory *entity.ArticleCategory) (*entity.ArticleCategory, error) {
	SQL := `INSERT INTO kategori_artikel (nama_kategori) VALUES ($1) RETURNING kategori_artikel_id,nama_kategori`
	rows, err := r.DB.Query(ctx, SQL, articleCategory.NamaKategori)
	if err != nil {
		return nil, err
	}
	articleCategory, err = pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[entity.ArticleCategory])
	if err != nil {
		return nil, err
	}
	return articleCategory, nil
}

func (r *ArticleCategoryRepositoryImpl) UpdateArticleCategory(ctx context.Context, articleCategory *entity.ArticleCategory) (*entity.ArticleCategory, error) {
	SQL := `UPDATE kategori_artikel SET nama_kategori = $1 WHERE kategori_artikel_id = $2 RETURNING kategori_artikel_id,nama_kategori`
	rows, err := r.DB.Query(ctx, SQL, articleCategory.NamaKategori, articleCategory.ArticleCategoryId)
	if err != nil {
		return nil, err
	}
	articleCategory, err = pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[entity.ArticleCategory])
	if err != nil {
		return nil, err
	}
	return articleCategory, nil
}

func (r *ArticleCategoryRepositoryImpl) DeleteArticleCategory(ctx context.Context, articleCategoryId string) error {
	SQL := `DELETE FROM kategori_artikel WHERE kategori_artikel_id = $1`
	_, err := r.DB.Exec(ctx, SQL, articleCategoryId)
	if err != nil {
		return err
	}
	return nil
}

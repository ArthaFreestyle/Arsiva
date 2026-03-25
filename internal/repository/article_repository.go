package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type ArticleRepository interface {
	GetAllArticle(ctx context.Context, page int, size int, search string) ([]*entity.Article, int, error)
	GetArticleBySlug(ctx context.Context, slug string) (*entity.Article, error)
	GetArticleById(ctx context.Context, articleId string) (*entity.Article, error)
	CreateArticle(ctx context.Context, article *entity.Article) (*entity.Article, error)
	UpdateArticle(ctx context.Context, article *entity.Article) (*entity.Article, error)
	DeleteArticle(ctx context.Context, articleId string) error
}

type articleRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewArticleRepository(db *pgxpool.Pool, log *logrus.Logger) ArticleRepository {
	return &articleRepositoryImpl{
		DB:  db,
		Log: log,
	}
}

func (r *articleRepositoryImpl) GetAllArticle(ctx context.Context, page int, size int, search string) ([]*entity.Article, int, error) {
	offset := (page - 1) * size
	searchPattern := "%" + search + "%"

	var total int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM artikel WHERE judul ILIKE $1`,
		searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	SQL := `SELECT 
    a.artikel_id,
    a.slug,
    a.judul,
    JSON_BUILD_OBJECT(
        'kategori_id', k.kategori_artikel_id, 
        'nama_kategori', k.nama_kategori
    ) AS kategori,
    a.status,
    a.excerpt,
    JSON_BUILD_OBJECT(
        'user_id', u.user_id, 
        'username', u.username
    ) AS "user",
    a.created_at,
    a.thumbnail 
FROM artikel a
LEFT JOIN kategori_artikel k ON a.kategori_id = k.kategori_artikel_id
LEFT JOIN users u ON a.created_by = u.user_id
WHERE a.judul ILIKE $1
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3`

	rows, err := r.DB.Query(ctx, SQL, searchPattern, size, offset)
	if err != nil {
		return nil, 0, err
	}
	articles, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[entity.Article])
	if err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}

func (r *articleRepositoryImpl) GetArticleBySlug(ctx context.Context, slug string) (*entity.Article, error) {
	SQL := `SELECT 
    a.artikel_id,
    a.slug,
    a.judul,
    a.konten,
    JSON_BUILD_OBJECT(
        'kategori_id', k.kategori_artikel_id, 
        'nama_kategori', k.nama_kategori
    ) AS kategori,
    a.status,
    a.excerpt,
    JSON_BUILD_OBJECT(
        'user_id', u.user_id, 
        'username', u.username
    ) AS "user",
    a.created_at,
    a.thumbnail 
FROM artikel a
LEFT JOIN kategori_artikel k ON a.kategori_id = k.kategori_artikel_id
LEFT JOIN users u ON a.created_by = u.user_id
WHERE a.slug = $1`
	rows, err := r.DB.Query(ctx, SQL, slug)
	if err != nil {
		return nil, err
	}
	article, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[entity.Article])
	if err != nil {
		return nil, err
	}
	return article, nil
}

func (r *articleRepositoryImpl) GetArticleById(ctx context.Context, articleId string) (*entity.Article, error) {
	SQL := `SELECT 
    a.artikel_id,
    a.slug,
    a.judul,
    a.konten,
    JSON_BUILD_OBJECT(
        'kategori_id', k.kategori_artikel_id, 
        'nama_kategori', k.nama_kategori
    ) AS kategori,
    a.status,
    a.excerpt,
    JSON_BUILD_OBJECT(
        'user_id', u.user_id, 
        'username', u.username
    ) AS "user",
    a.created_at,
    a.thumbnail 
FROM artikel a
LEFT JOIN kategori_artikel k ON a.kategori_id = k.kategori_artikel_id
LEFT JOIN users u ON a.created_by = u.user_id
WHERE a.artikel_id = $1`
	rows, err := r.DB.Query(ctx, SQL, articleId)
	if err != nil {
		return nil, err
	}
	article, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[entity.Article])
	if err != nil {
		return nil, err
	}
	return article, nil
}

func (r *articleRepositoryImpl) CreateArticle(ctx context.Context, article *entity.Article) (*entity.Article, error) {
	SQL := `INSERT INTO artikel (slug,judul,konten,kategori_id,status,excerpt,created_by,thumbnail) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING artikel_id,slug,judul,konten,kategori_id,status,excerpt,created_by,created_at,thumbnail`
	rows, err := r.DB.Query(ctx, SQL, article.Slug, article.Judul, article.Konten, article.KategoriId.ArticleCategoryId, article.Status, article.Excerpt, article.CreatedBy.UserId, article.Thumbnail)
	if err != nil {
		return nil, err
	}
	article, err = pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[entity.Article])
	if err != nil {
		return nil, err
	}
	return article, nil
}

func (r *articleRepositoryImpl) UpdateArticle(ctx context.Context, article *entity.Article) (*entity.Article, error) {
	SQL := `UPDATE artikel SET slug = $1,judul = $2,konten = $3,kategori_id = $4,status = $5,excerpt = $6,thumbnail = $7 WHERE artikel_id = $8 RETURNING artikel_id,slug,judul,konten,kategori_id,status,excerpt,created_by,created_at,thumbnail`
	rows, err := r.DB.Query(ctx, SQL, article.Slug, article.Judul, article.Konten, article.KategoriId.ArticleCategoryId, article.Status, article.Excerpt, article.Thumbnail, article.ArticleId)
	if err != nil {
		return nil, err
	}
	article, err = pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[entity.Article])
	if err != nil {
		return nil, err
	}
	return article, nil
}

func (r *articleRepositoryImpl) DeleteArticle(ctx context.Context, ArticleId string) error {
	SQL := `UPDATE artikel SET is_active = FALSE WHERE artikel_id = $1`
	_, err := r.DB.Exec(ctx, SQL, ArticleId)
	if err != nil {
		return err
	}
	return nil
}

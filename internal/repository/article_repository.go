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
        'ArticleCategoryId', k.kategori_artikel_id::text, 
        'NamaKategori', k.nama_kategori
    ) AS kategori,
    a.status,
    a.excerpt,
    JSON_BUILD_OBJECT(
        'UserId', u.user_id::text, 
        'Username', u.username
    ) AS "user",
    a.created_at,
    COALESCE(ass.url, '') AS thumbnail,
    a.thumbnail_asset_id
FROM artikel a
LEFT JOIN kategori_artikel k ON a.kategori_id = k.kategori_artikel_id
LEFT JOIN users u ON a.created_by = u.user_id
LEFT JOIN assets ass ON a.thumbnail_asset_id = ass.asset_id
WHERE a.judul ILIKE $1
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3`

	rows, err := r.DB.Query(ctx, SQL, searchPattern, size, offset)
	if err != nil {
		return nil, 0, err
	}
	articles, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.Article])
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
        'ArticleCategoryId', k.kategori_artikel_id::text, 
        'NamaKategori', k.nama_kategori
    ) AS kategori,
    a.status,
    a.excerpt,
    JSON_BUILD_OBJECT(
        'UserId', u.user_id::text, 
        'Username', u.username
    ) AS "user",
    a.created_at,
    COALESCE(ass.url, '') AS thumbnail,
    a.thumbnail_asset_id
FROM artikel a
LEFT JOIN kategori_artikel k ON a.kategori_id = k.kategori_artikel_id
LEFT JOIN users u ON a.created_by = u.user_id
LEFT JOIN assets ass ON a.thumbnail_asset_id = ass.asset_id
WHERE a.slug = $1`
	rows, err := r.DB.Query(ctx, SQL, slug)
	if err != nil {
		return nil, err
	}
	article, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Article])
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
        'ArticleCategoryId', k.kategori_artikel_id::text, 
        'NamaKategori', k.nama_kategori
    ) AS kategori,
    a.status,
    a.excerpt,
    JSON_BUILD_OBJECT(
        'UserId', u.user_id::text, 
        'Username', u.username
    ) AS "user",
    a.created_at,
    COALESCE(ass.url, '') AS thumbnail,
    a.thumbnail_asset_id
FROM artikel a
LEFT JOIN kategori_artikel k ON a.kategori_id = k.kategori_artikel_id
LEFT JOIN users u ON a.created_by = u.user_id
LEFT JOIN assets ass ON a.thumbnail_asset_id = ass.asset_id
WHERE a.artikel_id = $1`
	rows, err := r.DB.Query(ctx, SQL, articleId)
	if err != nil {
		return nil, err
	}
	article, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Article])
	if err != nil {
		return nil, err
	}
	return article, nil
}

func (r *articleRepositoryImpl) CreateArticle(ctx context.Context, article *entity.Article) (*entity.Article, error) {
	SQL := `INSERT INTO artikel (slug,judul,kategori_id,created_by,thumbnail_asset_id) VALUES ($1,$2,$3,$4,$5) RETURNING artikel_id`
    var id string
	err := r.DB.QueryRow(ctx, SQL, article.Slug, article.Judul, article.KategoriId, article.CreatedBy.UserId, article.ThumbnailAssetId).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetArticleById(ctx, id)
}

func (r *articleRepositoryImpl) UpdateArticle(ctx context.Context, article *entity.Article) (*entity.Article, error) {
	SQL := `UPDATE artikel SET slug = $1,judul = $2,konten = $3,kategori_id = $4,status = $5,excerpt = $6,thumbnail_asset_id = $7 WHERE artikel_id = $8 RETURNING artikel_id`
    var id string
	err := r.DB.QueryRow(ctx, SQL, 
		article.Slug, 
		article.Judul, 
		article.Konten, 
		article.KategoriId, 
		article.Status, 
		article.Excerpt, 
		article.ThumbnailAssetId, 
		article.ArticleId,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetArticleById(ctx, id)
}

func (r *articleRepositoryImpl) DeleteArticle(ctx context.Context, ArticleId string) error {
	SQL := `UPDATE artikel SET is_active = FALSE WHERE artikel_id = $1`
	_, err := r.DB.Exec(ctx, SQL, ArticleId)
	if err != nil {
		return err
	}
	return nil
}

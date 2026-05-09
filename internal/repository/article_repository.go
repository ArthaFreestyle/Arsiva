package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"
	"fmt"

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

	GetAllArticleManage(ctx context.Context, page int, size int, search string, userId string, role string) ([]*entity.Article, int, error)
	GetArticleByIdManage(ctx context.Context, articleId string, userId string, role string) (*entity.Article, error)
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
		`SELECT COUNT(*) FROM artikel a WHERE a.status = 'published' AND a.is_active = TRUE AND a.judul ILIKE $1`,
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
WHERE a.status = 'published' AND a.is_active = TRUE AND a.judul ILIKE $1
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
WHERE a.slug = $1 AND a.status = 'published' AND a.is_active = TRUE`
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
WHERE a.artikel_id = $1 AND a.status = 'published' AND a.is_active = TRUE`
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

func (r *articleRepositoryImpl) getArticleByIdUnfiltered(ctx context.Context, articleId string) (*entity.Article, error) {
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
	return r.getArticleByIdUnfiltered(ctx, id)
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
	return r.getArticleByIdUnfiltered(ctx, id)
}

func (r *articleRepositoryImpl) DeleteArticle(ctx context.Context, ArticleId string) error {
	SQL := `UPDATE artikel SET is_active = FALSE WHERE artikel_id = $1`
	_, err := r.DB.Exec(ctx, SQL, ArticleId)
	if err != nil {
		return err
	}
	return nil
}

func (r *articleRepositoryImpl) GetAllArticleManage(ctx context.Context, page int, size int, search string, userId string, role string) ([]*entity.Article, int, error) {
	offset := (page - 1) * size
	searchPattern := "%" + search + "%"

	var whereClause string
	var countArgs, queryArgs []interface{}

	if role == "super_admin" {
		whereClause = "WHERE a.is_active = TRUE AND a.judul ILIKE $1"
		countArgs = []interface{}{searchPattern}
		queryArgs = []interface{}{searchPattern, size, offset}
	} else {
		whereClause = "WHERE a.is_active = TRUE AND a.created_by = $1 AND a.judul ILIKE $2"
		countArgs = []interface{}{userId, searchPattern}
		queryArgs = []interface{}{userId, searchPattern, size, offset}
	}

	var total int
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM artikel a %s`, whereClause)
	err := r.DB.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limitOffset := fmt.Sprintf("LIMIT $%d OFFSET $%d", len(countArgs)+1, len(countArgs)+2)

	SQL := fmt.Sprintf(`SELECT
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
%s
ORDER BY a.created_at DESC
%s`, whereClause, limitOffset)

	rows, err := r.DB.Query(ctx, SQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	articles, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.Article])
	if err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}

func (r *articleRepositoryImpl) GetArticleByIdManage(ctx context.Context, articleId string, userId string, role string) (*entity.Article, error) {
	var whereClause string
	var args []interface{}

	if role == "super_admin" {
		whereClause = "WHERE a.artikel_id = $1 AND a.is_active = TRUE"
		args = []interface{}{articleId}
	} else {
		whereClause = "WHERE a.artikel_id = $1 AND a.created_by = $2 AND a.is_active = TRUE"
		args = []interface{}{articleId, userId}
	}

	SQL := fmt.Sprintf(`SELECT
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
%s`, whereClause)

	rows, err := r.DB.Query(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	article, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Article])
	if err != nil {
		return nil, err
	}
	return article, nil
}

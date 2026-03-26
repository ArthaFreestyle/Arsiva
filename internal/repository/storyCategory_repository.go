package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type StoryCategoryRepository interface {
	GetAllStoryCategories(ctx context.Context, page int, size int, search string) ([]*entity.StoryCategory, int, error)
	GetStoryCategoryById(ctx context.Context, storyCategoryId string) (*entity.StoryCategory, error)
	CreateStoryCategory(ctx context.Context, storyCategory *entity.StoryCategory) (*entity.StoryCategory, error)
	UpdateStoryCategory(ctx context.Context, storyCategory *entity.StoryCategory) (*entity.StoryCategory, error)
	DeleteStoryCategory(ctx context.Context, storyCategoryId string) error
}

type StoryCategoryRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewStoryCategoryRepository(db *pgxpool.Pool, log *logrus.Logger) StoryCategoryRepository {
	return &StoryCategoryRepositoryImpl{
		DB:  db,
		Log: log,
	}
}

func (r *StoryCategoryRepositoryImpl) GetAllStoryCategories(ctx context.Context, page int, size int, search string) ([]*entity.StoryCategory, int, error) {
	offset := (page - 1) * size
	searchPattern := "%" + search + "%"

	var total int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM kategori_cerita WHERE nama_kategori ILIKE $1`,
		searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	SQL := `SELECT kategori_id, nama_kategori FROM kategori_cerita WHERE nama_kategori ILIKE $1 ORDER BY kategori_id ASC LIMIT $2 OFFSET $3`
	rows, err := r.DB.Query(ctx, SQL, searchPattern, size, offset)
	if err != nil {
		return nil, 0, err
	}
	storyCategories, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[entity.StoryCategory])
	if err != nil {
		return nil, 0, err
	}
	return storyCategories, total, nil
}

func (r *StoryCategoryRepositoryImpl) GetStoryCategoryById(ctx context.Context, storyCategoryId string) (*entity.StoryCategory, error) {
	SQL := `SELECT kategori_id, nama_kategori FROM kategori_cerita WHERE kategori_id = $1`
	rows, err := r.DB.Query(ctx, SQL, storyCategoryId)
	if err != nil {
		return nil, err
	}
	storyCategory, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[entity.StoryCategory])
	if err != nil {
		return nil, err
	}
	return storyCategory, nil
}

func (r *StoryCategoryRepositoryImpl) CreateStoryCategory(ctx context.Context, storyCategory *entity.StoryCategory) (*entity.StoryCategory, error) {
	SQL := `INSERT INTO kategori_cerita (nama_kategori) VALUES ($1) RETURNING kategori_id, nama_kategori`
	rows, err := r.DB.Query(ctx, SQL, storyCategory.NamaKategori)
	if err != nil {
		return nil, err
	}
	storyCategory, err = pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[entity.StoryCategory])
	if err != nil {
		return nil, err
	}
	return storyCategory, nil
}

func (r *StoryCategoryRepositoryImpl) UpdateStoryCategory(ctx context.Context, storyCategory *entity.StoryCategory) (*entity.StoryCategory, error) {
	SQL := `UPDATE kategori_cerita SET nama_kategori = $1 WHERE kategori_id = $2 RETURNING kategori_id, nama_kategori`
	rows, err := r.DB.Query(ctx, SQL, storyCategory.NamaKategori, storyCategory.StoryCategoryId)
	if err != nil {
		return nil, err
	}
	storyCategory, err = pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[entity.StoryCategory])
	if err != nil {
		return nil, err
	}
	return storyCategory, nil
}

func (r *StoryCategoryRepositoryImpl) DeleteStoryCategory(ctx context.Context, storyCategoryId string) error {
	SQL := `DELETE FROM kategori_cerita WHERE kategori_id = $1`
	_, err := r.DB.Exec(ctx, SQL, storyCategoryId)
	if err != nil {
		return err
	}
	return nil
}

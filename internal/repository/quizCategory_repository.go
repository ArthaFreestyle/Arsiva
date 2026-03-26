package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type QuizCategoryRepository interface {
	GetAllQuizCategories(ctx context.Context, page int, size int, search string) ([]*entity.QuizCategory, int, error)
	GetQuizCategoryById(ctx context.Context, quizCategoryId string) (*entity.QuizCategory, error)
	CreateQuizCategory(ctx context.Context, quizCategory *entity.QuizCategory) (*entity.QuizCategory, error)
	UpdateQuizCategory(ctx context.Context, quizCategory *entity.QuizCategory) (*entity.QuizCategory, error)
	DeleteQuizCategory(ctx context.Context, quizCategoryId string) error
}

type QuizCategoryRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewQuizCategoryRepository(db *pgxpool.Pool, log *logrus.Logger) QuizCategoryRepository {
	return &QuizCategoryRepositoryImpl{
		DB:  db,
		Log: log,
	}
}

func (r *QuizCategoryRepositoryImpl) GetAllQuizCategories(ctx context.Context, page int, size int, search string) ([]*entity.QuizCategory, int, error) {
	offset := (page - 1) * size
	searchPattern := "%" + search + "%"

	var total int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM kategori_kuis WHERE nama_kategori ILIKE $1`,
		searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	SQL := `SELECT kategori_id, nama_kategori, created_at, created_by::text, deskripsi 
            FROM kategori_kuis 
            WHERE nama_kategori ILIKE $1 
            ORDER BY kategori_id ASC 
            LIMIT $2 OFFSET $3`
            
	rows, err := r.DB.Query(ctx, SQL, searchPattern, size, offset)
	if err != nil {
		return nil, 0, err
	}
	quizCategories, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.QuizCategory])
	if err != nil {
		return nil, 0, err
	}
	return quizCategories, total, nil
}

func (r *QuizCategoryRepositoryImpl) GetQuizCategoryById(ctx context.Context, quizCategoryId string) (*entity.QuizCategory, error) {
	SQL := `SELECT kategori_id, nama_kategori, created_at, created_by::text, deskripsi 
            FROM kategori_kuis WHERE kategori_id = $1`
	rows, err := r.DB.Query(ctx, SQL, quizCategoryId)
	if err != nil {
		return nil, err
	}
	quizCategory, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.QuizCategory])
	if err != nil {
		return nil, err
	}
	return quizCategory, nil
}

func (r *QuizCategoryRepositoryImpl) CreateQuizCategory(ctx context.Context, quizCategory *entity.QuizCategory) (*entity.QuizCategory, error) {
	SQL := `INSERT INTO kategori_kuis (nama_kategori, created_at, created_by, deskripsi) 
            VALUES ($1, $2, $3, $4) 
            RETURNING kategori_id, nama_kategori, created_at, created_by::text, deskripsi`
	rows, err := r.DB.Query(ctx, SQL, quizCategory.NamaKategori, quizCategory.CreatedAt, quizCategory.CreatedBy, quizCategory.Deskripsi)
	if err != nil {
		return nil, err
	}
	quizCategory, err = pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.QuizCategory])
	if err != nil {
		return nil, err
	}
	return quizCategory, nil
}

func (r *QuizCategoryRepositoryImpl) UpdateQuizCategory(ctx context.Context, quizCategory *entity.QuizCategory) (*entity.QuizCategory, error) {
	SQL := `UPDATE kategori_kuis 
            SET nama_kategori = $1, deskripsi = $2 
            WHERE kategori_id = $3 
            RETURNING kategori_id, nama_kategori, created_at, created_by::text, deskripsi`
	rows, err := r.DB.Query(ctx, SQL, quizCategory.NamaKategori, quizCategory.Deskripsi, quizCategory.QuizCategoryId)
	if err != nil {
		return nil, err
	}
	quizCategory, err = pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.QuizCategory])
	if err != nil {
		return nil, err
	}
	return quizCategory, nil
}

func (r *QuizCategoryRepositoryImpl) DeleteQuizCategory(ctx context.Context, quizCategoryId string) error {
	SQL := `DELETE FROM kategori_kuis WHERE kategori_id = $1`
	_, err := r.DB.Exec(ctx, SQL, quizCategoryId)
	if err != nil {
		return err
	}
	return nil
}

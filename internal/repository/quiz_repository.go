package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type QuizRepository interface {
	GetAll(ctx context.Context) ([]*entity.Quiz, error)
	GetByID(ctx context.Context, puzzleId string) (*entity.Quiz, error)
	Create(ctx context.Context, puzzle *entity.Quiz) (*entity.Quiz, error)
	Update(ctx context.Context, puzzle *entity.Quiz) (*entity.Quiz, error)
	Delete(ctx context.Context, puzzleId string) error
}

type quizRepository struct {
	db *pgxpool.Pool
	Log *logrus.Logger
}

func NewQuizRepository(db *pgxpool.Pool, log *logrus.Logger) QuizRepository {
	return &quizRepository{db: db, Log: log}
}

func (r *quizRepository) GetAll(ctx context.Context) ([]*entity.Quiz, error) {
	rows, err := r.db.Query(ctx, "SELECT * FROM puzzles")
	if err != nil {
		return nil, err
	}
	quizzes,err := pgx.CollectRows(rows,pgx.RowToAddrOfStructByNameLax[entity.Quiz])
	if err != nil {
		return nil, err
	}
	return quizzes, nil
}

func (r *quizRepository) GetByID(ctx context.Context, id string) (*entity.Quiz, error) {
	rows,err := r.db.Query(ctx, "SELECT * FROM quizzes WHERE puzzle_id = $1", id)
	if err != nil {
		return nil, err
	}
	
	quiz,err := pgx.CollectOneRow(rows,pgx.RowToAddrOfStructByNameLax[entity.Quiz])
	return quiz, err
}

func (r *quizRepository) Create(ctx context.Context, quiz *entity.Quiz) (*entity.Quiz, error) {
	rows,err := r.db.Query(ctx, "INSERT INTO quizzes (judul, gambar, thumbnail, kategori, xp_reward, created_by, is_published) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING puzzle_id", quiz.Judul, quiz.Gambar, quiz.Thumbnail, quiz.KategoriId, quiz.XpReward, quiz.CreatedBy, quiz.IsPublished)
	if err != nil {
		return nil, err
	}
	
	quiz,err = pgx.CollectOneRow(rows,pgx.RowToAddrOfStructByNameLax[entity.Quiz])
	if err != nil {
		return nil, err
	}
	return quiz, nil
}

func (r *quizRepository) Update(ctx context.Context, quiz *entity.Quiz) (*entity.Quiz, error) {
	rows,err := r.db.Query(ctx, "UPDATE quizzes SET judul = $1, gambar = $2, thumbnail = $3, kategori = $4, xp_reward = $5, created_by = $6, is_published = $7 WHERE puzzle_id = $8 RETURNING puzzle_id", quiz.Judul, quiz.Gambar, quiz.Thumbnail, quiz.KategoriId, quiz.XpReward, quiz.CreatedBy, quiz.IsPublished, quiz.QuizId)
	if err != nil {
		return nil, err
	}
	
	quiz,err = pgx.CollectOneRow(rows,pgx.RowToAddrOfStructByNameLax[entity.Quiz])
	if err != nil {
		return nil, err
	}
	return quiz, nil
}

func (r *quizRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM quizzes WHERE puzzle_id = $1", id)
	return err
}
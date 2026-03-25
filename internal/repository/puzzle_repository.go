package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type PuzzleRepository interface {
	FindAll(ctx context.Context, page int, size int, search string) ([]*entity.Puzzle, int, error)
	FindById(ctx context.Context, puzzleId string) (*entity.Puzzle, error)
	Create(ctx context.Context, puzzle *entity.Puzzle) (*entity.Puzzle, error)
	Update(ctx context.Context, puzzle *entity.Puzzle) (*entity.Puzzle, error)
	Delete(ctx context.Context, puzzleId string) (error)
}

type puzzleRepositoryImpl struct {
	DB *pgxpool.Pool
	Log *logrus.Logger
}

func NewPuzzleRepository(db *pgxpool.Pool,log *logrus.Logger) PuzzleRepository {
	return &puzzleRepositoryImpl{
		DB: db,
		Log: log,
	}
}

func (r *puzzleRepositoryImpl) FindAll(ctx context.Context, page int, size int, search string) ([]*entity.Puzzle, int, error) {
	offset := (page - 1) * size
	searchPattern := "%" + search + "%"

	var total int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM puzzles WHERE is_published = true AND judul ILIKE $1`,
		searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	SQL := `SELECT p.puzzle_id,p.judul,p.thumbnail,p.kategori,p.xp_reward,
	JSON_BUILD_OBJECT(
		'user_id',u.user_id,
		'username',u.username
	) AS user,
	p.created_at,p.is_published
	FROM puzzles p 
	JOIN users u ON p.created_by = u.user_id 
	WHERE p.is_published = true AND p.judul ILIKE $1
	ORDER BY p.created_at DESC
	LIMIT $2 OFFSET $3`

	rows,err := r.DB.Query(ctx,SQL,searchPattern,size,offset)
	if err != nil {
		return nil,0,err
	}
	
	puzzles,err := pgx.CollectRows(rows,pgx.RowToAddrOfStructByNameLax[entity.Puzzle])
	if err != nil {
		return nil,0,err
	}
	return puzzles,total,nil
}

func (r *puzzleRepositoryImpl) FindById(ctx context.Context, puzzleId string) (*entity.Puzzle, error) {
	SQL := `SELECT p.puzzle_id,p.judul,p.gambar,p.kategori,p.xp_reward,
	p.created_at,p.is_published
	FROM puzzles p 
	WHERE p.is_published = true 
	AND p.puzzle_id = $1`

	rows,err := r.DB.Query(ctx,SQL,puzzleId)
	if err != nil {
		return nil,err
	}

	puzzle,err := pgx.CollectOneRow(rows,pgx.RowToAddrOfStructByNameLax[entity.Puzzle])
	if err != nil {
		return nil,err
	}
	return puzzle,nil
}

func (r *puzzleRepositoryImpl) Create(ctx context.Context, puzzle *entity.Puzzle) (*entity.Puzzle, error) {
	SQL := `INSERT INTO puzzles (judul,gambar,thumbnail,kategori,xp_reward,created_by,is_published) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING puzzle_id`
	
	rows,err := r.DB.Query(ctx,SQL,puzzle.Judul,puzzle.Gambar,puzzle.Thumbnail,puzzle.Kategori,puzzle.XpReward,puzzle.CreatedBy.UserId,puzzle.IsPublished)
	if err != nil {
		return nil,err
	}

	puzzle,err = pgx.CollectOneRow(rows,pgx.RowToAddrOfStructByNameLax[entity.Puzzle])
	if err != nil {
		return nil,err
	}
	return puzzle,nil
}

func (r *puzzleRepositoryImpl) Update(ctx context.Context, puzzle *entity.Puzzle) (*entity.Puzzle, error) {
	SQL := `UPDATE puzzles SET judul = $1,gambar = $2,thumbnail = $3,kategori = $4,xp_reward = $5,is_published = $6 WHERE puzzle_id = $7 RETURNING puzzle_id`
	
	rows,err := r.DB.Query(ctx,SQL,puzzle.Judul,puzzle.Gambar,puzzle.Thumbnail,puzzle.Kategori,puzzle.XpReward,puzzle.IsPublished,puzzle.PuzzleId)
	if err != nil {
		return nil,err
	}

	puzzle,err = pgx.CollectOneRow(rows,pgx.RowToAddrOfStructByNameLax[entity.Puzzle])
	if err != nil {
		return nil,err
	}
	return puzzle,nil
}

func (r *puzzleRepositoryImpl) Delete(ctx context.Context, puzzleId string) (error) {
	SQL := `DELETE FROM puzzles WHERE puzzle_id = $1`
	
	_,err := r.DB.Exec(ctx,SQL,puzzleId)
	if err != nil {
		return err
	}
	return nil
}
	
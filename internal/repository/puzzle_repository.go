package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"
	"fmt"

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

	FindAllManage(ctx context.Context, page int, size int, search string, userId string, role string) ([]*entity.Puzzle, int, error)
	FindByIdManage(ctx context.Context, puzzleId string, userId string, role string) (*entity.Puzzle, error)
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

	SQL := `SELECT p.puzzle_id,p.judul,COALESCE(ass_g.url, '') AS gambar, p.gambar_asset_id, 
	COALESCE(ass_t.url, '') AS thumbnail, p.thumbnail_asset_id, p.kategori,p.xp_reward,
	JSON_BUILD_OBJECT(
		'user_id',u.user_id::text,
		'username',u.username
	) AS user,
	p.created_at,p.is_published
	FROM puzzles p 
	JOIN users u ON p.created_by = u.user_id 
	LEFT JOIN assets ass_t ON p.thumbnail_asset_id = ass_t.asset_id
	LEFT JOIN assets ass_g ON p.gambar_asset_id = ass_g.asset_id
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
	SQL := `SELECT p.puzzle_id,p.judul,COALESCE(ass_g.url, '') AS gambar, p.gambar_asset_id, 
	COALESCE(ass_t.url, '') AS thumbnail, p.thumbnail_asset_id, p.kategori,p.xp_reward,
	JSON_BUILD_OBJECT(
		'user_id',u.user_id::text,
		'username',u.username
	) AS user,
	p.created_at,p.is_published
	FROM puzzles p 
	JOIN users u ON p.created_by = u.user_id 
	LEFT JOIN assets ass_g ON p.gambar_asset_id = ass_g.asset_id
	LEFT JOIN assets ass_t ON p.thumbnail_asset_id = ass_t.asset_id
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
	SQL := `INSERT INTO puzzles (judul,gambar_asset_id,thumbnail_asset_id,kategori,xp_reward,created_by,is_published) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING puzzle_id`
	
    var id string
	err := r.DB.QueryRow(ctx,SQL,puzzle.Judul,puzzle.GambarAssetId,puzzle.ThumbnailAssetId,puzzle.Kategori,puzzle.XpReward,puzzle.CreatedBy.UserId,puzzle.IsPublished).Scan(&id)
	if err != nil {
		return nil,err
	}

	return r.findByIdUnfiltered(ctx, id)
}

func (r *puzzleRepositoryImpl) Update(ctx context.Context, puzzle *entity.Puzzle) (*entity.Puzzle, error) {
	SQL := `UPDATE puzzles SET judul = $1,gambar_asset_id = $2,thumbnail_asset_id = $3,kategori = $4,xp_reward = $5,is_published = $6 WHERE puzzle_id = $7 RETURNING puzzle_id`
	
	var id string
	err := r.DB.QueryRow(ctx,SQL,puzzle.Judul,puzzle.GambarAssetId,puzzle.ThumbnailAssetId,puzzle.Kategori,puzzle.XpReward,puzzle.IsPublished,puzzle.PuzzleId).Scan(&id)
	if err != nil {
		return nil,err
	}

	return r.findByIdUnfiltered(ctx, id)
}

func (r *puzzleRepositoryImpl) Delete(ctx context.Context, puzzleId string) (error) {
	SQL := `DELETE FROM puzzles WHERE puzzle_id = $1`
	
	_,err := r.DB.Exec(ctx,SQL,puzzleId)
	if err != nil {
		return err
	}
	return nil
}

func (r *puzzleRepositoryImpl) findByIdUnfiltered(ctx context.Context, puzzleId string) (*entity.Puzzle, error) {
	SQL := `SELECT p.puzzle_id,p.judul,COALESCE(ass_g.url, '') AS gambar, p.gambar_asset_id, 
	COALESCE(ass_t.url, '') AS thumbnail, p.thumbnail_asset_id, p.kategori,p.xp_reward,
	JSON_BUILD_OBJECT(
		'user_id',u.user_id::text,
		'username',u.username
	) AS user,
	p.created_at,p.is_published
	FROM puzzles p 
	JOIN users u ON p.created_by = u.user_id 
	LEFT JOIN assets ass_g ON p.gambar_asset_id = ass_g.asset_id
	LEFT JOIN assets ass_t ON p.thumbnail_asset_id = ass_t.asset_id
	WHERE p.puzzle_id = $1`

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

func (r *puzzleRepositoryImpl) FindAllManage(ctx context.Context, page int, size int, search string, userId string, role string) ([]*entity.Puzzle, int, error) {
	offset := (page - 1) * size
	searchPattern := "%" + search + "%"

	var whereClause string
	var countArgs, queryArgs []interface{}

	if role == "super_admin" {
		whereClause = "WHERE p.judul ILIKE $1"
		countArgs = []interface{}{searchPattern}
		queryArgs = []interface{}{searchPattern, size, offset}
	} else {
		whereClause = "WHERE p.created_by = $1 AND p.judul ILIKE $2"
		countArgs = []interface{}{userId, searchPattern}
		queryArgs = []interface{}{userId, searchPattern, size, offset}
	}

	var total int
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM puzzles p %s`, whereClause)
	err := r.DB.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limitOffset := fmt.Sprintf("LIMIT $%d OFFSET $%d", len(countArgs)+1, len(countArgs)+2)

	SQL := fmt.Sprintf(`SELECT p.puzzle_id, p.judul,
		COALESCE(ass_g.url, '') AS gambar, p.gambar_asset_id,
		COALESCE(ass_t.url, '') AS thumbnail, p.thumbnail_asset_id,
		p.kategori, p.xp_reward,
		JSON_BUILD_OBJECT(
			'user_id', u.user_id::text,
			'username', u.username
		) AS user,
		p.created_at, p.is_published
	FROM puzzles p
	JOIN users u ON p.created_by = u.user_id
	LEFT JOIN assets ass_t ON p.thumbnail_asset_id = ass_t.asset_id
	LEFT JOIN assets ass_g ON p.gambar_asset_id = ass_g.asset_id
	%s
	ORDER BY p.created_at DESC
	%s`, whereClause, limitOffset)

	rows, err := r.DB.Query(ctx, SQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}

	puzzles, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.Puzzle])
	if err != nil {
		return nil, 0, err
	}
	return puzzles, total, nil
}

func (r *puzzleRepositoryImpl) FindByIdManage(ctx context.Context, puzzleId string, userId string, role string) (*entity.Puzzle, error) {
	var whereClause string
	var args []interface{}

	if role == "super_admin" {
		whereClause = "WHERE p.puzzle_id = $1"
		args = []interface{}{puzzleId}
	} else {
		whereClause = "WHERE p.puzzle_id = $1 AND p.created_by = $2"
		args = []interface{}{puzzleId, userId}
	}

	SQL := fmt.Sprintf(`SELECT p.puzzle_id, p.judul,
		COALESCE(ass_g.url, '') AS gambar, p.gambar_asset_id,
		COALESCE(ass_t.url, '') AS thumbnail, p.thumbnail_asset_id,
		p.kategori, p.xp_reward,
		JSON_BUILD_OBJECT(
			'user_id', u.user_id::text,
			'username', u.username
		) AS user,
		p.created_at, p.is_published
	FROM puzzles p
	JOIN users u ON p.created_by = u.user_id
	LEFT JOIN assets ass_g ON p.gambar_asset_id = ass_g.asset_id
	LEFT JOIN assets ass_t ON p.thumbnail_asset_id = ass_t.asset_id
	%s`, whereClause)

	rows, err := r.DB.Query(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}

	puzzle, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Puzzle])
	if err != nil {
		return nil, err
	}
	return puzzle, nil
}
	
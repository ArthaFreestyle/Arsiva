package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type SekolahRepository interface {
	Create(ctx context.Context, sekolah *entity.Sekolah) (*entity.Sekolah, error)
	FindById(ctx context.Context, sekolahId string) (*entity.Sekolah, error)
	FindAll(ctx context.Context, search string, limit int, offset int) ([]*entity.Sekolah, int, error)
	Update(ctx context.Context, sekolah *entity.Sekolah) (*entity.Sekolah, error)
	Delete(ctx context.Context, sekolahId string) error
	FindGurusBySekolahId(ctx context.Context, sekolahId string) ([]*entity.Guru, error)
	CountGurusBySekolahId(ctx context.Context, sekolahId string) (int, error)
}

type sekolahRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewSekolahRepository(db *pgxpool.Pool, log *logrus.Logger) SekolahRepository {
	return &sekolahRepositoryImpl{
		DB:  db,
		Log: log,
	}
}

func (r *sekolahRepositoryImpl) Create(ctx context.Context, sekolah *entity.Sekolah) (*entity.Sekolah, error) {
	query := `
		INSERT INTO sekolah (nama_sekolah, alamat_sekolah)
		VALUES ($1, $2)
		RETURNING sekolah_id::text, nama_sekolah, alamat_sekolah
	`
	rows, err := r.DB.Query(ctx, query, sekolah.NamaSekolah, sekolah.AlamatSekolah)
	if err != nil {
		r.Log.Errorf("Error Create sekolah: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Sekolah])
	if err != nil {
		r.Log.Errorf("Error collecting row Create sekolah: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *sekolahRepositoryImpl) FindById(ctx context.Context, sekolahId string) (*entity.Sekolah, error) {
	id, err := strconv.Atoi(sekolahId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT sekolah_id::text, nama_sekolah, alamat_sekolah
		FROM sekolah
		WHERE sekolah_id = $1
	`
	rows, err := r.DB.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error query FindById sekolah: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Sekolah])
	if err != nil {
		r.Log.Errorf("Error collecting row FindById sekolah: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *sekolahRepositoryImpl) FindAll(ctx context.Context, search string, limit int, offset int) ([]*entity.Sekolah, int, error) {
	searchPattern := "%" + search + "%"

	countQuery := `SELECT COUNT(*) FROM sekolah WHERE nama_sekolah ILIKE $1`
	var total int
	err := r.DB.QueryRow(ctx, countQuery, searchPattern).Scan(&total)
	if err != nil {
		r.Log.Errorf("Error counting FindAll sekolah: %v", err)
		return nil, 0, err
	}

	query := `
		SELECT sekolah_id::text, nama_sekolah, alamat_sekolah
		FROM sekolah
		WHERE nama_sekolah ILIKE $1
		ORDER BY sekolah_id ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.DB.Query(ctx, query, searchPattern, limit, offset)
	if err != nil {
		r.Log.Errorf("Error query FindAll sekolah: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	sekolahs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.Sekolah])
	if err != nil {
		r.Log.Errorf("Error collecting rows FindAll sekolah: %v", err)
		return nil, 0, err
	}
	return sekolahs, total, nil
}

func (r *sekolahRepositoryImpl) Update(ctx context.Context, sekolah *entity.Sekolah) (*entity.Sekolah, error) {
	id, err := strconv.Atoi(sekolah.SekolahId)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE sekolah
		SET nama_sekolah = $1, alamat_sekolah = $2
		WHERE sekolah_id = $3
		RETURNING sekolah_id::text, nama_sekolah, alamat_sekolah
	`
	rows, err := r.DB.Query(ctx, query, sekolah.NamaSekolah, sekolah.AlamatSekolah, id)
	if err != nil {
		r.Log.Errorf("Error Update sekolah: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Sekolah])
	if err != nil {
		r.Log.Errorf("Error collecting row Update sekolah: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *sekolahRepositoryImpl) Delete(ctx context.Context, sekolahId string) error {
	id, err := strconv.Atoi(sekolahId)
	if err != nil {
		return err
	}

	query := `DELETE FROM sekolah WHERE sekolah_id = $1`
	_, err = r.DB.Exec(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error Delete sekolah: %v", err)
		return err
	}
	return nil
}

func (r *sekolahRepositoryImpl) FindGurusBySekolahId(ctx context.Context, sekolahId string) ([]*entity.Guru, error) {
	id, err := strconv.Atoi(sekolahId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT g.guru_id::text AS guru_id, g.nip, g.bidang_ajar, u.username
		FROM guru g
		JOIN users u ON g.user_id = u.user_id
		WHERE g.sekolah_id = $1
	`
	rows, err := r.DB.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error query FindGurusBySekolahId: %v", err)
		return nil, err
	}
	defer rows.Close()

	gurus, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.Guru])
	if err != nil {
		r.Log.Errorf("Error collecting rows FindGurusBySekolahId: %v", err)
		return nil, err
	}
	return gurus, nil
}

func (r *sekolahRepositoryImpl) CountGurusBySekolahId(ctx context.Context, sekolahId string) (int, error) {
	id, err := strconv.Atoi(sekolahId)
	if err != nil {
		return 0, err
	}

	query := `SELECT COUNT(*) FROM guru WHERE sekolah_id = $1`
	var count int
	err = r.DB.QueryRow(ctx, query, id).Scan(&count)
	if err != nil {
		r.Log.Errorf("Error CountGurusBySekolahId: %v", err)
		return 0, err
	}
	return count, nil
}

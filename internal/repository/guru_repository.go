package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type GuruRepository interface {
	Create(ctx context.Context, guru *entity.Guru) (*entity.Guru, error)
	FindById(ctx context.Context, guruId string) (*entity.Guru, error)
	FindByUserId(ctx context.Context, userId string) (*entity.Guru, error)
	FindAll(ctx context.Context, search string, limit int, offset int) ([]*entity.Guru, int, error)
	Update(ctx context.Context, guru *entity.Guru) (*entity.Guru, error)
	Delete(ctx context.Context, guruId string) error
	FindSekolahByGuruId(ctx context.Context, guruId string) (*entity.Sekolah, error)
	FindGroupsByGuruId(ctx context.Context, guruId string) ([]*entity.Group, error)
}

type guruRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewGuruRepository(db *pgxpool.Pool, log *logrus.Logger) GuruRepository {
	return &guruRepositoryImpl{
		DB:  db,
		Log: log,
	}
}

func (r *guruRepositoryImpl) Create(ctx context.Context, guru *entity.Guru) (*entity.Guru, error) {
	userId, err := strconv.Atoi(guru.UserId)
	if err != nil {
		return nil, err
	}

	query := `
		WITH inserted AS (
			INSERT INTO guru (user_id, sekolah_id, nip, bidang_ajar)
			VALUES ($1, $2, $3, $4)
			RETURNING guru_id, user_id, sekolah_id, nip, bidang_ajar
		)
		SELECT i.guru_id::text,
		       i.user_id::text,
		       COALESCE(i.sekolah_id::text, '') AS sekolah_id,
		       COALESCE(i.nip, '') AS nip,
		       COALESCE(i.bidang_ajar, '') AS bidang_ajar,
		       u.username,
		       u.email
		FROM inserted i
		JOIN users u ON i.user_id = u.user_id
	`

	var sekolahId *int
	if guru.SekolahId != "" {
		sid, err := strconv.Atoi(guru.SekolahId)
		if err != nil {
			return nil, err
		}
		sekolahId = &sid
	}

	rows, err := r.DB.Query(ctx, query, userId, sekolahId, guru.NIP, guru.BidangAjar)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "nip") {
				return nil, errNIPAlreadyUsed
			}
			return nil, errUserAlreadyGuru
		}
		r.Log.Errorf("Error Create guru: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Guru])
	if err != nil {
		r.Log.Errorf("Error collecting row Create guru: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *guruRepositoryImpl) FindById(ctx context.Context, guruId string) (*entity.Guru, error) {
	id, err := strconv.Atoi(guruId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT g.guru_id::text,
		       COALESCE(g.user_id::text, '') AS user_id,
		       COALESCE(g.sekolah_id::text, '') AS sekolah_id,
		       COALESCE(g.nip, '') AS nip,
		       COALESCE(g.bidang_ajar, '') AS bidang_ajar,
		       COALESCE(u.username, '') AS username,
		       COALESCE(u.email, '') AS email
		FROM guru g
		LEFT JOIN users u ON g.user_id = u.user_id
		WHERE g.guru_id = $1
	`
	rows, err := r.DB.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error query FindById guru: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Guru])
	if err != nil {
		r.Log.Errorf("Error collecting row FindById guru: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *guruRepositoryImpl) FindByUserId(ctx context.Context, userId string) (*entity.Guru, error) {
	id, err := strconv.Atoi(userId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT g.guru_id::text,
		       COALESCE(g.user_id::text, '') AS user_id,
		       COALESCE(g.sekolah_id::text, '') AS sekolah_id,
		       COALESCE(g.nip, '') AS nip,
		       COALESCE(g.bidang_ajar, '') AS bidang_ajar
		FROM guru g
		WHERE g.user_id = $1
	`
	rows, err := r.DB.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error query FindByUserId guru: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Guru])
	if err != nil {
		// ErrNoRows means the guru profile hasn't been created yet — not a server error.
		if !errors.Is(err, pgx.ErrNoRows) {
			r.Log.Errorf("Error collecting row FindByUserId guru: %v", err)
		}
		return nil, err
	}
	return result, nil
}

func (r *guruRepositoryImpl) FindAll(ctx context.Context, search string, limit int, offset int) ([]*entity.Guru, int, error) {
	searchPattern := "%" + search + "%"

	countQuery := `
		SELECT COUNT(*)
		FROM guru g
		LEFT JOIN users u ON g.user_id = u.user_id
		WHERE (COALESCE(g.nip, '') ILIKE $1 OR COALESCE(u.username, '') ILIKE $1)
	`
	var total int
	err := r.DB.QueryRow(ctx, countQuery, searchPattern).Scan(&total)
	if err != nil {
		r.Log.Errorf("Error counting FindAll guru: %v", err)
		return nil, 0, err
	}

	query := `
		SELECT g.guru_id::text,
		       COALESCE(g.user_id::text, '') AS user_id,
		       COALESCE(g.sekolah_id::text, '') AS sekolah_id,
		       COALESCE(g.nip, '') AS nip,
		       COALESCE(g.bidang_ajar, '') AS bidang_ajar,
		       COALESCE(u.username, '') AS username,
		       COALESCE(u.email, '') AS email
		FROM guru g
		LEFT JOIN users u ON g.user_id = u.user_id
		WHERE (COALESCE(g.nip, '') ILIKE $1 OR COALESCE(u.username, '') ILIKE $1)
		ORDER BY g.guru_id ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.DB.Query(ctx, query, searchPattern, limit, offset)
	if err != nil {
		r.Log.Errorf("Error query FindAll guru: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	gurus, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.Guru])
	if err != nil {
		r.Log.Errorf("Error collecting rows FindAll guru: %v", err)
		return nil, 0, err
	}
	return gurus, total, nil
}

func (r *guruRepositoryImpl) Update(ctx context.Context, guru *entity.Guru) (*entity.Guru, error) {
	id, err := strconv.Atoi(guru.GuruId)
	if err != nil {
		return nil, err
	}

	var sekolahId *int
	if guru.SekolahId != "" {
		sid, err := strconv.Atoi(guru.SekolahId)
		if err != nil {
			return nil, err
		}
		sekolahId = &sid
	}

	query := `
		WITH updated AS (
			UPDATE guru
			SET nip = $1, bidang_ajar = $2, sekolah_id = $3
			WHERE guru_id = $4
			RETURNING guru_id, user_id, sekolah_id, nip, bidang_ajar
		)
		SELECT u_data.guru_id::text,
		       COALESCE(u_data.user_id::text, '') AS user_id,
		       COALESCE(u_data.sekolah_id::text, '') AS sekolah_id,
		       COALESCE(u_data.nip, '') AS nip,
		       COALESCE(u_data.bidang_ajar, '') AS bidang_ajar,
		       COALESCE(usr.username, '') AS username,
		       COALESCE(usr.email, '') AS email
		FROM updated u_data
		LEFT JOIN users usr ON u_data.user_id = usr.user_id
	`
	rows, err := r.DB.Query(ctx, query, guru.NIP, guru.BidangAjar, sekolahId, id)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "nip") {
				return nil, errNIPAlreadyUsed
			}
		}
		r.Log.Errorf("Error Update guru: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Guru])
	if err != nil {
		r.Log.Errorf("Error collecting row Update guru: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *guruRepositoryImpl) Delete(ctx context.Context, guruId string) error {
	id, err := strconv.Atoi(guruId)
	if err != nil {
		return err
	}

	query := `DELETE FROM guru WHERE guru_id = $1`
	_, err = r.DB.Exec(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error Delete guru: %v", err)
		return err
	}
	return nil
}

func (r *guruRepositoryImpl) FindSekolahByGuruId(ctx context.Context, guruId string) (*entity.Sekolah, error) {
	id, err := strconv.Atoi(guruId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT s.sekolah_id::text, s.nama_sekolah, s.alamat_sekolah
		FROM sekolah s
		JOIN guru g ON s.sekolah_id = g.sekolah_id
		WHERE g.guru_id = $1
	`
	rows, err := r.DB.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error query FindSekolahByGuruId: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Sekolah])
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *guruRepositoryImpl) FindGroupsByGuruId(ctx context.Context, guruId string) ([]*entity.Group, error) {
	id, err := strconv.Atoi(guruId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT group_id::text, group_name
		FROM groups
		WHERE created_by = $1
		ORDER BY created_at ASC
	`
	rows, err := r.DB.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error query FindGroupsByGuruId: %v", err)
		return nil, err
	}
	defer rows.Close()

	groups, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.Group])
	if err != nil {
		r.Log.Errorf("Error collecting rows FindGroupsByGuruId: %v", err)
		return nil, err
	}
	return groups, nil
}

var errNIPAlreadyUsed = &pgconn.PgError{Code: "23505", ConstraintName: "guru_nip_key"}
var errUserAlreadyGuru = &pgconn.PgError{Code: "23505", ConstraintName: "guru_user_id_key"}

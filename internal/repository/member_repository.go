package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type MemberRepository interface {
	Create(ctx context.Context, member *entity.Member) (*entity.Member, error)
	FindById(ctx context.Context, memberId string) (*entity.Member, error)
	FindByUserId(ctx context.Context, userId string) (*entity.Member, error)
	FindAll(ctx context.Context, search string, limit int, offset int) ([]*entity.Member, int, error)
	Update(ctx context.Context, member *entity.Member) (*entity.Member, error)
	Delete(ctx context.Context, memberId string) error
	FindSekolahByMemberId(ctx context.Context, memberId string) (*entity.Sekolah, error)
}

type memberRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewMemberRepository(db *pgxpool.Pool, log *logrus.Logger) MemberRepository {
	return &memberRepositoryImpl{
		DB:  db,
		Log: log,
	}
}

func (r *memberRepositoryImpl) Create(ctx context.Context, member *entity.Member) (*entity.Member, error) {
	userId, err := strconv.Atoi(member.UserId)
	if err != nil {
		return nil, err
	}

	var sekolahId *int
	if member.SekolahId != "" {
		sid, err := strconv.Atoi(member.SekolahId)
		if err != nil {
			return nil, err
		}
		sekolahId = &sid
	}

	query := `
		WITH inserted AS (
			INSERT INTO members (user_id, sekolah_id, nis)
			VALUES ($1, $2, $3)
			RETURNING member_id, user_id, sekolah_id, nis, total_xp, level,
			          foto_profil, bio, tanggal_lahir, jenis_kelamin, minat, last_active
		)
		SELECT i.member_id::text,
		       COALESCE(i.user_id::text, '') AS user_id,
		       COALESCE(i.sekolah_id::text, '') AS sekolah_id,
		       COALESCE(i.nis, '') AS nis,
		       i.total_xp,
		       i.level,
		       COALESCE(i.foto_profil, '') AS foto_profil,
		       COALESCE(i.bio, '') AS bio,
		       COALESCE(i.tanggal_lahir::text, '') AS tanggal_lahir,
		       COALESCE(i.jenis_kelamin::text, '') AS jenis_kelamin,
		       COALESCE(i.minat, '') AS minat,
		       COALESCE(i.last_active::text, '') AS last_active,
		       COALESCE(u.username, '') AS username,
		       COALESCE(u.email, '') AS email
		FROM inserted i
		JOIN users u ON i.user_id = u.user_id
	`

	rows, err := r.DB.Query(ctx, query, userId, sekolahId, member.NIS)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "user_id") {
				return nil, errUserAlreadyMember
			}
		}
		r.Log.Errorf("Error Create member: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Member])
	if err != nil {
		r.Log.Errorf("Error collecting row Create member: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *memberRepositoryImpl) FindById(ctx context.Context, memberId string) (*entity.Member, error) {
	id, err := strconv.Atoi(memberId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT m.member_id::text,
		       COALESCE(m.user_id::text, '') AS user_id,
		       COALESCE(m.sekolah_id::text, '') AS sekolah_id,
		       COALESCE(m.nis, '') AS nis,
		       m.total_xp,
		       m.level,
		       COALESCE(m.foto_profil, '') AS foto_profil,
		       COALESCE(m.bio, '') AS bio,
		       COALESCE(m.tanggal_lahir::text, '') AS tanggal_lahir,
		       COALESCE(m.jenis_kelamin::text, '') AS jenis_kelamin,
		       COALESCE(m.minat, '') AS minat,
		       COALESCE(m.last_active::text, '') AS last_active,
		       COALESCE(u.username, '') AS username,
		       COALESCE(u.email, '') AS email
		FROM members m
		LEFT JOIN users u ON m.user_id = u.user_id
		WHERE m.member_id = $1
	`
	rows, err := r.DB.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error query FindById member: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Member])
	if err != nil {
		r.Log.Errorf("Error collecting row FindById member: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *memberRepositoryImpl) FindByUserId(ctx context.Context, userId string) (*entity.Member, error) {
	id, err := strconv.Atoi(userId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT m.member_id::text,
		       COALESCE(m.user_id::text, '') AS user_id,
		       COALESCE(m.sekolah_id::text, '') AS sekolah_id,
		       COALESCE(m.nis, '') AS nis,
		       m.total_xp,
		       m.level,
		       COALESCE(m.foto_profil, '') AS foto_profil,
		       COALESCE(m.bio, '') AS bio,
		       COALESCE(m.tanggal_lahir::text, '') AS tanggal_lahir,
		       COALESCE(m.jenis_kelamin::text, '') AS jenis_kelamin,
		       COALESCE(m.minat, '') AS minat,
		       COALESCE(m.last_active::text, '') AS last_active
		FROM members m
		WHERE m.user_id = $1
	`
	rows, err := r.DB.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error query FindByUserId member: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Member])
	if err != nil {
		r.Log.Errorf("Error collecting row FindByUserId member: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *memberRepositoryImpl) FindAll(ctx context.Context, search string, limit int, offset int) ([]*entity.Member, int, error) {
	searchPattern := "%" + search + "%"

	countQuery := `
		SELECT COUNT(*)
		FROM members m
		LEFT JOIN users u ON m.user_id = u.user_id
		WHERE (COALESCE(m.nis, '') ILIKE $1 OR COALESCE(u.username, '') ILIKE $1)
	`
	var total int
	err := r.DB.QueryRow(ctx, countQuery, searchPattern).Scan(&total)
	if err != nil {
		r.Log.Errorf("Error counting FindAll member: %v", err)
		return nil, 0, err
	}

	query := `
		SELECT m.member_id::text,
		       COALESCE(m.user_id::text, '') AS user_id,
		       COALESCE(m.sekolah_id::text, '') AS sekolah_id,
		       COALESCE(m.nis, '') AS nis,
		       m.total_xp,
		       m.level,
		       COALESCE(m.foto_profil, '') AS foto_profil,
		       COALESCE(m.bio, '') AS bio,
		       COALESCE(m.tanggal_lahir::text, '') AS tanggal_lahir,
		       COALESCE(m.jenis_kelamin::text, '') AS jenis_kelamin,
		       COALESCE(m.minat, '') AS minat,
		       COALESCE(m.last_active::text, '') AS last_active,
		       COALESCE(u.username, '') AS username,
		       COALESCE(u.email, '') AS email
		FROM members m
		LEFT JOIN users u ON m.user_id = u.user_id
		WHERE (COALESCE(m.nis, '') ILIKE $1 OR COALESCE(u.username, '') ILIKE $1)
		ORDER BY m.member_id ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.DB.Query(ctx, query, searchPattern, limit, offset)
	if err != nil {
		r.Log.Errorf("Error query FindAll member: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	members, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.Member])
	if err != nil {
		r.Log.Errorf("Error collecting rows FindAll member: %v", err)
		return nil, 0, err
	}
	return members, total, nil
}

func (r *memberRepositoryImpl) Update(ctx context.Context, member *entity.Member) (*entity.Member, error) {
	id, err := strconv.Atoi(member.MemberId)
	if err != nil {
		return nil, err
	}

	var sekolahId *int
	if member.SekolahId != "" {
		sid, err := strconv.Atoi(member.SekolahId)
		if err != nil {
			return nil, err
		}
		sekolahId = &sid
	}

	var tanggalLahir *string
	if member.TanggalLahir != "" {
		tanggalLahir = &member.TanggalLahir
	}

	var jenisKelamin *string
	if member.JenisKelamin != "" {
		jk := string(member.JenisKelamin)
		jenisKelamin = &jk
	}

	query := `
		WITH updated AS (
			UPDATE members
			SET sekolah_id = $1,
			    nis = $2,
			    foto_profil = $3,
			    bio = $4,
			    tanggal_lahir = $5,
			    jenis_kelamin = $6,
			    minat = $7
			WHERE member_id = $8
			RETURNING member_id, user_id, sekolah_id, nis, total_xp, level,
			          foto_profil, bio, tanggal_lahir, jenis_kelamin, minat, last_active
		)
		SELECT u_data.member_id::text,
		       COALESCE(u_data.user_id::text, '') AS user_id,
		       COALESCE(u_data.sekolah_id::text, '') AS sekolah_id,
		       COALESCE(u_data.nis, '') AS nis,
		       u_data.total_xp,
		       u_data.level,
		       COALESCE(u_data.foto_profil, '') AS foto_profil,
		       COALESCE(u_data.bio, '') AS bio,
		       COALESCE(u_data.tanggal_lahir::text, '') AS tanggal_lahir,
		       COALESCE(u_data.jenis_kelamin::text, '') AS jenis_kelamin,
		       COALESCE(u_data.minat, '') AS minat,
		       COALESCE(u_data.last_active::text, '') AS last_active,
		       COALESCE(usr.username, '') AS username,
		       COALESCE(usr.email, '') AS email
		FROM updated u_data
		LEFT JOIN users usr ON u_data.user_id = usr.user_id
	`
	rows, err := r.DB.Query(ctx, query, sekolahId, member.NIS, member.FotoProfil, member.Bio, tanggalLahir, jenisKelamin, member.Minat, id)
	if err != nil {
		r.Log.Errorf("Error Update member: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Member])
	if err != nil {
		r.Log.Errorf("Error collecting row Update member: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *memberRepositoryImpl) Delete(ctx context.Context, memberId string) error {
	id, err := strconv.Atoi(memberId)
	if err != nil {
		return err
	}

	query := `DELETE FROM members WHERE member_id = $1`
	_, err = r.DB.Exec(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error Delete member: %v", err)
		return err
	}
	return nil
}

func (r *memberRepositoryImpl) FindSekolahByMemberId(ctx context.Context, memberId string) (*entity.Sekolah, error) {
	id, err := strconv.Atoi(memberId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT s.sekolah_id::text, s.nama_sekolah, s.alamat_sekolah
		FROM sekolah s
		JOIN members m ON s.sekolah_id = m.sekolah_id
		WHERE m.member_id = $1
	`
	rows, err := r.DB.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error query FindSekolahByMemberId: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Sekolah])
	if err != nil {
		return nil, err
	}
	return result, nil
}

var errUserAlreadyMember = &pgconn.PgError{Code: "23505", ConstraintName: "members_user_id_key"}

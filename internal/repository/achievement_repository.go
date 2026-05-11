package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type AchievementRepository interface {
	Create(ctx context.Context, ach *entity.Achievement) (*entity.Achievement, error)
	FindById(ctx context.Context, achievementId string) (*entity.Achievement, error)
	FindByNama(ctx context.Context, nama string) (*entity.Achievement, error)
	FindAll(ctx context.Context, search string, tier string, limit int, offset int) ([]*entity.Achievement, int, error)
	Update(ctx context.Context, ach *entity.Achievement) (*entity.Achievement, error)
	Delete(ctx context.Context, achievementId string) error
	CountMembersAwarded(ctx context.Context, achievementId string) (int, error)
}

type achievementRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewAchievementRepository(db *pgxpool.Pool, log *logrus.Logger) AchievementRepository {
	return &achievementRepositoryImpl{
		DB:  db,
		Log: log,
	}
}

func (r *achievementRepositoryImpl) Create(ctx context.Context, ach *entity.Achievement) (*entity.Achievement, error) {
	query := `
		WITH inserted AS (
			INSERT INTO achievements (nama, deskripsi, badge_icon, xp_required, tier)
			VALUES ($1, $2, $3, $4, $5::tier_achievement_enum)
			RETURNING achievement_id, nama, deskripsi, badge_icon, xp_required, tier
		)
		SELECT achievement_id::text,
		       nama,
		       COALESCE(deskripsi, '') AS deskripsi,
		       badge_icon,
		       xp_required,
		       tier::text AS tier
		FROM inserted
	`
	rows, err := r.DB.Query(ctx, query, ach.Nama, ach.Deskripsi, ach.BadgeIcon, ach.XPRequired, string(ach.Tier))
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, errAchievementNamaDuplicate
		}
		r.Log.Errorf("Error Create achievement: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Achievement])
	if err != nil {
		r.Log.Errorf("Error collecting row Create achievement: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *achievementRepositoryImpl) FindById(ctx context.Context, achievementId string) (*entity.Achievement, error) {
	id, err := strconv.Atoi(achievementId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT achievement_id::text,
		       nama,
		       COALESCE(deskripsi, '') AS deskripsi,
		       badge_icon,
		       xp_required,
		       tier::text AS tier
		FROM achievements
		WHERE achievement_id = $1
	`
	rows, err := r.DB.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error query FindById achievement: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Achievement])
	if err != nil {
		r.Log.Errorf("Error collecting row FindById achievement: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *achievementRepositoryImpl) FindByNama(ctx context.Context, nama string) (*entity.Achievement, error) {
	query := `
		SELECT achievement_id::text,
		       nama,
		       COALESCE(deskripsi, '') AS deskripsi,
		       badge_icon,
		       xp_required,
		       tier::text AS tier
		FROM achievements
		WHERE nama = $1
	`
	rows, err := r.DB.Query(ctx, query, nama)
	if err != nil {
		r.Log.Errorf("Error query FindByNama achievement: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Achievement])
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *achievementRepositoryImpl) FindAll(ctx context.Context, search string, tier string, limit int, offset int) ([]*entity.Achievement, int, error) {
	searchPattern := "%" + search + "%"

	var total int
	var err error

	if tier != "" {
		countQuery := `
			SELECT COUNT(*)
			FROM achievements
			WHERE nama ILIKE $1
			  AND tier::text = $2
		`
		err = r.DB.QueryRow(ctx, countQuery, searchPattern, tier).Scan(&total)
	} else {
		countQuery := `
			SELECT COUNT(*)
			FROM achievements
			WHERE nama ILIKE $1
		`
		err = r.DB.QueryRow(ctx, countQuery, searchPattern).Scan(&total)
	}
	if err != nil {
		r.Log.Errorf("Error counting FindAll achievement: %v", err)
		return nil, 0, err
	}

	var rows pgx.Rows
	if tier != "" {
		query := `
			SELECT achievement_id::text,
			       nama,
			       COALESCE(deskripsi, '') AS deskripsi,
			       badge_icon,
			       xp_required,
			       tier::text AS tier
			FROM achievements
			WHERE nama ILIKE $1
			  AND tier::text = $2
			ORDER BY xp_required ASC, achievement_id ASC
			LIMIT $3 OFFSET $4
		`
		rows, err = r.DB.Query(ctx, query, searchPattern, tier, limit, offset)
	} else {
		query := `
			SELECT achievement_id::text,
			       nama,
			       COALESCE(deskripsi, '') AS deskripsi,
			       badge_icon,
			       xp_required,
			       tier::text AS tier
			FROM achievements
			WHERE nama ILIKE $1
			ORDER BY xp_required ASC, achievement_id ASC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.DB.Query(ctx, query, searchPattern, limit, offset)
	}
	if err != nil {
		r.Log.Errorf("Error query FindAll achievement: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	achievements, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.Achievement])
	if err != nil {
		r.Log.Errorf("Error collecting rows FindAll achievement: %v", err)
		return nil, 0, err
	}
	return achievements, total, nil
}

func (r *achievementRepositoryImpl) Update(ctx context.Context, ach *entity.Achievement) (*entity.Achievement, error) {
	id, err := strconv.Atoi(ach.AchievementId)
	if err != nil {
		return nil, err
	}

	query := `
		WITH updated AS (
			UPDATE achievements
			SET nama        = $1,
			    deskripsi   = $2,
			    badge_icon  = $3,
			    xp_required = $4,
			    tier        = $5::tier_achievement_enum
			WHERE achievement_id = $6
			RETURNING achievement_id, nama, deskripsi, badge_icon, xp_required, tier
		)
		SELECT achievement_id::text,
		       nama,
		       COALESCE(deskripsi, '') AS deskripsi,
		       badge_icon,
		       xp_required,
		       tier::text AS tier
		FROM updated
	`
	rows, err := r.DB.Query(ctx, query, ach.Nama, ach.Deskripsi, ach.BadgeIcon, ach.XPRequired, string(ach.Tier), id)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, errAchievementNamaDuplicate
		}
		r.Log.Errorf("Error Update achievement: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Achievement])
	if err != nil {
		r.Log.Errorf("Error collecting row Update achievement: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *achievementRepositoryImpl) Delete(ctx context.Context, achievementId string) error {
	id, err := strconv.Atoi(achievementId)
	if err != nil {
		return err
	}

	query := `DELETE FROM achievements WHERE achievement_id = $1`
	_, err = r.DB.Exec(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error Delete achievement: %v", err)
		return err
	}
	return nil
}

func (r *achievementRepositoryImpl) CountMembersAwarded(ctx context.Context, achievementId string) (int, error) {
	id, err := strconv.Atoi(achievementId)
	if err != nil {
		return 0, err
	}

	var count int
	query := `SELECT COUNT(*) FROM member_achievements WHERE achievement_id = $1`
	err = r.DB.QueryRow(ctx, query, id).Scan(&count)
	if err != nil {
		r.Log.Errorf("Error CountMembersAwarded achievement: %v", err)
		return 0, err
	}
	return count, nil
}

var errAchievementNamaDuplicate = &pgconn.PgError{Code: "23505", ConstraintName: "uq_achievements_nama"}

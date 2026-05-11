package repository

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
)

var (
	ErrMemberAchievementNotFound = errors.New("member achievement not found")
	ErrMemberAchievementExists   = errors.New("member achievement already exists")
	ErrAchievementFKViolation    = errors.New("achievement not found")
)

type MemberAchievementRepository interface {
	FindAllByMemberId(ctx context.Context, memberId string) ([]*entity.MemberAchievement, error)
	FindOne(ctx context.Context, memberId, achievementId string) (*entity.MemberAchievement, error)
	Exists(ctx context.Context, memberId, achievementId string) (bool, error)
	Create(ctx context.Context, memberId, achievementId string) (*entity.MemberAchievement, error)
	Delete(ctx context.Context, memberId, achievementId string) error
}

type memberAchievementRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewMemberAchievementRepository(db *pgxpool.Pool, log *logrus.Logger) MemberAchievementRepository {
	return &memberAchievementRepositoryImpl{DB: db, Log: log}
}

const memberAchievementJoinSelect = `
	SELECT a.achievement_id::text,
	       a.nama,
	       COALESCE(a.deskripsi, '') AS deskripsi,
	       a.badge_icon,
	       a.xp_required,
	       a.tier::text AS tier,
	       ma.unlocked_at::text AS unlocked_at
	FROM member_achievements ma
	JOIN achievements a ON a.achievement_id = ma.achievement_id
`

func (r *memberAchievementRepositoryImpl) FindAllByMemberId(ctx context.Context, memberId string) ([]*entity.MemberAchievement, error) {
	id, err := strconv.Atoi(memberId)
	if err != nil {
		return nil, err
	}

	query := memberAchievementJoinSelect + `WHERE ma.member_id = $1 ORDER BY ma.unlocked_at DESC`
	rows, err := r.DB.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error FindAllByMemberId achievements: %v", err)
		return nil, err
	}
	defer rows.Close()

	achievements, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.MemberAchievement])
	if err != nil {
		r.Log.Errorf("Error collecting rows FindAllByMemberId achievements: %v", err)
		return nil, err
	}
	return achievements, nil
}

func (r *memberAchievementRepositoryImpl) FindOne(ctx context.Context, memberId, achievementId string) (*entity.MemberAchievement, error) {
	mid, err := strconv.Atoi(memberId)
	if err != nil {
		return nil, ErrMemberAchievementNotFound
	}
	aid, err := strconv.Atoi(achievementId)
	if err != nil {
		return nil, ErrMemberAchievementNotFound
	}

	query := memberAchievementJoinSelect + `WHERE ma.member_id = $1 AND ma.achievement_id = $2`
	rows, err := r.DB.Query(ctx, query, mid, aid)
	if err != nil {
		r.Log.Errorf("Error FindOne member achievement: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.MemberAchievement])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMemberAchievementNotFound
		}
		r.Log.Errorf("Error collecting row FindOne member achievement: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *memberAchievementRepositoryImpl) Exists(ctx context.Context, memberId, achievementId string) (bool, error) {
	mid, err := strconv.Atoi(memberId)
	if err != nil {
		return false, err
	}
	aid, err := strconv.Atoi(achievementId)
	if err != nil {
		return false, err
	}

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM member_achievements WHERE member_id = $1 AND achievement_id = $2)`
	err = r.DB.QueryRow(ctx, query, mid, aid).Scan(&exists)
	if err != nil {
		r.Log.Errorf("Error Exists member achievement: %v", err)
		return false, err
	}
	return exists, nil
}

func (r *memberAchievementRepositoryImpl) Create(ctx context.Context, memberId, achievementId string) (*entity.MemberAchievement, error) {
	mid, err := strconv.Atoi(memberId)
	if err != nil {
		return nil, ErrMemberAchievementNotFound
	}
	aid, err := strconv.Atoi(achievementId)
	if err != nil {
		return nil, ErrAchievementFKViolation
	}

	query := `
		WITH ins AS (
			INSERT INTO member_achievements (member_id, achievement_id)
			VALUES ($1, $2)
			RETURNING member_id, achievement_id
		)
		SELECT a.achievement_id::text,
		       a.nama,
		       COALESCE(a.deskripsi, '') AS deskripsi,
		       a.badge_icon,
		       a.xp_required,
		       a.tier::text AS tier,
		       ma.unlocked_at::text AS unlocked_at
		FROM ins
		JOIN achievements a ON a.achievement_id = ins.achievement_id
		JOIN member_achievements ma ON ma.member_id = ins.member_id AND ma.achievement_id = ins.achievement_id
	`
	rows, err := r.DB.Query(ctx, query, mid, aid)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case "23505":
				return nil, ErrMemberAchievementExists
			case "23503":
				if pgErr.ConstraintName == "member_achievements_member_id_fkey" {
					r.Log.Errorf("FK violation on member_id in member_achievements — JWT references a nonexistent member: %v", err)
					return nil, err
				}
				return nil, ErrAchievementFKViolation
			}
		}
		r.Log.Errorf("Error Create member achievement: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.MemberAchievement])
	if err != nil {
		r.Log.Errorf("Error collecting row Create member achievement: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *memberAchievementRepositoryImpl) Delete(ctx context.Context, memberId, achievementId string) error {
	mid, err := strconv.Atoi(memberId)
	if err != nil {
		return ErrMemberAchievementNotFound
	}
	aid, err := strconv.Atoi(achievementId)
	if err != nil {
		return ErrMemberAchievementNotFound
	}

	query := `DELETE FROM member_achievements WHERE member_id = $1 AND achievement_id = $2`
	cmd, err := r.DB.Exec(ctx, query, mid, aid)
	if err != nil {
		r.Log.Errorf("Error Delete member achievement: %v", err)
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrMemberAchievementNotFound
	}
	return nil
}

package repository

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
)

type MemberAchievementRepository interface {
	FindAllByMemberId(ctx context.Context, memberId string) ([]*entity.MemberAchievement, error)
}

type memberAchievementRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewMemberAchievementRepository(db *pgxpool.Pool, log *logrus.Logger) MemberAchievementRepository {
	return &memberAchievementRepositoryImpl{DB: db, Log: log}
}

func (r *memberAchievementRepositoryImpl) FindAllByMemberId(ctx context.Context, memberId string) ([]*entity.MemberAchievement, error) {
	id, err := strconv.Atoi(memberId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT a.achievement_id::text,
		       a.nama,
		       COALESCE(a.deskripsi, '') AS deskripsi,
		       a.badge_icon,
		       a.xp_required,
		       a.tier::text AS tier,
		       ma.unlocked_at::text AS unlocked_at
		FROM member_achievements ma
		JOIN achievements a ON a.achievement_id = ma.achievement_id
		WHERE ma.member_id = $1
		ORDER BY ma.unlocked_at DESC
	`
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

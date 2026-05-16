package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
)

type LeaderboardPeriod string

const (
	LeaderboardPeriodAlltime LeaderboardPeriod = "alltime"
	LeaderboardPeriodMonthly LeaderboardPeriod = "monthly"
)

type LeaderboardRepository interface {
	// FetchPublic returns (rows, total, periodStart). sekolahId == 0 means no filter.
	// periodStart is the boundary timestamp for monthly; zero-value time.Time for alltime.
	FetchPublic(ctx context.Context, period LeaderboardPeriod, sekolahId int, limit, offset int) ([]entity.LeaderboardEntry, int, time.Time, error)

	// FetchGroup returns (rows, total). The caller is responsible for
	// verifying the caller's right to read this group BEFORE calling this.
	FetchGroup(ctx context.Context, groupId string, limit, offset int) ([]entity.LeaderboardEntry, int, error)

	// GetGroupHeader returns (group_name, group_thumbnail_url, found).
	GetGroupHeader(ctx context.Context, groupId string) (string, *string, bool, error)
}

type leaderboardRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewLeaderboardRepository(db *pgxpool.Pool, log *logrus.Logger) LeaderboardRepository {
	return &leaderboardRepositoryImpl{DB: db, Log: log}
}

func (r *leaderboardRepositoryImpl) FetchPublic(ctx context.Context, period LeaderboardPeriod, sekolahId int, limit, offset int) ([]entity.LeaderboardEntry, int, time.Time, error) {
	var sekolahParam *int
	if sekolahId != 0 {
		sekolahParam = &sekolahId
	}

	if period == LeaderboardPeriodMonthly {
		return r.fetchPublicMonthly(ctx, sekolahParam, limit, offset)
	}
	return r.fetchPublicAlltime(ctx, sekolahParam, limit, offset)
}

func (r *leaderboardRepositoryImpl) fetchPublicAlltime(ctx context.Context, sekolahId *int, limit, offset int) ([]entity.LeaderboardEntry, int, time.Time, error) {
	// RANK() so ties share a rank (1, 1, 3, ...) — conventional leaderboard behavior.
	query := `
		SELECT
			m.member_id,
			u.username,
			m.foto_profil,
			m.level,
			m.total_xp,
			0::INT                                                       AS monthly_xp,
			0::INT                                                       AS group_xp,
			0::INT                                                       AS completed_count,
			s.sekolah_id,
			s.nama_sekolah,
			RANK() OVER (ORDER BY m.total_xp DESC, m.member_id ASC)     AS rank,
			COUNT(*) OVER ()                                             AS total_count
		FROM members m
		JOIN users u ON u.user_id = m.user_id AND u.role = 'member'
		LEFT JOIN sekolah s ON s.sekolah_id = m.sekolah_id
		WHERE ($1::INT IS NULL OR m.sekolah_id = $1)
		ORDER BY m.total_xp DESC, m.member_id ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.DB.Query(ctx, query, sekolahId, limit, offset)
	if err != nil {
		r.Log.Errorf("FetchPublic alltime: %v", err)
		return nil, 0, time.Time{}, err
	}
	defer rows.Close()

	entries, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[entity.LeaderboardEntry])
	if err != nil {
		r.Log.Errorf("FetchPublic alltime collect: %v", err)
		return nil, 0, time.Time{}, err
	}

	total := 0
	if len(entries) > 0 {
		total = entries[0].TotalCount
	}
	return entries, total, time.Time{}, nil
}

func (r *leaderboardRepositoryImpl) fetchPublicMonthly(ctx context.Context, sekolahId *int, limit, offset int) ([]entity.LeaderboardEntry, int, time.Time, error) {
	// month_start is computed once per query — every row uses the same boundary.
	// Asia/Jakarta (WIB, +07:00) so the month rolls at 00:00 Jakarta, not 00:00 UTC.
	// RANK() so ties share a rank (1, 1, 3, ...) — conventional leaderboard behavior.
	query := `
		WITH month_start AS (
			SELECT (date_trunc('month', NOW() AT TIME ZONE 'Asia/Jakarta')
			          AT TIME ZONE 'Asia/Jakarta') AS ts
		),
		agg AS (
			SELECT
				mp.member_id,
				SUM(mp.xp_reward) AS monthly_xp
			FROM member_progress mp, month_start
			WHERE mp.completed_at IS NOT NULL
			  AND mp.completed_at >= month_start.ts
			GROUP BY mp.member_id
			HAVING SUM(mp.xp_reward) > 0
		)
		SELECT
			m.member_id,
			u.username,
			m.foto_profil,
			m.level,
			m.total_xp,
			a.monthly_xp,
			0::INT                                                         AS group_xp,
			0::INT                                                         AS completed_count,
			s.sekolah_id,
			s.nama_sekolah,
			RANK() OVER (
				ORDER BY a.monthly_xp DESC, m.total_xp DESC, m.member_id ASC
			)                                                              AS rank,
			COUNT(*) OVER ()                                               AS total_count
		FROM agg a
		JOIN members m ON m.member_id = a.member_id
		JOIN users   u ON u.user_id   = m.user_id AND u.role = 'member'
		LEFT JOIN sekolah s ON s.sekolah_id = m.sekolah_id
		WHERE ($1::INT IS NULL OR m.sekolah_id = $1)
		ORDER BY a.monthly_xp DESC, m.total_xp DESC, m.member_id ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.DB.Query(ctx, query, sekolahId, limit, offset)
	if err != nil {
		r.Log.Errorf("FetchPublic monthly: %v", err)
		return nil, 0, time.Time{}, err
	}
	defer rows.Close()

	entries, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[entity.LeaderboardEntry])
	if err != nil {
		r.Log.Errorf("FetchPublic monthly collect: %v", err)
		return nil, 0, time.Time{}, err
	}

	// Compute period_start from DB so the boundary matches exactly what the query used.
	var periodStart time.Time
	err = r.DB.QueryRow(ctx, `
		SELECT date_trunc('month', NOW() AT TIME ZONE 'Asia/Jakarta')
		         AT TIME ZONE 'Asia/Jakarta'
	`).Scan(&periodStart)
	if err != nil {
		r.Log.Errorf("FetchPublic monthly period_start: %v", err)
		return nil, 0, time.Time{}, err
	}

	total := 0
	if len(entries) > 0 {
		total = entries[0].TotalCount
	}
	return entries, total, periodStart, nil
}

func (r *leaderboardRepositoryImpl) FetchGroup(ctx context.Context, groupId string, limit, offset int) ([]entity.LeaderboardEntry, int, error) {
	// LEFT JOIN member_progress so members with zero completions still appear.
	// RANK() so ties share a rank (1, 1, 3, ...) — conventional leaderboard behavior.
	query := `
		WITH agg AS (
			SELECT
				gm.member_id,
				COALESCE(SUM(mp.xp_reward), 0)                                    AS group_xp,
				COUNT(mp.progres_id) FILTER (WHERE mp.completed_at IS NOT NULL)   AS completed_count
			FROM group_members gm
			LEFT JOIN member_progress mp
				ON mp.member_id = gm.member_id
			   AND mp.group_id  = gm.group_id
			WHERE gm.group_id = $1
			GROUP BY gm.member_id
		)
		SELECT
			a.member_id,
			u.username,
			m.foto_profil,
			m.level,
			m.total_xp,
			0::INT                                                                AS monthly_xp,
			a.group_xp,
			a.completed_count,
			NULL::INT                                                             AS sekolah_id,
			NULL::TEXT                                                            AS nama_sekolah,
			RANK() OVER (
				ORDER BY a.group_xp DESC, a.completed_count DESC, m.total_xp DESC, a.member_id ASC
			)                                                                     AS rank,
			COUNT(*) OVER ()                                                      AS total_count
		FROM agg a
		JOIN members m ON m.member_id = a.member_id
		JOIN users   u ON u.user_id   = m.user_id
		ORDER BY a.group_xp DESC, a.completed_count DESC, m.total_xp DESC, a.member_id ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.DB.Query(ctx, query, groupId, limit, offset)
	if err != nil {
		r.Log.Errorf("FetchGroup: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	entries, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[entity.LeaderboardEntry])
	if err != nil {
		r.Log.Errorf("FetchGroup collect: %v", err)
		return nil, 0, err
	}

	total := 0
	if len(entries) > 0 {
		total = entries[0].TotalCount
	}
	return entries, total, nil
}

func (r *leaderboardRepositoryImpl) GetGroupHeader(ctx context.Context, groupId string) (string, *string, bool, error) {
	var name string
	var thumbnail *string
	err := r.DB.QueryRow(ctx, `
		SELECT g.group_name, a.url
		FROM groups g
		LEFT JOIN assets a ON a.asset_id = g.group_thumbnail_asset_id
		WHERE g.group_id = $1
	`, groupId).Scan(&name, &thumbnail)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil, false, nil
		}
		r.Log.Errorf("GetGroupHeader: %v", err)
		return "", nil, false, err
	}
	return name, thumbnail, true, nil
}

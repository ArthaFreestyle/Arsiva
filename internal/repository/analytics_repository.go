package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
)

type AnalyticsInterval string

const (
	AnalyticsIntervalDaily   AnalyticsInterval = "daily"
	AnalyticsIntervalWeekly  AnalyticsInterval = "weekly"
	AnalyticsIntervalMonthly AnalyticsInterval = "monthly"
)

const analyticsDateLayout = "2006-01-02"

var errInvalidAnalyticsInterval = errors.New("invalid analytics interval")

// sqlParts maps an interval to its date_trunc unit and the interval literal
// for one bucket. Fixed mapping — the caller's raw string never reaches SQL.
func (i AnalyticsInterval) sqlParts() (unit string, step string, ok bool) {
	switch i {
	case AnalyticsIntervalDaily:
		return "day", "1 day", true
	case AnalyticsIntervalWeekly:
		return "week", "1 week", true
	case AnalyticsIntervalMonthly:
		return "month", "1 month", true
	}
	return "", "", false
}

type AnalyticsRepository interface {
	CountUsersByRole(ctx context.Context) (entity.UserRoleCount, error)
	CountContent(ctx context.Context) (entity.ContentCount, error)

	// UserGrowth returns the users created before `from` (baseline) and the
	// non-empty buckets in [from, to + 1 bucket). from/to must already be
	// snapped to bucket starts; only their date part is used. Empty buckets
	// are not returned — the caller zero-fills.
	UserGrowth(ctx context.Context, interval AnalyticsInterval, from, to time.Time) (entity.UserRoleCount, []entity.UserGrowthBucket, error)

	// ContentGrowth is UserGrowth for kuis / cerita / puzzle / artikel.
	ContentGrowth(ctx context.Context, interval AnalyticsInterval, from, to time.Time) (entity.ContentTypeCount, []entity.ContentGrowthBucket, error)
}

type analyticsRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewAnalyticsRepository(db *pgxpool.Pool, log *logrus.Logger) AnalyticsRepository {
	return &analyticsRepositoryImpl{DB: db, Log: log}
}

// Soft-deleted users (is_active = FALSE) are excluded everywhere so the
// summary total and the last cumulative growth point always agree.
func (r *analyticsRepositoryImpl) CountUsersByRole(ctx context.Context) (entity.UserRoleCount, error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE role = 'super_admin'),
			COUNT(*) FILTER (WHERE role = 'guru'),
			COUNT(*) FILTER (WHERE role = 'member')
		FROM users
		WHERE is_active = TRUE`

	var c entity.UserRoleCount
	err := r.DB.QueryRow(ctx, query).Scan(&c.SuperAdmin, &c.Guru, &c.Member)
	return c, err
}

// kuis / cerita_interaktif / puzzles are hard-deleted and use is_published;
// artikel is soft-deleted (is_active) and uses status_enum.
func (r *analyticsRepositoryImpl) CountContent(ctx context.Context) (entity.ContentCount, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM kuis),
			(SELECT COUNT(*) FROM kuis WHERE is_published),
			(SELECT COUNT(*) FROM cerita_interaktif),
			(SELECT COUNT(*) FROM cerita_interaktif WHERE is_published),
			(SELECT COUNT(*) FROM puzzles),
			(SELECT COUNT(*) FROM puzzles WHERE is_published),
			(SELECT COUNT(*) FROM artikel WHERE is_active),
			(SELECT COUNT(*) FROM artikel WHERE is_active AND status = 'published')`

	var c entity.ContentCount
	err := r.DB.QueryRow(ctx, query).Scan(
		&c.KuisTotal, &c.KuisPublished,
		&c.CeritaTotal, &c.CeritaPublished,
		&c.PuzzleTotal, &c.PuzzlePublished,
		&c.ArtikelTotal, &c.ArtikelPublished,
	)
	return c, err
}

// created_at columns are TIMESTAMP without time zone holding Jakarta
// wall-clock time, so the raw column is bucketed as-is (no AT TIME ZONE) and
// the range bounds are passed as 'YYYY-MM-DD' text cast to ::timestamp —
// passing a time.Time would let pgx shift it through UTC.
func (r *analyticsRepositoryImpl) UserGrowth(ctx context.Context, interval AnalyticsInterval, from, to time.Time) (entity.UserRoleCount, []entity.UserGrowthBucket, error) {
	var baseline entity.UserRoleCount
	unit, step, ok := interval.sqlParts()
	if !ok {
		return baseline, nil, errInvalidAnalyticsInterval
	}
	fromStr, toStr := from.Format(analyticsDateLayout), to.Format(analyticsDateLayout)

	baselineQuery := `
		SELECT
			COUNT(*) FILTER (WHERE role = 'super_admin'),
			COUNT(*) FILTER (WHERE role = 'guru'),
			COUNT(*) FILTER (WHERE role = 'member')
		FROM users
		WHERE is_active = TRUE AND created_at < $1::timestamp`
	if err := r.DB.QueryRow(ctx, baselineQuery, fromStr).Scan(&baseline.SuperAdmin, &baseline.Guru, &baseline.Member); err != nil {
		return baseline, nil, err
	}

	// Half-open range [from, to + 1 bucket) — BETWEEN would drop late rows of the last bucket.
	bucketQuery := `
		SELECT
			date_trunc($3::text, created_at) AS bucket,
			COUNT(*) FILTER (WHERE role = 'super_admin'),
			COUNT(*) FILTER (WHERE role = 'guru'),
			COUNT(*) FILTER (WHERE role = 'member')
		FROM users
		WHERE is_active = TRUE
		  AND created_at >= $1::timestamp
		  AND created_at <  $2::timestamp + $4::interval
		GROUP BY 1
		ORDER BY 1`
	rows, err := r.DB.Query(ctx, bucketQuery, fromStr, toStr, unit, step)
	if err != nil {
		return baseline, nil, err
	}
	defer rows.Close()

	var buckets []entity.UserGrowthBucket
	for rows.Next() {
		var b entity.UserGrowthBucket
		if err := rows.Scan(&b.Bucket, &b.SuperAdmin, &b.Guru, &b.Member); err != nil {
			return baseline, nil, err
		}
		buckets = append(buckets, b)
	}
	return baseline, buckets, rows.Err()
}

// contentRowsCTE unions the creation time of every content row, applying the
// same filters as CountContent's totals (artikel excludes soft-deleted rows).
const contentRowsCTE = `
	WITH content_rows AS (
		SELECT 'kuis' AS type, created_at FROM kuis
		UNION ALL SELECT 'cerita', created_at FROM cerita_interaktif
		UNION ALL SELECT 'puzzle', created_at FROM puzzles
		UNION ALL SELECT 'artikel', created_at FROM artikel WHERE is_active = TRUE
	)`

func (r *analyticsRepositoryImpl) ContentGrowth(ctx context.Context, interval AnalyticsInterval, from, to time.Time) (entity.ContentTypeCount, []entity.ContentGrowthBucket, error) {
	var baseline entity.ContentTypeCount
	unit, step, ok := interval.sqlParts()
	if !ok {
		return baseline, nil, errInvalidAnalyticsInterval
	}
	fromStr, toStr := from.Format(analyticsDateLayout), to.Format(analyticsDateLayout)

	baselineQuery := contentRowsCTE + `
		SELECT
			COUNT(*) FILTER (WHERE type = 'kuis'),
			COUNT(*) FILTER (WHERE type = 'cerita'),
			COUNT(*) FILTER (WHERE type = 'puzzle'),
			COUNT(*) FILTER (WHERE type = 'artikel')
		FROM content_rows
		WHERE created_at < $1::timestamp`
	if err := r.DB.QueryRow(ctx, baselineQuery, fromStr).Scan(&baseline.Kuis, &baseline.Cerita, &baseline.Puzzle, &baseline.Artikel); err != nil {
		return baseline, nil, err
	}

	bucketQuery := contentRowsCTE + `
		SELECT
			date_trunc($3::text, created_at) AS bucket,
			COUNT(*) FILTER (WHERE type = 'kuis'),
			COUNT(*) FILTER (WHERE type = 'cerita'),
			COUNT(*) FILTER (WHERE type = 'puzzle'),
			COUNT(*) FILTER (WHERE type = 'artikel')
		FROM content_rows
		WHERE created_at >= $1::timestamp
		  AND created_at <  $2::timestamp + $4::interval
		GROUP BY 1
		ORDER BY 1`
	rows, err := r.DB.Query(ctx, bucketQuery, fromStr, toStr, unit, step)
	if err != nil {
		return baseline, nil, err
	}
	defer rows.Close()

	var buckets []entity.ContentGrowthBucket
	for rows.Next() {
		var b entity.ContentGrowthBucket
		if err := rows.Scan(&b.Bucket, &b.Kuis, &b.Cerita, &b.Puzzle, &b.Artikel); err != nil {
			return baseline, nil, err
		}
		buckets = append(buckets, b)
	}
	return baseline, buckets, rows.Err()
}

package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
)

type AnalyticsUseCase interface {
	GetUserSummary(ctx context.Context) (*model.UserSummaryResponse, error)
	GetContentSummary(ctx context.Context) (*model.ContentSummaryResponse, error)
	GetUserGrowth(ctx context.Context, req *model.GrowthRequest) (*model.UserGrowthResponse, error)
	GetContentGrowth(ctx context.Context, req *model.GrowthRequest) (*model.ContentGrowthResponse, error)
}

type analyticsUseCaseImpl struct {
	Repo repository.AnalyticsRepository
	Log  *logrus.Logger
	// Now is injectable so tests can pin "today".
	Now func() time.Time
}

func NewAnalyticsUseCase(repo repository.AnalyticsRepository, log *logrus.Logger) AnalyticsUseCase {
	return &analyticsUseCaseImpl{
		Repo: repo,
		Log:  log,
		Now:  time.Now,
	}
}

const analyticsDateLayout = "2006-01-02"

// Indonesia has no DST, so a fixed +07:00 zone is exact (same as the leaderboard).
var analyticsWIB = time.FixedZone("WIB", 7*3600)

// growthIntervalLimits: default number of buckets when `from` is omitted, and
// the maximum number of buckets a single request may span.
var growthIntervalLimits = map[repository.AnalyticsInterval]struct{ def, max int }{
	repository.AnalyticsIntervalDaily:   {def: 30, max: 92},
	repository.AnalyticsIntervalWeekly:  {def: 12, max: 104},
	repository.AnalyticsIntervalMonthly: {def: 12, max: 60},
}

func (u *analyticsUseCaseImpl) GetUserSummary(ctx context.Context) (*model.UserSummaryResponse, error) {
	counts, err := u.Repo.CountUsersByRole(ctx)
	if err != nil {
		u.Log.Warnf("GetUserSummary CountUsersByRole: %v", err)
		return nil, fiber.ErrInternalServerError
	}
	return converter.ToUserSummaryResponse(counts), nil
}

func (u *analyticsUseCaseImpl) GetContentSummary(ctx context.Context) (*model.ContentSummaryResponse, error) {
	counts, err := u.Repo.CountContent(ctx)
	if err != nil {
		u.Log.Warnf("GetContentSummary CountContent: %v", err)
		return nil, fiber.ErrInternalServerError
	}
	return converter.ToContentSummaryResponse(counts), nil
}

func (u *analyticsUseCaseImpl) GetUserGrowth(ctx context.Context, req *model.GrowthRequest) (*model.UserGrowthResponse, error) {
	rng, err := u.resolveGrowthRange(req)
	if err != nil {
		return nil, err
	}

	baseline, rows, err := u.Repo.UserGrowth(ctx, rng.interval, rng.from, rng.to)
	if err != nil {
		u.Log.Warnf("GetUserGrowth UserGrowth: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	byBucket := make(map[string]entity.UserRoleCount, len(rows))
	for _, r := range rows {
		byBucket[r.Bucket.Format(analyticsDateLayout)] = r.UserRoleCount
	}

	cumulative := baseline.SuperAdmin + baseline.Guru + baseline.Member
	series := make([]model.UserGrowthPoint, 0, len(rng.buckets))
	for _, b := range rng.buckets {
		key := b.Format(analyticsDateLayout)
		c := byBucket[key] // zero value for empty buckets
		newUsers := c.SuperAdmin + c.Guru + c.Member
		cumulative += newUsers
		series = append(series, model.UserGrowthPoint{
			PeriodStart: key,
			NewUsers:    newUsers,
			NewByRole:   converter.ToUserRoleCountResponse(c),
			Cumulative:  cumulative,
		})
	}

	return &model.UserGrowthResponse{
		Interval: string(rng.interval),
		From:     rng.from.Format(analyticsDateLayout),
		To:       rng.to.Format(analyticsDateLayout),
		Series:   series,
	}, nil
}

func (u *analyticsUseCaseImpl) GetContentGrowth(ctx context.Context, req *model.GrowthRequest) (*model.ContentGrowthResponse, error) {
	rng, err := u.resolveGrowthRange(req)
	if err != nil {
		return nil, err
	}

	baseline, rows, err := u.Repo.ContentGrowth(ctx, rng.interval, rng.from, rng.to)
	if err != nil {
		u.Log.Warnf("GetContentGrowth ContentGrowth: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	byBucket := make(map[string]entity.ContentTypeCount, len(rows))
	for _, r := range rows {
		byBucket[r.Bucket.Format(analyticsDateLayout)] = r.ContentTypeCount
	}

	cumulative := baseline
	series := make([]model.ContentGrowthPoint, 0, len(rng.buckets))
	for _, b := range rng.buckets {
		key := b.Format(analyticsDateLayout)
		c := byBucket[key]
		cumulative.Kuis += c.Kuis
		cumulative.Cerita += c.Cerita
		cumulative.Puzzle += c.Puzzle
		cumulative.Artikel += c.Artikel
		series = append(series, model.ContentGrowthPoint{
			PeriodStart: key,
			New:         converter.ToContentTypeCountResponse(c),
			Cumulative:  converter.ToContentTypeCountResponse(cumulative),
		})
	}

	return &model.ContentGrowthResponse{
		Interval: string(rng.interval),
		From:     rng.from.Format(analyticsDateLayout),
		To:       rng.to.Format(analyticsDateLayout),
		Series:   series,
	}, nil
}

// growthRange is a validated, bucket-aligned request range. Dates are
// calendar dates (midnight UTC) representing Jakarta days.
type growthRange struct {
	interval repository.AnalyticsInterval
	from     time.Time // first bucket start
	to       time.Time // last bucket start
	buckets  []time.Time
}

func (u *analyticsUseCaseImpl) resolveGrowthRange(req *model.GrowthRequest) (*growthRange, error) {
	interval := repository.AnalyticsInterval(req.Interval)
	if req.Interval == "" {
		interval = repository.AnalyticsIntervalDaily
	}
	limits, ok := growthIntervalLimits[interval]
	if !ok {
		return nil, fiber.NewError(fiber.StatusBadRequest, "interval tidak valid, gunakan 'daily', 'weekly', atau 'monthly'")
	}

	now := u.Now().In(analyticsWIB)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	to := today
	if req.To != "" {
		parsed, err := time.Parse(analyticsDateLayout, req.To)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "to tidak valid, gunakan format YYYY-MM-DD")
		}
		if parsed.Before(today) {
			to = parsed
		}
	}

	var from time.Time
	if req.From != "" {
		parsed, err := time.Parse(analyticsDateLayout, req.From)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "from tidak valid, gunakan format YYYY-MM-DD")
		}
		if to.Before(parsed) {
			return nil, fiber.NewError(fiber.StatusBadRequest, "to tidak boleh sebelum from")
		}
		from = snapToBucket(interval, parsed)
	} else {
		from = addBuckets(interval, snapToBucket(interval, to), -(limits.def - 1))
	}
	to = snapToBucket(interval, to)

	count := bucketCount(interval, from, to)
	if count > limits.max {
		return nil, fiber.NewError(fiber.StatusBadRequest,
			fmt.Sprintf("rentang terlalu panjang untuk interval %s (maks %d)", interval, limits.max))
	}

	buckets := make([]time.Time, count)
	for i := range buckets {
		buckets[i] = addBuckets(interval, from, i)
	}

	return &growthRange{interval: interval, from: from, to: to, buckets: buckets}, nil
}

// snapToBucket returns the start of the bucket containing d. Weeks start on
// Monday, matching Postgres date_trunc('week', ...).
func snapToBucket(interval repository.AnalyticsInterval, d time.Time) time.Time {
	switch interval {
	case repository.AnalyticsIntervalWeekly:
		daysSinceMonday := (int(d.Weekday()) + 6) % 7
		return d.AddDate(0, 0, -daysSinceMonday)
	case repository.AnalyticsIntervalMonthly:
		return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	return d
}

// addBuckets moves a bucket start by n buckets. Monthly starts are always the
// 1st, so AddDate never overflows into the following month.
func addBuckets(interval repository.AnalyticsInterval, d time.Time, n int) time.Time {
	switch interval {
	case repository.AnalyticsIntervalWeekly:
		return d.AddDate(0, 0, 7*n)
	case repository.AnalyticsIntervalMonthly:
		return d.AddDate(0, n, 0)
	}
	return d.AddDate(0, 0, n)
}

// bucketCount is the number of buckets from `from` to `to` inclusive; both
// must be bucket starts.
func bucketCount(interval repository.AnalyticsInterval, from, to time.Time) int {
	switch interval {
	case repository.AnalyticsIntervalWeekly:
		return int(to.Sub(from).Hours()/24)/7 + 1
	case repository.AnalyticsIntervalMonthly:
		return (to.Year()-from.Year())*12 + int(to.Month()-from.Month()) + 1
	}
	return int(to.Sub(from).Hours()/24) + 1
}

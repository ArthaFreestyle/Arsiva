package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/repository"
)

type ProgressSessionUseCase interface {
	Start(ctx context.Context, req *model.ProgressStartRequest, claims *model.Claims) (*model.ProgressStartResponse, error)
	Answer(ctx context.Context, req *model.ProgressAnswerRequest, claims *model.Claims) (*model.ProgressAnswerResponse, error)
	Scene(ctx context.Context, req *model.ProgressSceneRequest, claims *model.Claims) (*model.ProgressFinalizeResponse, error)
	Solve(ctx context.Context, req *model.ProgressSolveRequest, claims *model.Claims) (*model.ProgressFinalizeResponse, error)
	Submit(ctx context.Context, req *model.ProgressSubmitRequest, claims *model.Claims) (*model.ProgressFinalizeResponse, error)
	GetSession(ctx context.Context, contentType string, contentId int, claims *model.Claims) (*model.ProgressSessionResponse, error)

	// Finalize moves a Redis session into member_progress and optionally credits total_xp.
	// Safe to call multiple times — second call is a no-op returning the existing progres_id.
	// cause is "submit" or "expired" — used for logging only.
	Finalize(ctx context.Context, sessionKey string, cause string) (progresID int, err error)

	// ListExpiredSessionKeys returns session keys whose expiry <= now (used by the worker).
	ListExpiredSessionKeys(ctx context.Context) ([]string, error)
}

type progressSessionUseCaseImpl struct {
	Repo         repository.MemberProgressRepository
	Redis        *redis.Client
	Log          *logrus.Logger
	Validator    *validator.Validate
	Gamification GamificationUseCase
}

func NewProgressSessionUseCase(
	repo repository.MemberProgressRepository,
	redisClient *redis.Client,
	log *logrus.Logger,
	validate *validator.Validate,
	gamification GamificationUseCase,
) ProgressSessionUseCase {
	return &progressSessionUseCaseImpl{
		Repo:         repo,
		Redis:        redisClient,
		Log:          log,
		Validator:    validate,
		Gamification: gamification,
	}
}

// ─── Redis key helpers ───────────────────────────────────────────────────────

const progressActiveZSet = "progress:active"

func sessionKey(memberId, contentType string, contentId int) string {
	return fmt.Sprintf("progress:session:%s:%s:%d", memberId, contentType, contentId)
}

// ─── Start ───────────────────────────────────────────────────────────────────

func (u *progressSessionUseCaseImpl) Start(ctx context.Context, req *model.ProgressStartRequest, claims *model.Claims) (*model.ProgressStartResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Start: invalid request: %v", err)
		return nil, fiber.ErrBadRequest
	}

	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}

	// Validate content exists and is published.
	exists, err := u.Repo.CheckContentExists(ctx, req.ContentType, req.ContentId)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}
	if !exists {
		return nil, fiber.NewError(fiber.StatusNotFound, "konten tidak ditemukan atau belum dipublikasikan")
	}

	key := sessionKey(memberId, req.ContentType, req.ContentId)

	// Return 409 if an active session already exists for this triple.
	existing, err := u.Redis.HGet(ctx, key, "expires_at").Result()
	if err == nil && existing != "" {
		expiresAt, _ := strconv.ParseInt(existing, 10, 64)
		return nil, fiber.NewError(fiber.StatusConflict, fmt.Sprintf(
			"sesi aktif sudah ada; expires_at=%d", expiresAt,
		))
	}

	// Compute max_score once at session start.
	maxScore, err := u.computeMaxScore(ctx, req.ContentType, req.ContentId)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	now := time.Now().Unix()
	expiresAt := now + int64(req.DurationSeconds)
	ttl := time.Duration(req.DurationSeconds+1800) * time.Second

	answersJSON, _ := json.Marshal(map[string]int{})

	fields := map[string]any{
		"member_id":        memberId,
		"group_id":         req.GroupId,
		"content_type":     req.ContentType,
		"content_id":       req.ContentId,
		"started_at":       now,
		"expires_at":       expiresAt,
		"duration_seconds": req.DurationSeconds,
		"max_score":        maxScore,
		"running_skor":     0,
		"answers":          string(answersJSON),
		"state":            "active",
	}

	pipe := u.Redis.Pipeline()
	pipe.HMSet(ctx, key, fields)
	pipe.Expire(ctx, key, ttl)
	pipe.ZAdd(ctx, progressActiveZSet, redis.Z{Score: float64(expiresAt), Member: key})
	if _, err = pipe.Exec(ctx); err != nil {
		u.Log.Errorf("Start: redis pipeline: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return &model.ProgressStartResponse{
		SessionKey: key,
		ExpiresAt:  expiresAt,
		MaxScore:   maxScore,
	}, nil
}

func (u *progressSessionUseCaseImpl) computeMaxScore(ctx context.Context, contentType string, contentId int) (int, error) {
	switch contentType {
	case "kuis":
		return u.Repo.GetKuisMaxScore(ctx, contentId)
	case "cerita":
		return u.Repo.GetCeritaMaxScore(ctx, contentId)
	case "puzzle":
		return 1, nil
	default:
		return 0, fiber.ErrBadRequest
	}
}

// ─── Answer ──────────────────────────────────────────────────────────────────

func (u *progressSessionUseCaseImpl) Answer(ctx context.Context, req *model.ProgressAnswerRequest, claims *model.Claims) (*model.ProgressAnswerResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Answer: invalid request: %v", err)
		return nil, fiber.ErrBadRequest
	}

	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}

	key := sessionKey(memberId, req.ContentType, req.ContentId)
	session, err := u.loadActiveSession(ctx, key)
	if err != nil {
		return nil, err
	}

	// Validate pertanyaan belongs to this kuis.
	ok, err := u.Repo.CheckPertanyaanBelongsToKuis(ctx, req.PertanyaanId, req.ContentId)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}
	if !ok {
		return nil, fiber.NewError(fiber.StatusBadRequest, "pertanyaan tidak ditemukan dalam kuis ini")
	}

	// Idempotency: reject duplicate answers.
	pertanyaanKey := strconv.Itoa(req.PertanyaanId)
	if _, exists := session.Answers[pertanyaanKey]; exists {
		return nil, fiber.NewError(fiber.StatusConflict, "pertanyaan sudah dijawab")
	}

	score, err := u.Repo.GetJawabanScore(ctx, req.JawabanId, req.PertanyaanId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fiber.NewError(fiber.StatusBadRequest, "jawaban tidak valid")
		}
		return nil, fiber.ErrInternalServerError
	}

	session.Answers[pertanyaanKey] = score
	session.RunningSkor += score

	if err = u.persistAnswers(ctx, key, session); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return &model.ProgressAnswerResponse{RunningSkor: session.RunningSkor}, nil
}

// ─── Scene ───────────────────────────────────────────────────────────────────

func (u *progressSessionUseCaseImpl) Scene(ctx context.Context, req *model.ProgressSceneRequest, claims *model.Claims) (*model.ProgressFinalizeResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Scene: invalid request: %v", err)
		return nil, fiber.ErrBadRequest
	}

	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}

	key := sessionKey(memberId, req.ContentType, req.ContentId)
	if _, err := u.loadActiveSession(ctx, key); err != nil {
		return nil, err
	}

	// Validate scene belongs to this cerita.
	ok, err := u.Repo.CheckSceneBelongsToCerita(ctx, req.SceneId, req.ContentId)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}
	if !ok {
		return nil, fiber.NewError(fiber.StatusBadRequest, "scene tidak ditemukan dalam cerita ini")
	}

	isEnding, endingPoint, err := u.Repo.GetSceneEndingInfo(ctx, req.SceneId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fiber.NewError(fiber.StatusBadRequest, "scene tidak ditemukan")
		}
		return nil, fiber.ErrInternalServerError
	}

	// Record scene visit in Redis.
	sceneKey := strconv.Itoa(req.SceneId)
	pipe := u.Redis.Pipeline()
	pipe.HSet(ctx, key, fmt.Sprintf("answers.%s", sceneKey), endingPoint)
	if isEnding {
		pipe.HSet(ctx, key, "running_skor", endingPoint)
	}
	if _, err = pipe.Exec(ctx); err != nil {
		u.Log.Errorf("Scene: redis pipeline: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	if isEnding {
		progresId, err := u.Finalize(ctx, key, "submit")
		if err != nil {
			return nil, err
		}
		return &model.ProgressFinalizeResponse{ProgresId: progresId}, nil
	}

	return nil, nil
}

// ─── Solve ───────────────────────────────────────────────────────────────────

func (u *progressSessionUseCaseImpl) Solve(ctx context.Context, req *model.ProgressSolveRequest, claims *model.Claims) (*model.ProgressFinalizeResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Solve: invalid request: %v", err)
		return nil, fiber.ErrBadRequest
	}

	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}

	key := sessionKey(memberId, req.ContentType, req.ContentId)
	if _, err := u.loadActiveSession(ctx, key); err != nil {
		return nil, err
	}

	skor := 0
	if req.Solved {
		skor = 1
	}

	if err := u.Redis.HSet(ctx, key, "running_skor", skor).Err(); err != nil {
		u.Log.Errorf("Solve: redis set running_skor: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	progresId, err := u.Finalize(ctx, key, "submit")
	if err != nil {
		return nil, err
	}
	return &model.ProgressFinalizeResponse{ProgresId: progresId}, nil
}

// ─── Submit ──────────────────────────────────────────────────────────────────

func (u *progressSessionUseCaseImpl) Submit(ctx context.Context, req *model.ProgressSubmitRequest, claims *model.Claims) (*model.ProgressFinalizeResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Submit: invalid request: %v", err)
		return nil, fiber.ErrBadRequest
	}

	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}

	key := sessionKey(memberId, req.ContentType, req.ContentId)

	progresId, err := u.Finalize(ctx, key, "submit")
	if err != nil {
		return nil, err
	}
	return &model.ProgressFinalizeResponse{ProgresId: progresId}, nil
}

// ─── GetSession ──────────────────────────────────────────────────────────────

func (u *progressSessionUseCaseImpl) GetSession(ctx context.Context, contentType string, contentId int, claims *model.Claims) (*model.ProgressSessionResponse, error) {
	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}

	key := sessionKey(memberId, contentType, contentId)
	data, err := u.Redis.HGetAll(ctx, key).Result()
	if err != nil || len(data) == 0 {
		return nil, fiber.ErrNotFound
	}

	answers := map[string]int{}
	if raw, ok := data["answers"]; ok {
		_ = json.Unmarshal([]byte(raw), &answers)
	}

	contentIdInt, _ := strconv.Atoi(data["content_id"])
	expiresAt, _ := strconv.ParseInt(data["expires_at"], 10, 64)
	durationSeconds, _ := strconv.Atoi(data["duration_seconds"])
	maxScore, _ := strconv.Atoi(data["max_score"])
	runningSkor, _ := strconv.Atoi(data["running_skor"])

	return &model.ProgressSessionResponse{
		SessionKey:      key,
		MemberId:        data["member_id"],
		GroupId:         data["group_id"],
		ContentType:     data["content_type"],
		ContentId:       contentIdInt,
		ExpiresAt:       expiresAt,
		DurationSeconds: durationSeconds,
		MaxScore:        maxScore,
		RunningSkor:     runningSkor,
		State:           data["state"],
		Answers:         answers,
	}, nil
}

// claimSessionScript atomically transitions state active→flushing.
// Returns 1 if claimed, 0 if already flushing (another caller), -1 if gone/done.
var claimSessionScript = redis.NewScript(`
local state = redis.call('HGET', KEYS[1], 'state')
if state == 'active' then
    redis.call('HSET', KEYS[1], 'state', 'flushing')
    return 1
elseif state == 'flushing' then
    return 0
else
    return -1
end
`)

// ─── Finalize ────────────────────────────────────────────────────────────────

func (u *progressSessionUseCaseImpl) Finalize(ctx context.Context, key string, cause string) (int, error) {
	// Step 1: atomically transition state active→flushing via Lua script.
	result, err := claimSessionScript.Run(ctx, u.Redis, []string{key}).Int()
	if err != nil {
		u.Log.Errorf("Finalize claimSession script: %v", err)
		return 0, fiber.ErrInternalServerError
	}

	switch result {
	case -1:
		// Session gone or already fully flushed.
		return 0, nil
	case 0:
		// Another goroutine is mid-finalize. Wait briefly then read its progres_id.
		time.Sleep(100 * time.Millisecond)
		existing, err := u.Redis.HGet(ctx, key, "progres_id").Result()
		if err == nil && existing != "" {
			id, _ := strconv.Atoi(existing)
			return id, nil
		}
		// Session already cleaned up by the other goroutine.
		return 0, nil
	}
	// result == 1: we claimed the session.

	// Step 2: read session data.
	data, err := u.Redis.HGetAll(ctx, key).Result()
	if err != nil || len(data) == 0 {
		// Already flushed.
		return 0, nil
	}

	memberIdStr := data["member_id"]
	groupId := data["group_id"]
	contentType := data["content_type"]
	contentIdStr := data["content_id"]
	runningSkorStr := data["running_skor"]
	maxScoreStr := data["max_score"]
	durationSecondsStr := data["duration_seconds"]

	memberId, _ := strconv.Atoi(memberIdStr)
	contentId, _ := strconv.Atoi(contentIdStr)
	runningSkor, _ := strconv.Atoi(runningSkorStr)
	maxScore, _ := strconv.Atoi(maxScoreStr)
	durationSeconds, _ := strconv.Atoi(durationSecondsStr)

	// Step 3: fetch fresh xp_reward from content table (guru may have changed it mid-session).
	xpReward, err := u.Repo.GetContentXpReward(ctx, contentType, contentId)
	if err != nil {
		if u.Log != nil {
			u.Log.Errorf("Finalize GetContentXpReward (%s/%d): %v", contentType, contentId, err)
		}
		// Restore state so a retry can succeed.
		_ = u.Redis.HSet(ctx, key, "state", "active").Err()
		return 0, fiber.ErrInternalServerError
	}

	// Step 4: award XP only on perfect score.
	awardedXp := 0
	if runningSkor == maxScore {
		awardedXp = xpReward
	}

	progress := &entity.MemberProgress{
		MemberId:    memberId,
		GroupId:     groupId,
		ContentType: contentType,
		ContentId:   contentId,
		Skor:        runningSkor,
		Duration:    durationSeconds,
	}

	// Step 5: persist to Postgres in a single transaction.
	progresId, err := u.Repo.SaveProgress(ctx, progress, awardedXp)
	if err != nil {
		if u.Log != nil {
			u.Log.Errorf("Finalize SaveProgress (cause=%s key=%s): %v", cause, key, err)
		}
		// Leave the key intact with state=active so a retry can succeed.
		_ = u.Redis.HSet(ctx, key, "state", "active").Err()
		return 0, fiber.ErrInternalServerError
	}

	if u.Log != nil {
		u.Log.Infof("Finalize: %s progres_id=%d skor=%d/%d xp=%d (cause=%s)", key, progresId, runningSkor, maxScore, awardedXp, cause)
	}

	// Step 5b: update streak + daily-task progress. Must not fail the finalize —
	// progress is already committed, so we log any error and continue.
	if u.Gamification != nil {
		if err2 := u.Gamification.HandleContentFinished(ctx, memberId, contentType, awardedXp); err2 != nil {
			if u.Log != nil {
				u.Log.Errorf("Finalize HandleContentFinished (key=%s): %v", key, err2)
			}
		}
	}

	// Step 6: write progres_id back to Redis so a racing caller from step 1 can read it.
	_ = u.Redis.HSet(ctx, key, "progres_id", progresId).Err()

	// Step 7: clean up Redis.
	pipe := u.Redis.Pipeline()
	pipe.Del(ctx, key)
	pipe.ZRem(ctx, progressActiveZSet, key)
	_, _ = pipe.Exec(ctx)

	return progresId, nil
}

// ─── ListExpiredSessionKeys ───────────────────────────────────────────────────

func (u *progressSessionUseCaseImpl) ListExpiredSessionKeys(ctx context.Context) ([]string, error) {
	now := float64(time.Now().Unix())
	// Cap at 1000 keys per tick to avoid blocking.
	keys, err := u.Redis.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:     progressActiveZSet,
		Start:   "-inf",
		Stop:    fmt.Sprintf("%f", now),
		ByScore: true,
		Offset:  0,
		Count:   1000,
	}).Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

type sessionData struct {
	RunningSkor int
	MaxScore    int
	Answers     map[string]int
	State       string
}

func (u *progressSessionUseCaseImpl) loadActiveSession(ctx context.Context, key string) (*sessionData, error) {
	data, err := u.Redis.HGetAll(ctx, key).Result()
	if err != nil || len(data) == 0 {
		return nil, fiber.NewError(fiber.StatusNotFound, "sesi tidak ditemukan")
	}
	if data["state"] != "active" {
		return nil, fiber.NewError(fiber.StatusConflict, "sesi sedang diproses atau sudah selesai")
	}

	answers := map[string]int{}
	if raw, ok := data["answers"]; ok {
		_ = json.Unmarshal([]byte(raw), &answers)
	}

	runningSkor, _ := strconv.Atoi(data["running_skor"])
	maxScore, _ := strconv.Atoi(data["max_score"])

	return &sessionData{
		RunningSkor: runningSkor,
		MaxScore:    maxScore,
		Answers:     answers,
		State:       data["state"],
	}, nil
}

func (u *progressSessionUseCaseImpl) persistAnswers(ctx context.Context, key string, s *sessionData) error {
	answersJSON, err := json.Marshal(s.Answers)
	if err != nil {
		return err
	}
	return u.Redis.HMSet(ctx, key, map[string]any{
		"running_skor": s.RunningSkor,
		"answers":      string(answersJSON),
	}).Err()
}

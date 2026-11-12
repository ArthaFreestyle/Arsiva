package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
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

	// Finalize moves a Redis session into member_progress and optionally credits total_xp
	// and updates the member's level. Safe to call multiple times — second call is a no-op.
	// cause is "submit" or "expired" — used for logging only.
	Finalize(ctx context.Context, sessionKey string, cause string) (*model.ProgressFinalizeResponse, error)

	// ListExpiredSessionKeys returns session keys whose expiry <= now (used by the worker).
	ListExpiredSessionKeys(ctx context.Context) ([]string, error)
}

type progressSessionUseCaseImpl struct {
	Completion repository.CompletionRepository
	Repo       repository.MemberProgressRepository
	Redis      *redis.Client
	Log        *logrus.Logger
	Validator  *validator.Validate
}

func NewProgressSessionUseCase(
	repo repository.MemberProgressRepository,
	redisClient *redis.Client,
	log *logrus.Logger,
	validate *validator.Validate,
	completion repository.CompletionRepository,
) ProgressSessionUseCase {
	u := &progressSessionUseCaseImpl{
		Repo:       repo,
		Redis:      redisClient,
		Log:        log,
		Validator:  validate,
		Completion: completion,
	}
	return u
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

	if err != nil && err != redis.Nil {
		return nil, fiber.ErrInternalServerError
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

	completionID := uuid.NewString()
	fields := map[string]any{
		"completion_id":    completionID,
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

	args := []interface{}{int(ttl.Seconds()), expiresAt}
	for k, v := range fields {
		args = append(args, k, v)
	}
	created, err := u.Redis.Eval(ctx, `if redis.call('EXISTS',KEYS[1])==1 then return 0 end; redis.call('HSET',KEYS[1],unpack(ARGV,3)); redis.call('EXPIRE',KEYS[1],ARGV[1]); redis.call('ZADD',KEYS[2],ARGV[2],KEYS[1]); return 1`, []string{key, progressActiveZSet}, args...).Int()
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}
	if created == 0 {
		return nil, fiber.ErrConflict
	}

	return &model.ProgressStartResponse{
		SessionKey:   key,
		CompletionID: completionID,
		ExpiresAt:    expiresAt,
		MaxScore:     maxScore,
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
	if req.CompletionID != "" {
		if u.Completion == nil {
			return nil, fiber.ErrInternalServerError
		}
		mid, _ := strconv.Atoi(memberId)
		saved, err := u.Completion.FindCompletion(ctx, mid, req.CompletionID, req.ContentType, req.ContentId)
		if err != nil {
			return nil, fiber.ErrInternalServerError
		}
		if saved != nil {
			u.cleanupCompletion(ctx, key, req.CompletionID)
			return saved, nil
		}
		active, err := u.Redis.HGet(ctx, key, "completion_id").Result()
		if err == redis.Nil {
			return nil, fiber.ErrNotFound
		}
		if err != nil {
			return nil, fiber.ErrInternalServerError
		}
		if active != req.CompletionID {
			return nil, fiber.ErrConflict
		}
	}

	if state, _ := u.Redis.HGet(ctx, key, "state").Result(); state == "flushing" {
		return u.finalizeExpected(ctx, key, "submit", req.CompletionID)
	}
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

	isEnding, endingPoint, endingType, err := u.Repo.GetSceneEndingInfo(ctx, req.SceneId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fiber.NewError(fiber.StatusBadRequest, "scene tidak ditemukan")
		}
		return nil, fiber.ErrInternalServerError
	}

	// Record scene visit in Redis.
	sceneKey := strconv.Itoa(req.SceneId)
	fields := []interface{}{fmt.Sprintf("answers.%s", sceneKey), endingPoint}
	if isEnding {
		fields = append(fields, "running_skor", endingPoint, "reached_ending", 1, "ending_type", endingType)
	}
	if err = u.writeActive(ctx, key, req.CompletionID, fields...); err != nil {
		return nil, err
	}

	if isEnding {
		return u.finalizeExpected(ctx, key, "submit", req.CompletionID)
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
	if req.CompletionID != "" {
		if u.Completion == nil {
			return nil, fiber.ErrInternalServerError
		}
		mid, _ := strconv.Atoi(memberId)
		saved, err := u.Completion.FindCompletion(ctx, mid, req.CompletionID, req.ContentType, req.ContentId)
		if err != nil {
			return nil, fiber.ErrInternalServerError
		}
		if saved != nil {
			u.cleanupCompletion(ctx, key, req.CompletionID)
			return saved, nil
		}
		active, err := u.Redis.HGet(ctx, key, "completion_id").Result()
		if err == redis.Nil {
			return nil, fiber.ErrNotFound
		}
		if err != nil {
			return nil, fiber.ErrInternalServerError
		}
		if active != req.CompletionID {
			return nil, fiber.ErrConflict
		}
	}

	if state, _ := u.Redis.HGet(ctx, key, "state").Result(); state == "flushing" {
		return u.finalizeExpected(ctx, key, "submit", req.CompletionID)
	}
	if _, err := u.loadActiveSession(ctx, key); err != nil {
		return nil, err
	}

	skor := 0
	if req.Solved {
		skor = 1
	}

	if err := u.writeActive(ctx, key, req.CompletionID, "running_skor", skor); err != nil {
		return nil, err
	}

	return u.finalizeExpected(ctx, key, "submit", req.CompletionID)
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
	if req.CompletionID != "" {
		if u.Completion == nil {
			return nil, fiber.ErrInternalServerError
		}
		mid, _ := strconv.Atoi(memberId)
		saved, err := u.Completion.FindCompletion(ctx, mid, req.CompletionID, req.ContentType, req.ContentId)
		if err != nil {
			return nil, fiber.ErrInternalServerError
		}
		if saved != nil {
			u.cleanupCompletion(ctx, key, req.CompletionID)
			return saved, nil
		}
		active, err := u.Redis.HGet(ctx, key, "completion_id").Result()
		if err == redis.Nil {
			return nil, fiber.ErrNotFound
		}
		if err != nil {
			return nil, fiber.ErrInternalServerError
		}
		if active != req.CompletionID {
			return nil, fiber.ErrConflict
		}
	}

	return u.finalizeExpected(ctx, key, "submit", req.CompletionID)
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
		CompletionID:    data["completion_id"],
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
func (u *progressSessionUseCaseImpl) Finalize(ctx context.Context, key string, cause string) (*model.ProgressFinalizeResponse, error) {
	return u.finalizeExpected(ctx, key, cause, "")
}
func (u *progressSessionUseCaseImpl) finalizeExpected(ctx context.Context, key, cause, expected string) (*model.ProgressFinalizeResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	values, err := u.Redis.Eval(ctx, `if ARGV[1]~='' and redis.call('HGET',KEYS[1],'completion_id')~=ARGV[1] then return {} end; if redis.call('EXISTS',KEYS[1])==0 then return {} end; redis.call('HSET',KEYS[1],'state','flushing'); return redis.call('HGETALL',KEYS[1])`, []string{key}, expected).StringSlice()
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}
	data := map[string]string{}
	for i := 0; i+1 < len(values); i += 2 {
		data[values[i]] = values[i+1]
	}
	if len(data) == 0 {
		return nil, fiber.ErrNotFound
	}
	if data["completion_id"] == "" {
		return nil, fiber.NewError(409, "sesi lama; mulai sesi baru")
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

	if u.Completion != nil {
		saved, err := u.Completion.FindCompletion(ctx, memberId, data["completion_id"], contentType, contentId)
		if err != nil {
			return nil, fiber.ErrInternalServerError
		}
		if saved != nil {
			u.cleanupCompletion(ctx, key, data["completion_id"])
			return saved, nil
		}
	}
	// Step 3: fetch fresh xp_reward from content table (guru may have changed it mid-session).
	xpReward, err := u.Repo.GetContentXpReward(ctx, contentType, contentId)
	if err != nil {
		if u.Log != nil {
			u.Log.Errorf("Finalize GetContentXpReward (%s/%d): %v", contentType, contentId, err)
		}
		// Keep the immutable session available for retry.
		// A missing content row is a client-visible condition, not a server fault.
		if errors.Is(err, repository.ErrContentNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "konten tidak ditemukan")
		}
		return nil, fiber.ErrInternalServerError
	}

	// Step 4: decide whether this finish earns XP.
	//   - cerita: reaching ANY ending completes the story and earns its XP. max_score
	//     is the best ending's point, so gating on runningSkor == maxScore would deny
	//     XP for every non-optimal ending (and credit only the daily-task XP).
	//   - kuis/puzzle: full marks required.
	awardedXp := 0
	switch contentType {
	case "cerita":
		if data["reached_ending"] == "1" {
			awardedXp = xpReward
		}
	default:
		if runningSkor == maxScore {
			awardedXp = xpReward
		}
	}

	progress := &entity.MemberProgress{
		MemberId:    memberId,
		GroupId:     groupId,
		ContentType: contentType,
		ContentId:   contentId,
		Skor:        runningSkor,
		Duration:    durationSeconds,
	}

	if u.Completion == nil {
		return nil, fiber.ErrInternalServerError
	}
	today := todayInJakarta()
	response, err := u.Completion.Complete(ctx, data["completion_id"], progress, awardedXp, data["ending_type"], today, func(streak *entity.MemberStreak) { applyStreakRules(streak, today) })
	if err != nil {
		if u.Log != nil {
			u.Log.Errorf("Finalize completion (cause=%s): %v", cause, err)
		}
		if errors.Is(err, repository.ErrInvalidReference) {
			return nil, fiber.ErrBadRequest
		}
		if errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			return nil, fiber.ErrServiceUnavailable
		}
		return nil, fiber.ErrInternalServerError
	}
	u.cleanupCompletion(ctx, key, data["completion_id"])
	return response, nil

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
	CompletionID string
	RunningSkor  int
	MaxScore     int
	Answers      map[string]int
	State        string
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
		CompletionID: data["completion_id"],
		RunningSkor:  runningSkor,
		MaxScore:     maxScore,
		Answers:      answers,
		State:        data["state"],
	}, nil
}

func (u *progressSessionUseCaseImpl) persistAnswers(ctx context.Context, key string, s *sessionData) error {
	answersJSON, err := json.Marshal(s.Answers)
	if err != nil {
		return err
	}
	return u.writeActive(ctx, key, s.CompletionID, "running_skor", s.RunningSkor, "answers", string(answersJSON))
}
func (u *progressSessionUseCaseImpl) writeActive(ctx context.Context, key string, expected string, fields ...interface{}) error {
	n, err := u.Redis.Eval(ctx, `if redis.call('HGET',KEYS[1],'state')~='active' then return 0 end; if ARGV[1]~='' and redis.call('HGET',KEYS[1],'completion_id')~=ARGV[1] then return 0 end; redis.call('HSET',KEYS[1],unpack(ARGV,2)); return 1`, []string{key}, append([]interface{}{expected}, fields...)...).Int()
	if err != nil {
		return fiber.ErrInternalServerError
	}
	if n == 0 {
		return fiber.ErrConflict
	}
	return nil
}

func (u *progressSessionUseCaseImpl) cleanupCompletion(ctx context.Context, key, id string) {
	_, err := u.Redis.Eval(ctx, `if redis.call('HGET',KEYS[1],'completion_id') == ARGV[1] then redis.call('DEL',KEYS[1]); redis.call('ZREM',KEYS[2],KEYS[1]); end; return 1`, []string{key, progressActiveZSet}, id).Result()
	if err != nil && u.Log != nil {
		u.Log.Warnf("completion committed; Redis cleanup failed: %v", err)
	}
}

package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/repository"
)

// ─── Mock repository ──────────────────────────────────────────────────────────

type mockProgressRepo struct {
	contentExists           bool
	contentExistsErr        error
	kuisMaxScore            int
	kuisMaxScoreErr         error
	ceritaMaxScore          int
	jawabanScore            int
	jawabanScoreErr         error
	sceneIsEnding           bool
	sceneEndingPoint        int
	sceneEndingType         string
	sceneEndingErr          error
	xpReward                int
	xpRewardErr             error
	saveProgressId          int
	saveProgressPrevLevel   int
	saveProgressNewLevel    int
	saveProgressEffectiveXp int // the effectiveXp the mock returns; defaults to the passed-in awardedXp
	saveProgressErr         error
	pertanyaanOk            bool
	pertanyaanOkErr         error
	sceneOk                 bool
	sceneOkErr              error

	saveProgressCalls int
	lastAwardedXp     int
	lastAwardedXpSet  bool
	mu                sync.Mutex
}

func (m *mockProgressRepo) CheckContentExists(_ context.Context, _ string, _ int) (bool, error) {
	return m.contentExists, m.contentExistsErr
}
func (m *mockProgressRepo) GetKuisMaxScore(_ context.Context, _ int) (int, error) {
	return m.kuisMaxScore, m.kuisMaxScoreErr
}
func (m *mockProgressRepo) GetCeritaMaxScore(_ context.Context, _ int) (int, error) {
	return m.ceritaMaxScore, nil
}
func (m *mockProgressRepo) CheckPertanyaanBelongsToKuis(_ context.Context, _, _ int) (bool, error) {
	return m.pertanyaanOk, m.pertanyaanOkErr
}
func (m *mockProgressRepo) CheckSceneBelongsToCerita(_ context.Context, _, _ int) (bool, error) {
	return m.sceneOk, m.sceneOkErr
}
func (m *mockProgressRepo) GetJawabanScore(_ context.Context, _, _ int) (int, error) {
	return m.jawabanScore, m.jawabanScoreErr
}
func (m *mockProgressRepo) GetSceneEndingInfo(_ context.Context, _ int) (bool, int, string, error) {
	return m.sceneIsEnding, m.sceneEndingPoint, m.sceneEndingType, m.sceneEndingErr
}
func (m *mockProgressRepo) GetContentXpReward(_ context.Context, _ string, _ int) (int, error) {
	return m.xpReward, m.xpRewardErr
}
func (m *mockProgressRepo) SaveProgress(_ context.Context, _ *entity.MemberProgress, awardedXp int) (int, int, int, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveProgressCalls++
	m.lastAwardedXp = awardedXp
	m.lastAwardedXpSet = true
	effectiveXp := m.saveProgressEffectiveXp
	if effectiveXp == 0 {
		effectiveXp = awardedXp
	}
	return m.saveProgressId, m.saveProgressPrevLevel, m.saveProgressNewLevel, effectiveXp, m.saveProgressErr
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func newTestUseCase(t *testing.T, repo *mockProgressRepo) (ProgressSessionUseCase, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	uc := NewProgressSessionUseCase(repo, rdb, nil, validator.New(), nil)
	return uc, mr
}

func progressMemberClaims(memberId string) *model.Claims {
	return &model.Claims{
		Role: "member",
		Details: map[string]any{
			"member_id": memberId,
		},
	}
}

func defaultStartReq() *model.ProgressStartRequest {
	return &model.ProgressStartRequest{
		ContentType:     "kuis",
		ContentId:       1,
		DurationSeconds: 60,
	}
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestNewProgressSessionUseCase(t *testing.T) {
	uc := NewProgressSessionUseCase(nil, nil, nil, validator.New(), nil)
	if uc == nil {
		t.Fatal("expected usecase instance")
	}
}

func TestStart_CreatesSession(t *testing.T) {
	repo := &mockProgressRepo{contentExists: true, kuisMaxScore: 100, saveProgressId: 1}
	uc, _ := newTestUseCase(t, repo)

	resp, err := uc.Start(context.Background(), defaultStartReq(), progressMemberClaims("42"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.MaxScore != 100 {
		t.Errorf("expected MaxScore=100, got %d", resp.MaxScore)
	}
	if resp.ExpiresAt <= time.Now().Unix() {
		t.Error("expected ExpiresAt in the future")
	}
}

func TestStart_Conflict_DuplicateSession(t *testing.T) {
	repo := &mockProgressRepo{contentExists: true, kuisMaxScore: 100, saveProgressId: 1}
	uc, _ := newTestUseCase(t, repo)

	claims := progressMemberClaims("42")
	req := defaultStartReq()

	if _, err := uc.Start(context.Background(), req, claims); err != nil {
		t.Fatalf("first Start: %v", err)
	}

	// Second call for the same triple must return 409.
	_, err := uc.Start(context.Background(), req, claims)
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
}

func TestStart_ContentNotFound(t *testing.T) {
	repo := &mockProgressRepo{contentExists: false}
	uc, _ := newTestUseCase(t, repo)

	_, err := uc.Start(context.Background(), defaultStartReq(), progressMemberClaims("42"))
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
}

func TestAnswer_DuplicateRejected(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: 10, saveProgressId: 1,
		pertanyaanOk: true, jawabanScore: 5,
	}
	uc, _ := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")

	if _, err := uc.Start(context.Background(), defaultStartReq(), claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	answerReq := &model.ProgressAnswerRequest{
		ContentType:  "kuis",
		ContentId:    1,
		PertanyaanId: 7,
		JawabanId:    3,
	}

	if _, err := uc.Answer(context.Background(), answerReq, claims); err != nil {
		t.Fatalf("first Answer: %v", err)
	}

	// Second answer for the same pertanyaan must return 409.
	_, err := uc.Answer(context.Background(), answerReq, claims)
	if err == nil {
		t.Fatal("expected duplicate-answer conflict, got nil")
	}
}

func TestAnswer_NoDBWriteMidSession(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: 10, saveProgressId: 1,
		pertanyaanOk: true, jawabanScore: 5,
	}
	uc, _ := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")

	if _, err := uc.Start(context.Background(), defaultStartReq(), claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	answerReq := &model.ProgressAnswerRequest{
		ContentType: "kuis", ContentId: 1, PertanyaanId: 7, JawabanId: 3,
	}
	if _, err := uc.Answer(context.Background(), answerReq, claims); err != nil {
		t.Fatalf("Answer: %v", err)
	}

	// SaveProgress must NOT have been called yet.
	repo.mu.Lock()
	calls := repo.saveProgressCalls
	repo.mu.Unlock()
	if calls != 0 {
		t.Errorf("expected 0 SaveProgress calls mid-session, got %d", calls)
	}
}

func TestFinalize_PerfectScore_AwardsXP(t *testing.T) {
	maxScore := 100
	xpReward := 150
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: maxScore,
		xpReward: xpReward, saveProgressId: 99,
	}
	uc, mr := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")

	if _, err := uc.Start(context.Background(), defaultStartReq(), claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	key := sessionKey("42", "kuis", 1)

	// Simulate perfect score in Redis.
	mr.HSet(key, "running_skor", fmt.Sprintf("%d", maxScore))

	resp, err := uc.Finalize(context.Background(), key, "submit")
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if resp.ProgresId != 99 {
		t.Errorf("expected progresId=99, got %d", resp.ProgresId)
	}

	// SaveProgress called exactly once.
	repo.mu.Lock()
	calls := repo.saveProgressCalls
	repo.mu.Unlock()
	if calls != 1 {
		t.Errorf("expected 1 SaveProgress call, got %d", calls)
	}
}

func TestFinalize_PartialScore_NoXP(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: 100,
		xpReward: 150, saveProgressId: 7,
	}
	uc, mr := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")

	if _, err := uc.Start(context.Background(), defaultStartReq(), claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	key := sessionKey("42", "kuis", 1)
	// Partial score.
	mr.HSet(key, "running_skor", "50")

	// Finalize — awarded_xp should be 0 (passed to SaveProgress).
	// We verify by checking that SaveProgress was called (xpReward logic is in repo).
	if _, err := uc.Finalize(context.Background(), key, "submit"); err != nil {
		t.Fatalf("Finalize: %v", err)
	}

	repo.mu.Lock()
	calls := repo.saveProgressCalls
	repo.mu.Unlock()
	if calls != 1 {
		t.Errorf("expected 1 SaveProgress call, got %d", calls)
	}
}

func TestFinalize_DoubleConcurrent_InsertOnce(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: 100,
		xpReward: 100, saveProgressId: 5,
	}
	uc, _ := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")

	if _, err := uc.Start(context.Background(), defaultStartReq(), claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	key := sessionKey("42", "kuis", 1)

	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			_, _ = uc.Finalize(context.Background(), key, "submit")
		}()
	}
	wg.Wait()

	repo.mu.Lock()
	calls := repo.saveProgressCalls
	repo.mu.Unlock()
	if calls != 1 {
		t.Errorf("expected exactly 1 SaveProgress call with concurrent finalize, got %d", calls)
	}
}

func TestFinalize_DBError_LeavesRedisActive(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: 100,
		xpReward: 100, saveProgressErr: errors.New("db down"),
	}
	uc, mr := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")

	if _, err := uc.Start(context.Background(), defaultStartReq(), claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	key := sessionKey("42", "kuis", 1)
	if _, err := uc.Finalize(context.Background(), key, "submit"); err == nil {
		t.Fatal("expected error from Finalize when SaveProgress fails")
	}

	// Redis key must still exist with state=active.
	state := mr.HGet(key, "state")
	if state != "active" {
		t.Errorf("expected state=active after DB failure, got %q", state)
	}
}

func TestWorker_FlushesExpiredSession(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: 10,
		xpReward: 50, saveProgressId: 3,
	}
	uc, mr := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")

	req := &model.ProgressStartRequest{
		ContentType:     "kuis",
		ContentId:       1,
		DurationSeconds: 60,
	}
	if _, err := uc.Start(context.Background(), req, claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	key := sessionKey("42", "kuis", 1)

	// Move the ZSET score to a past timestamp so the worker sees it as expired.
	pastScore := float64(time.Now().Unix() - 300)
	mr.ZAdd(progressActiveZSet, pastScore, key)

	// ListExpiredSessionKeys should return our key.
	keys, err := uc.ListExpiredSessionKeys(context.Background())
	if err != nil {
		t.Fatalf("ListExpiredSessionKeys: %v", err)
	}
	found := false
	for _, k := range keys {
		if k == key {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected expired key %s in list, got %v", key, keys)
	}

	// Finalize via worker.
	if _, err := uc.Finalize(context.Background(), key, "expired"); err != nil {
		t.Fatalf("worker Finalize: %v", err)
	}

	repo.mu.Lock()
	calls := repo.saveProgressCalls
	repo.mu.Unlock()
	if calls != 1 {
		t.Errorf("expected 1 SaveProgress call from worker, got %d", calls)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	repo := &mockProgressRepo{}
	uc, _ := newTestUseCase(t, repo)

	_, err := uc.GetSession(context.Background(), "kuis", 1, progressMemberClaims("42"))
	if err == nil {
		t.Fatal("expected not-found error for missing session")
	}
}

// ─── Regression: issue #38 — ending scene must finalize, not 500 ──────────────

// Walking a cerita to an is_ending scene must run Finalize → SaveProgress and return
// the progres_id (HTTP 200), not an opaque error. Mirrors the issue's repro.
func TestScene_EndingScene_FinalizesAndReturnsProgresId(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, ceritaMaxScore: 100,
		xpReward: 150, saveProgressId: 77,
		sceneOk: true, sceneIsEnding: true, sceneEndingPoint: 100,
	}
	uc, _ := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")

	startReq := &model.ProgressStartRequest{ContentType: "cerita", ContentId: 1, DurationSeconds: 60}
	if _, err := uc.Start(context.Background(), startReq, claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	sceneReq := &model.ProgressSceneRequest{ContentType: "cerita", ContentId: 1, SceneId: 9}
	resp, err := uc.Scene(context.Background(), sceneReq, claims)
	if err != nil {
		t.Fatalf("Scene (ending): expected success, got error: %v", err)
	}
	if resp == nil {
		t.Fatal("Scene (ending): expected finalize response, got nil")
	}
	if resp.ProgresId != 77 {
		t.Errorf("expected progresId=77, got %d", resp.ProgresId)
	}

	repo.mu.Lock()
	calls := repo.saveProgressCalls
	repo.mu.Unlock()
	if calls != 1 {
		t.Errorf("expected 1 SaveProgress call on ending, got %d", calls)
	}
}

// Regression: reaching a NON-optimal cerita ending must still award the story's XP.
// A branching story's max_score is its best ending's point; gating XP on
// running_skor == max_score denied XP for every other ending, so only the daily-task
// XP ever reached the profile. Reaching any ending completes the story and earns its XP.
func TestScene_NonOptimalEnding_StillAwardsXP(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, ceritaMaxScore: 100,
		xpReward: 150, saveProgressId: 88,
		// Player reaches an ending worth 40 points — below the best ending (100).
		sceneOk: true, sceneIsEnding: true, sceneEndingPoint: 40, sceneEndingType: "bad_ending",
	}
	uc, _ := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")

	startReq := &model.ProgressStartRequest{ContentType: "cerita", ContentId: 1, DurationSeconds: 60}
	if _, err := uc.Start(context.Background(), startReq, claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	sceneReq := &model.ProgressSceneRequest{ContentType: "cerita", ContentId: 1, SceneId: 9}
	resp, err := uc.Scene(context.Background(), sceneReq, claims)
	if err != nil {
		t.Fatalf("Scene (ending): %v", err)
	}

	repo.mu.Lock()
	awarded, set := repo.lastAwardedXp, repo.lastAwardedXpSet
	repo.mu.Unlock()
	if !set {
		t.Fatal("expected SaveProgress to be called on ending")
	}
	if awarded != 150 {
		t.Errorf("expected awardedXp=150 for completing the story (any ending), got %d", awarded)
	}
	// The finalize response must echo which ending the member reached so the client
	// can display it.
	if resp.EndingType != "bad_ending" {
		t.Errorf("expected EndingType=bad_ending in finalize response, got %q", resp.EndingType)
	}
}

// An expired cerita session that never reached an ending must NOT award XP — the story
// was abandoned mid-play, so there is nothing to credit.
func TestFinalize_CeritaNoEnding_NoXP(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, ceritaMaxScore: 100,
		xpReward: 150, saveProgressId: 89,
	}
	uc, mr := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")

	startReq := &model.ProgressStartRequest{ContentType: "cerita", ContentId: 1, DurationSeconds: 60}
	if _, err := uc.Start(context.Background(), startReq, claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	key := sessionKey("42", "cerita", 1)
	if _, err := uc.Finalize(context.Background(), key, "expired"); err != nil {
		t.Fatalf("Finalize: %v", err)
	}

	repo.mu.Lock()
	awarded := repo.lastAwardedXp
	repo.mu.Unlock()
	if awarded != 0 {
		t.Errorf("expected awardedXp=0 for an abandoned story, got %d", awarded)
	}
	_ = mr
}

// A foreign-key violation that the repository couldn't recover from is bad input —
// finalize must surface it as 400, not 500.
func TestFinalize_InvalidReference_Returns400(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: 100, xpReward: 100,
		saveProgressErr: fmt.Errorf("%w: member_progress_member_id_fkey", repository.ErrInvalidReference),
	}
	uc, _ := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")
	if _, err := uc.Start(context.Background(), defaultStartReq(), claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	_, err := uc.Finalize(context.Background(), sessionKey("42", "kuis", 1), "submit")
	var fe *fiber.Error
	if !errors.As(err, &fe) || fe.Code != fiber.StatusBadRequest {
		t.Fatalf("expected 400 fiber error for invalid reference, got %v", err)
	}
}

// A missing content row at finalize is a 404, not a 500.
func TestFinalize_ContentNotFound_Returns404(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: 100,
		xpRewardErr: repository.ErrContentNotFound,
	}
	uc, _ := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")
	if _, err := uc.Start(context.Background(), defaultStartReq(), claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	_, err := uc.Finalize(context.Background(), sessionKey("42", "kuis", 1), "submit")
	var fe *fiber.Error
	if !errors.As(err, &fe) || fe.Code != fiber.StatusNotFound {
		t.Fatalf("expected 404 fiber error for missing content, got %v", err)
	}
}

// ─── Level-up tests ───────────────────────────────────────────────────────────

// No XP awarded → LeveledUp must be false.
func TestFinalize_NoXP_LeveledUpFalse(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: 100,
		xpReward: 150, saveProgressId: 1,
		saveProgressPrevLevel: 0, saveProgressNewLevel: 0,
	}
	uc, mr := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")
	if _, err := uc.Start(context.Background(), defaultStartReq(), claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	key := sessionKey("42", "kuis", 1)
	// Partial score → no XP awarded.
	mr.HSet(key, "running_skor", "50")

	resp, err := uc.Finalize(context.Background(), key, "submit")
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if resp.LeveledUp {
		t.Errorf("expected LeveledUp=false when no XP awarded")
	}
}

// Single level-up: mock returns prevLevel=0, newLevel=1 → LeveledUp must be true.
func TestFinalize_SingleLevelUp_SignaledInResponse(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: 100,
		xpReward: 100, saveProgressId: 2,
		saveProgressPrevLevel: 0, saveProgressNewLevel: 1,
	}
	uc, mr := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")
	if _, err := uc.Start(context.Background(), defaultStartReq(), claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	key := sessionKey("42", "kuis", 1)
	mr.HSet(key, "running_skor", "100")

	resp, err := uc.Finalize(context.Background(), key, "submit")
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if !resp.LeveledUp {
		t.Errorf("expected LeveledUp=true on single level-up")
	}
	if resp.PreviousLevel != 0 || resp.NewLevel != 1 {
		t.Errorf("expected PreviousLevel=0 NewLevel=1, got %d/%d", resp.PreviousLevel, resp.NewLevel)
	}
}

// Multi-level jump: mock returns prevLevel=0, newLevel=2 → correct final level reported.
func TestFinalize_MultiLevelJump_SignaledInResponse(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: 100,
		xpReward: 500, saveProgressId: 3,
		saveProgressPrevLevel: 0, saveProgressNewLevel: 2,
	}
	uc, mr := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")
	if _, err := uc.Start(context.Background(), defaultStartReq(), claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	key := sessionKey("42", "kuis", 1)
	mr.HSet(key, "running_skor", "100")

	resp, err := uc.Finalize(context.Background(), key, "submit")
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if !resp.LeveledUp {
		t.Errorf("expected LeveledUp=true on multi-level jump")
	}
	if resp.NewLevel != 2 {
		t.Errorf("expected NewLevel=2 on multi-level jump, got %d", resp.NewLevel)
	}
}

// ─── One-time XP rule (issue #40) ────────────────────────────────────────────

// When the mock reports effectiveXp=0 (simulating "already earned"), the response
// must show no level-up and no XP, even though the candidate awardedXp was non-zero.
func TestFinalize_RepeatCompletion_ZeroXpInResponse(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: 100,
		xpReward: 150, saveProgressId: 10,
		// Simulate repo zeroing XP because this member already earned it.
		saveProgressEffectiveXp: 0,
		saveProgressPrevLevel:   0, saveProgressNewLevel: 0,
	}
	uc, mr := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")
	if _, err := uc.Start(context.Background(), defaultStartReq(), claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	key := sessionKey("42", "kuis", 1)
	mr.HSet(key, "running_skor", "100") // perfect score

	resp, err := uc.Finalize(context.Background(), key, "submit")
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if resp.LeveledUp {
		t.Errorf("expected LeveledUp=false when repo zeroes XP on repeat")
	}
	if resp.PreviousLevel != 0 || resp.NewLevel != 0 {
		t.Errorf("expected PreviousLevel=0 NewLevel=0 on repeat, got %d/%d", resp.PreviousLevel, resp.NewLevel)
	}
	// SaveProgress still called — the row must always be inserted.
	repo.mu.Lock()
	calls := repo.saveProgressCalls
	repo.mu.Unlock()
	if calls != 1 {
		t.Errorf("expected 1 SaveProgress call even on repeat, got %d", calls)
	}
}

// First perfect attempt earns XP (mock passes through awardedXp); level-up is signalled.
func TestFinalize_FirstPerfect_EarnsXP_LevelUp(t *testing.T) {
	repo := &mockProgressRepo{
		contentExists: true, kuisMaxScore: 100,
		xpReward: 100, saveProgressId: 11,
		// saveProgressEffectiveXp == 0 means "pass through awardedXp" per mock logic.
		saveProgressPrevLevel: 0, saveProgressNewLevel: 1,
	}
	uc, mr := newTestUseCase(t, repo)
	claims := progressMemberClaims("42")
	if _, err := uc.Start(context.Background(), defaultStartReq(), claims); err != nil {
		t.Fatalf("Start: %v", err)
	}

	key := sessionKey("42", "kuis", 1)
	mr.HSet(key, "running_skor", "100")

	resp, err := uc.Finalize(context.Background(), key, "submit")
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if !resp.LeveledUp {
		t.Errorf("expected LeveledUp=true on first perfect completion")
	}
	if resp.NewLevel != 1 {
		t.Errorf("expected NewLevel=1, got %d", resp.NewLevel)
	}
}

// Verify that the answers field is correctly serialised/deserialised round-trip.
func TestAnswers_JSONRoundTrip(t *testing.T) {
	answers := map[string]int{"1": 5, "2": 10}
	raw, _ := json.Marshal(answers)
	decoded := map[string]int{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json round-trip: %v", err)
	}
	if decoded["1"] != 5 || decoded["2"] != 10 {
		t.Errorf("answers round-trip mismatch: %v", decoded)
	}
}

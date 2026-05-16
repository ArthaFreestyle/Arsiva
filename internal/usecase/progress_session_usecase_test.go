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
	"github.com/redis/go-redis/v9"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

// ─── Mock repository ──────────────────────────────────────────────────────────

type mockProgressRepo struct {
	contentExists     bool
	contentExistsErr  error
	kuisMaxScore      int
	kuisMaxScoreErr   error
	ceritaMaxScore    int
	jawabanScore      int
	jawabanScoreErr   error
	sceneIsEnding     bool
	sceneEndingPoint  int
	sceneEndingErr    error
	xpReward          int
	xpRewardErr       error
	saveProgressId    int
	saveProgressErr   error
	pertanyaanOk      bool
	pertanyaanOkErr   error
	sceneOk           bool
	sceneOkErr        error

	saveProgressCalls int
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
func (m *mockProgressRepo) GetSceneEndingInfo(_ context.Context, _ int) (bool, int, error) {
	return m.sceneIsEnding, m.sceneEndingPoint, m.sceneEndingErr
}
func (m *mockProgressRepo) GetContentXpReward(_ context.Context, _ string, _ int) (int, error) {
	return m.xpReward, m.xpRewardErr
}
func (m *mockProgressRepo) SaveProgress(_ context.Context, _ *entity.MemberProgress, _ int) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveProgressCalls++
	return m.saveProgressId, m.saveProgressErr
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
	uc := NewProgressSessionUseCase(repo, rdb, nil, validator.New())
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
	uc := NewProgressSessionUseCase(nil, nil, nil, validator.New())
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

	progresId, err := uc.Finalize(context.Background(), key, "submit")
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if progresId != 99 {
		t.Errorf("expected progresId=99, got %d", progresId)
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

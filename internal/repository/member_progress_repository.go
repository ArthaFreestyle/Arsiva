package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/utils"
)

var ErrContentNotFound = errors.New("content not found or not published")

// ErrInvalidReference is returned when SaveProgress hits a foreign-key violation it
// cannot recover from (e.g. member_id does not reference a real member). The caller
// should treat this as bad input (4xx) rather than an opaque server error.
var ErrInvalidReference = errors.New("progress references a row that does not exist")

type MemberProgressRepository interface {
	// Content validation — checks exists + is_published in one query.
	CheckContentExists(ctx context.Context, contentType string, contentId int) (bool, error)

	// Max score per content type — fetched once at session start and cached in Redis.
	GetKuisMaxScore(ctx context.Context, kuisId int) (int, error)
	GetCeritaMaxScore(ctx context.Context, ceritaId int) (int, error)
	// Puzzle max score is always 1 (binary), no DB query needed.

	// Validates that a pertanyaan belongs to the given kuis.
	CheckPertanyaanBelongsToKuis(ctx context.Context, pertanyaanId, kuisId int) (bool, error)

	// Validates that a scene belongs to the given cerita.
	CheckSceneBelongsToCerita(ctx context.Context, sceneId, ceritaId int) (bool, error)

	// Returns the score of a selected jawaban choice.
	GetJawabanScore(ctx context.Context, jawabanId, pertanyaanId int) (int, error)

	// Returns the scene's is_ending and ending_point fields.
	GetSceneEndingInfo(ctx context.Context, sceneId int) (isEnding bool, endingPoint int, err error)

	// Returns the xp_reward from the content table (read fresh at finalize time).
	GetContentXpReward(ctx context.Context, contentType string, contentId int) (int, error)

	// Atomically inserts member_progress and optionally credits total_xp and recomputes level.
	// Returns the new progres_id, previous level, new level, and the XP actually awarded
	// (0 if the member already earned XP for this content, even when awardedXp > 0).
	SaveProgress(ctx context.Context, progress *entity.MemberProgress, awardedXp int) (progresId int, prevLevel int, newLevel int, effectiveXp int, err error)
}

type memberProgressRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewMemberProgressRepository(db *pgxpool.Pool, log *logrus.Logger) MemberProgressRepository {
	return &memberProgressRepositoryImpl{DB: db, Log: log}
}

func (r *memberProgressRepositoryImpl) CheckContentExists(ctx context.Context, contentType string, contentId int) (bool, error) {
	var query string
	switch contentType {
	case "kuis":
		query = `SELECT 1 FROM kuis WHERE kuis_id = $1 AND is_published = true`
	case "cerita":
		query = `SELECT 1 FROM cerita_interaktif WHERE cerita_id = $1 AND is_published = true`
	case "puzzle":
		query = `SELECT 1 FROM puzzles WHERE puzzle_id = $1 AND is_published = true`
	default:
		return false, ErrContentNotFound
	}

	var dummy int
	err := r.DB.QueryRow(ctx, query, contentId).Scan(&dummy)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		r.Log.Errorf("CheckContentExists %s/%d: %v", contentType, contentId, err)
		return false, err
	}
	return true, nil
}

func (r *memberProgressRepositoryImpl) GetKuisMaxScore(ctx context.Context, kuisId int) (int, error) {
	// Sum of the highest-scoring choice per question — no N+1, done in one query.
	query := `
		SELECT COALESCE(SUM(max_choice.score), 0)
		FROM pertanyaan_kuis q
		JOIN LATERAL (
			SELECT MAX(score) AS score
			FROM pilihan_kuis
			WHERE pertanyaan_id = q.pertanyaan_id
		) max_choice ON true
		WHERE q.kuis_id = $1
	`
	var maxScore int
	err := r.DB.QueryRow(ctx, query, kuisId).Scan(&maxScore)
	if err != nil {
		r.Log.Errorf("GetKuisMaxScore kuis_id=%d: %v", kuisId, err)
		return 0, err
	}
	return maxScore, nil
}

func (r *memberProgressRepositoryImpl) GetCeritaMaxScore(ctx context.Context, ceritaId int) (int, error) {
	query := `
		SELECT COALESCE(MAX(ending_point), 0)
		FROM scene
		WHERE cerita_id = $1 AND is_ending = true
	`
	var maxScore int
	err := r.DB.QueryRow(ctx, query, ceritaId).Scan(&maxScore)
	if err != nil {
		r.Log.Errorf("GetCeritaMaxScore cerita_id=%d: %v", ceritaId, err)
		return 0, err
	}
	return maxScore, nil
}

func (r *memberProgressRepositoryImpl) CheckPertanyaanBelongsToKuis(ctx context.Context, pertanyaanId, kuisId int) (bool, error) {
	var dummy int
	err := r.DB.QueryRow(ctx,
		`SELECT 1 FROM pertanyaan_kuis WHERE pertanyaan_id = $1 AND kuis_id = $2`,
		pertanyaanId, kuisId).Scan(&dummy)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		r.Log.Errorf("CheckPertanyaanBelongsToKuis: %v", err)
		return false, err
	}
	return true, nil
}

func (r *memberProgressRepositoryImpl) CheckSceneBelongsToCerita(ctx context.Context, sceneId, ceritaId int) (bool, error) {
	var dummy int
	err := r.DB.QueryRow(ctx,
		`SELECT 1 FROM scene WHERE scene_id = $1 AND cerita_id = $2`,
		sceneId, ceritaId).Scan(&dummy)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		r.Log.Errorf("CheckSceneBelongsToCerita: %v", err)
		return false, err
	}
	return true, nil
}

func (r *memberProgressRepositoryImpl) GetJawabanScore(ctx context.Context, jawabanId, pertanyaanId int) (int, error) {
	var score int
	err := r.DB.QueryRow(ctx,
		`SELECT score FROM pilihan_kuis WHERE jawaban_id = $1 AND pertanyaan_id = $2`,
		jawabanId, pertanyaanId).Scan(&score)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, pgx.ErrNoRows
	}
	if err != nil {
		r.Log.Errorf("GetJawabanScore jawaban_id=%d: %v", jawabanId, err)
		return 0, err
	}
	return score, nil
}

func (r *memberProgressRepositoryImpl) GetSceneEndingInfo(ctx context.Context, sceneId int) (bool, int, error) {
	var isEnding bool
	var endingPoint int
	err := r.DB.QueryRow(ctx,
		`SELECT is_ending, COALESCE(ending_point, 0) FROM scene WHERE scene_id = $1`,
		sceneId).Scan(&isEnding, &endingPoint)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, 0, pgx.ErrNoRows
	}
	if err != nil {
		r.Log.Errorf("GetSceneEndingInfo scene_id=%d: %v", sceneId, err)
		return false, 0, err
	}
	return isEnding, endingPoint, nil
}

func (r *memberProgressRepositoryImpl) GetContentXpReward(ctx context.Context, contentType string, contentId int) (int, error) {
	var query string
	switch contentType {
	case "kuis":
		query = `SELECT xp_reward FROM kuis WHERE kuis_id = $1`
	case "cerita":
		query = `SELECT xp_reward FROM cerita_interaktif WHERE cerita_id = $1`
	case "puzzle":
		query = `SELECT xp_reward FROM puzzles WHERE puzzle_id = $1`
	default:
		return 0, ErrContentNotFound
	}

	var xpReward int
	err := r.DB.QueryRow(ctx, query, contentId).Scan(&xpReward)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrContentNotFound
	}
	if err != nil {
		r.Log.Errorf("GetContentXpReward %s/%d: %v", contentType, contentId, err)
		return 0, err
	}
	return xpReward, nil
}

func (r *memberProgressRepositoryImpl) SaveProgress(ctx context.Context, progress *entity.MemberProgress, awardedXp int) (int, int, int, int, error) {
	var groupId *string
	if progress.GroupId != "" {
		groupId = &progress.GroupId
	}

	progresId, prevLevel, newLevel, effectiveXp, err := r.insertProgress(ctx, progress, awardedXp, groupId)
	if err == nil {
		return progresId, prevLevel, newLevel, effectiveXp, nil
	}

	// Classify the failure. A foreign-key violation means the row references something
	// that no longer exists — member_progress has FKs on member_id and group_id
	// (migration 20260318204005). The most common cause of an ending-scene 500 is a
	// stale/invalid group_id that was attached at /start: the session plays through but
	// the finalize INSERT trips member_progress_group_id_fkey. Losing the whole
	// completion over a dropped group reference is the wrong trade-off, so we retry once
	// with group_id = NULL (keeping the score + XP) and only surface a hard error when
	// the broken reference is something we can't shed.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
		if pgErr.ConstraintName == "member_progress_group_id_fkey" && groupId != nil {
			r.Log.Warnf("SaveProgress: group_id=%q does not reference a valid group; persisting progress without it", *groupId)
			if progresId, prevLevel, newLevel, effectiveXp, err = r.insertProgress(ctx, progress, awardedXp, nil); err == nil {
				return progresId, prevLevel, newLevel, effectiveXp, nil
			}
		}
		r.Log.Errorf("SaveProgress foreign-key violation (constraint=%s detail=%s): %v", pgErr.ConstraintName, pgErr.Detail, err)
		return 0, 0, 0, 0, fmt.Errorf("%w: %s", ErrInvalidReference, pgErr.ConstraintName)
	}

	return 0, 0, 0, 0, err
}

// insertProgress runs the member_progress INSERT (and optional XP + level update) in a
// single transaction. Returns progres_id, previous level, new level, and the XP actually
// written — which may be 0 even when awardedXp > 0 if the member already earned XP for
// this (member_id, content_type, content_id) triple.
func (r *memberProgressRepositoryImpl) insertProgress(ctx context.Context, progress *entity.MemberProgress, awardedXp int, groupId *string) (progresId int, prevLevel int, newLevel int, effectiveXp int, err error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		r.Log.Errorf("SaveProgress begin tx: %v", err)
		return 0, 0, 0, 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Serialize concurrent finalizes for the same member+content so the EXISTS check
	// and INSERT can't interleave. hashtext produces an int4; cast to int8 for the lock.
	lockKey := fmt.Sprintf("mp:%d:%s:%d", progress.MemberId, progress.ContentType, progress.ContentId)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1)::bigint)`, lockKey); err != nil {
		r.Log.Errorf("SaveProgress advisory lock: %v", err)
		return 0, 0, 0, 0, err
	}

	// XP is awarded at most once per (member_id, content_type, content_id).
	// A prior 0-XP row (imperfect attempt) does NOT block a later perfect run.
	effectiveXp = awardedXp
	if awardedXp > 0 {
		var alreadyEarned bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM member_progress
				WHERE member_id    = $1
				  AND content_type = $2::content_type_enum
				  AND content_id   = $3
				  AND xp_reward    > 0
			)
		`, progress.MemberId, progress.ContentType, progress.ContentId).Scan(&alreadyEarned)
		if err != nil {
			r.Log.Errorf("SaveProgress xp-earned check: %v", err)
			return 0, 0, 0, 0, err
		}
		if alreadyEarned {
			effectiveXp = 0
		}
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO member_progress
			(member_id, group_id, content_type, content_id, skor, xp_reward, completed_at, duration)
		VALUES ($1, $2, $3::content_type_enum, $4, $5, $6, NOW(), $7)
		RETURNING progres_id
	`, progress.MemberId, groupId, progress.ContentType, progress.ContentId,
		progress.Skor, effectiveXp, progress.Duration).Scan(&progresId)
	if err != nil {
		r.Log.Errorf("SaveProgress insert: %v", err)
		return 0, 0, 0, 0, err
	}

	if effectiveXp > 0 {
		var newTotalXp int
		err = tx.QueryRow(ctx,
			`UPDATE members SET total_xp = total_xp + $1 WHERE member_id = $2 RETURNING total_xp`,
			effectiveXp, progress.MemberId).Scan(&newTotalXp)
		if err != nil {
			r.Log.Errorf("SaveProgress update xp: %v", err)
			return 0, 0, 0, 0, err
		}

		prevLevel = utils.LevelForXP(newTotalXp - effectiveXp)
		newLevel = utils.LevelForXP(newTotalXp)

		if newLevel != prevLevel {
			_, err = tx.Exec(ctx,
				`UPDATE members SET level = $1 WHERE member_id = $2`,
				newLevel, progress.MemberId)
			if err != nil {
				r.Log.Errorf("SaveProgress update level: %v", err)
				return 0, 0, 0, 0, err
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		r.Log.Errorf("SaveProgress commit: %v", err)
		return 0, 0, 0, 0, err
	}
	return progresId, prevLevel, newLevel, effectiveXp, nil
}

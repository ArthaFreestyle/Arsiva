package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func completionDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a disposable PostgreSQL server with CREATEDB")
	}
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("arsiva_completion_test_%d", time.Now().UnixNano())
	if _, err = admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.Database = name
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		admin.Exec(ctx, "DROP DATABASE "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)")
		admin.Close(ctx)
	})
	// Use real production migrations, including the new up/down pair.
	files, err := filepath.Glob("../../db/migrations_postgre/*.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		// Historical migration redefines an existing FK; normalize only the test database before applying it.
		if filepath.Base(f) == "20260511000001_add_fk_member_social_links.up.sql" {
			if _, err = db.Exec(ctx, `ALTER TABLE member_social_links DROP CONSTRAINT member_social_links_member_id_fkey`); err != nil {
				t.Fatal(err)
			}
		}
		sql, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec(ctx, string(sql)); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
	}
	return db
}
func TestCompletionTransactionIntegration(t *testing.T) {
	db := completionDB(t)
	ctx := context.Background()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.Exec(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}
	// Minimal required member fields come from actual schema.
	exec(`INSERT INTO members(member_id,total_xp,level) VALUES(1,350,1),(2,0,0)`)
	exec(`INSERT INTO achievements(nama,badge_icon,xp_required,tier) VALUES('Zero','/0',0,'bronze'),('Exact','/1',400,'bronze'),('Above','/2',401,'bronze'),('Later','/3',900,'bronze')`)
	repo := NewCompletionRepository(db, logrus.New())
	today := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	p := &entity.MemberProgress{MemberId: 1, ContentType: "kuis", ContentId: 10, Skor: 100, Duration: 60, GroupId: "missing"}
	complete := func(id string, xp int) (*model.ProgressFinalizeResponse, error) {
		return repo.Complete(ctx, id, p, xp, "", today, func(s *entity.MemberStreak) { s.CurrentStreak = 1; s.LastActiveDate = &today })
	}
	first, err := complete("first", 0)
	if err != nil {
		t.Fatal(err)
	}
	if first.TotalXP != 370 || first.BonusXP != 20 || first.NewLevel != 1 || len(first.Notifications) != 1 {
		t.Fatalf("first=%+v", first)
	}
	second, err := complete("second", 30)
	if err != nil {
		t.Fatal(err)
	}
	if second.TotalXP != 400 || second.NewLevel != 2 || !second.LeveledUp || len(second.Notifications) != 1 {
		t.Fatalf("second=%+v", second)
	}
	replay, err := complete("second", 999)
	if err != nil || !reflect.DeepEqual(second, replay) {
		t.Fatalf("replay mismatch %v %#v %#v", err, second, replay)
	}
	nr := &NotificationRepository{DB: db}
	n := second.Notifications[0]
	read, err := nr.Read(ctx, 1, mustID(t, n.NotificationID))
	if err != nil {
		t.Fatal(err)
	}
	again, err := nr.Read(ctx, 1, mustID(t, n.NotificationID))
	if err != nil || !read.ReadAt.Equal(*again.ReadAt) {
		t.Fatal("mark-read not idempotent")
	}
	if _, err = nr.Read(ctx, 2, mustID(t, n.NotificationID)); err != pgx.ErrNoRows {
		t.Fatalf("ownership: %v", err)
	}
	inbox, total, err := nr.List(ctx, 1, 1, 1, false)
	if err != nil || total != 2 || len(inbox) != 1 || inbox[0].NotificationID != n.NotificationID {
		t.Fatalf("inbox=%v total=%d err=%v", inbox, total, err)
	}
	if count, err := nr.Count(ctx, 1); err != nil || count != 1 {
		t.Fatal(count, err)
	}
	// Failure after progress/XP/daily tasks must roll everything back.
	exec(`CREATE FUNCTION reject_notification() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected failure'; END $$; CREATE TRIGGER reject_notification BEFORE INSERT ON member_notifications FOR EACH ROW EXECUTE FUNCTION reject_notification()`)
	p.ContentId = 11
	if _, err = complete("rollback", 100); err == nil {
		t.Fatal("expected injected failure")
	}
	var xp, progress int
	db.QueryRow(ctx, `SELECT total_xp FROM members WHERE member_id=1`).Scan(&xp)
	db.QueryRow(ctx, `SELECT count(*) FROM member_progress WHERE content_id=11`).Scan(&progress)
	if xp != 400 || progress != 0 {
		t.Fatalf("partial commit: xp=%d progress=%d", xp, progress)
	}
	exec(`DROP TRIGGER reject_notification ON member_notifications`)
	// Same completion and distinct completions contend on the member lock.
	var wg sync.WaitGroup
	errs := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := "concurrent"
			if i%2 == 0 {
				id = fmt.Sprintf("distinct-%d", i)
			}
			_, err := complete(id, 100)
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	db.QueryRow(ctx, `SELECT total_xp FROM members WHERE member_id=1`).Scan(&xp)
	if xp != 530 {
		t.Fatalf("duplicated XP or bonus: %d", xp)
	}
	var tasks int
	db.QueryRow(ctx, `SELECT current_count FROM daily_tasks WHERE member_id=1 AND task_type='earn_xp'`).Scan(&tasks)
	if tasks != 100 {
		t.Fatal(tasks)
	}
	// Reconciliation is silent by default, rerunnable; deletion isn't permanent revocation.
	exec(`INSERT INTO achievements(nama,badge_icon,xp_required,tier) VALUES('Backfill','/b',500,'bronze')`)
	if n, err := ReconcileMember(ctx, db, 1, false); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	if n, err := ReconcileMember(ctx, db, 1, false); err != nil || n != 0 {
		t.Fatal(n, err)
	}
	exec(`DELETE FROM member_achievements WHERE member_id=1 AND achievement_id=1`)
	if n, err := ReconcileMember(ctx, db, 1, true); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	var notifications int
	db.QueryRow(ctx, `SELECT count(*) FROM member_notifications WHERE member_id=1`).Scan(&notifications)
	if notifications != 3 {
		t.Fatal(notifications)
	}
	repeat, err := complete("repeat", 100)
	if err != nil || repeat.AwardedXP != 0 || repeat.NewLevel != 2 || len(repeat.Notifications) != 0 || repeat.Notifications == nil || repeat.NewlyUnlockedAchievements == nil {
		t.Fatalf("repeat=%+v %v", repeat, err)
	}
	if saved, err := repo.FindCompletion(ctx, 2, "second", "kuis", 10); err != nil || saved != nil {
		t.Fatal("cross-member replay")
	}

	exec(`INSERT INTO members(member_id,total_xp,level) VALUES(3,390,1),(4,0,0)`)
	bonus, err := repo.Complete(ctx, "bonus-only", &entity.MemberProgress{MemberId: 3, ContentType: "kuis", ContentId: 1}, 0, "", today, func(*entity.MemberStreak) {})
	if err != nil || bonus.TotalXP != 410 || !bonus.LeveledUp || bonus.AwardedXP != 0 || len(bonus.NewlyUnlockedAchievements) != 3 {
		t.Fatalf("bonus only: %+v %v", bonus, err)
	}
	multi, err := repo.Complete(ctx, "multi", &entity.MemberProgress{MemberId: 4, ContentType: "puzzle", ContentId: 1}, 1000, "", today, func(*entity.MemberStreak) {})
	if err != nil || len(multi.Notifications) != 5 {
		t.Fatalf("multiple thresholds: %+v %v", multi, err)
	}
	down, _ := os.ReadFile("../../db/migrations_postgre/20261112000001_completion_notifications.down.sql")
	exec(string(down))
	up, _ := os.ReadFile("../../db/migrations_postgre/20261112000001_completion_notifications.up.sql")
	exec(string(up))
}
func mustID(t *testing.T, s string) int64 {
	t.Helper()
	var n int64
	if _, err := fmt.Sscan(s, &n); err != nil {
		t.Fatal(err)
	}
	return n
}

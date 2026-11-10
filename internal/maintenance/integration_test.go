package maintenance

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// Tests create and drop a dedicated database; never mutate existing app tables.
func testDB(t *testing.T) (*pgx.Conn, *pgx.ConnConfig) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a PostgreSQL role with CREATEDB")
	}
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("arsiva_cleanup_test_%d", time.Now().UnixNano())
	if _, err = admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
		admin.Close(ctx)
		t.Fatal(err)
	}
	cfg := admin.Config().Copy()
	cfg.Database = name
	t.Cleanup(func() {
		admin.Exec(ctx, "DROP DATABASE "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)")
		admin.Close(ctx)
	})
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close(ctx) })
	execSQL(t, conn, `CREATE TYPE content_type_enum AS ENUM ('kuis','cerita','puzzle');
 CREATE TABLE assets (asset_id serial PRIMARY KEY, url text NOT NULL, is_used boolean DEFAULT false, created_at timestamp DEFAULT now(), deleted_at timestamp);
 CREATE TABLE kuis (kuis_id integer PRIMARY KEY, thumbnail_asset_id integer REFERENCES assets);
 CREATE TABLE cerita_interaktif (cerita_id integer PRIMARY KEY);
 CREATE TABLE puzzles (puzzle_id integer PRIMARY KEY);
 CREATE TABLE group_contents (group_content_id serial PRIMARY KEY, group_id text, content_type content_type_enum, content_id integer);
 CREATE TABLE artikel (konten text);`)
	return conn, cfg
}

func execSQL(t *testing.T, conn *pgx.Conn, sql string) {
	t.Helper()
	if _, err := conn.Exec(context.Background(), sql); err != nil {
		t.Fatal(err)
	}
}
func count(t *testing.T, conn *pgx.Conn, sql string) int {
	t.Helper()
	var n int
	if err := conn.QueryRow(context.Background(), sql).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
func install(t *testing.T, conn *pgx.Conn) {
	t.Helper()
	sql, err := os.ReadFile("../../db/migrations_postgre/20261110000001_group_content_integrity.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	execSQL(t, conn, string(sql))
}

func TestGroupContentIntegrity(t *testing.T) {
	conn, cfg := testDB(t)
	execSQL(t, conn, `INSERT INTO group_contents(group_id,content_type,content_id) VALUES ('old','kuis',999);`)
	install(t, conn)
	if count(t, conn, "SELECT count(*) FROM group_contents") != 0 {
		t.Fatal("historical orphan retained")
	}
	for _, item := range []struct{ kind, table, id string }{{"kuis", "kuis", "kuis_id"}, {"cerita", "cerita_interaktif", "cerita_id"}, {"puzzle", "puzzles", "puzzle_id"}} {
		execSQL(t, conn, fmt.Sprintf("INSERT INTO %s(%s) VALUES (5), (6)", item.table, item.id))
		execSQL(t, conn, fmt.Sprintf("INSERT INTO group_contents(group_id,content_type,content_id) VALUES ('a','%s',5),('b','%s',5)", item.kind, item.kind))
	}
	execSQL(t, conn, "BEGIN; DELETE FROM kuis WHERE kuis_id=5; ROLLBACK;")
	if count(t, conn, "SELECT count(*) FROM group_contents") != 6 {
		t.Fatal("rollback lost relations")
	}
	execSQL(t, conn, "DELETE FROM kuis WHERE kuis_id=5")
	if count(t, conn, "SELECT count(*) FROM group_contents") != 4 {
		t.Fatal("wrong-type relations deleted or kuis relations retained")
	}
	execSQL(t, conn, "DELETE FROM cerita_interaktif WHERE cerita_id=5; DELETE FROM puzzles WHERE puzzle_id=5; DELETE FROM puzzles WHERE puzzle_id=6;")
	if count(t, conn, "SELECT count(*) FROM group_contents") != 0 {
		t.Fatal("relations retained")
	}
	if _, err := conn.Exec(context.Background(), "INSERT INTO group_contents(content_type,content_id) VALUES ('kuis',999)"); err == nil {
		t.Fatal("missing target accepted")
	}
	execSQL(t, conn, "INSERT INTO group_contents(content_type,content_id) VALUES ('kuis',6)")
	if _, err := conn.Exec(context.Background(), "UPDATE group_contents SET content_id=999"); err == nil {
		t.Fatal("invalid update accepted")
	}
	execSQL(t, conn, `CREATE FUNCTION fail_relation_delete() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'simulated cleanup failure'; END $$;
 CREATE TRIGGER fail_cleanup BEFORE DELETE ON group_contents FOR EACH ROW EXECUTE FUNCTION fail_relation_delete();`)
	if _, err := conn.Exec(context.Background(), "DELETE FROM kuis WHERE kuis_id=6"); err == nil {
		t.Fatal("cleanup failure did not reject parent deletion")
	}
	if count(t, conn, "SELECT count(*) FROM kuis WHERE kuis_id=6") != 1 || count(t, conn, "SELECT count(*) FROM group_contents") != 1 {
		t.Fatal("cleanup failure did not roll back atomically")
	}
	execSQL(t, conn, "DROP TRIGGER fail_cleanup ON group_contents")

	// Force the insert-before-delete interleaving. DELETE waits for the target
	// lock, then must see and remove the newly committed relation.
	other, err := pgx.ConnectConfig(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close(context.Background())
	execSQL(t, conn, "BEGIN; INSERT INTO group_contents(content_type,content_id) VALUES ('kuis',6)")
	done := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := other.Exec(ctx, "DELETE FROM kuis WHERE kuis_id=6")
		done <- err
	}()
	select {
	case err := <-done:
		t.Fatalf("delete did not wait for insert: %v", err)
	case <-time.After(150 * time.Millisecond):
	}
	execSQL(t, conn, "COMMIT")
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if count(t, conn, "SELECT count(*) FROM group_contents") != 0 {
		t.Fatal("concurrent insert left orphan")
	}
	// Reverse interleaving: insert waits for delete, then rejects the missing target.
	execSQL(t, conn, "INSERT INTO kuis(kuis_id) VALUES (7)")
	execSQL(t, conn, "BEGIN; DELETE FROM kuis WHERE kuis_id=7")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := other.Exec(ctx, "INSERT INTO group_contents(content_type,content_id) VALUES ('kuis',7)")
		done <- err
	}()
	select {
	case err := <-done:
		t.Fatalf("insert did not wait for delete: %v", err)
	case <-time.After(150 * time.Millisecond):
	}
	execSQL(t, conn, "COMMIT")
	if err := <-done; err == nil {
		t.Fatal("concurrent insert accepted deleted target")
	}
	if count(t, conn, "SELECT count(*) FROM group_contents") != 0 {
		t.Fatal("reverse interleaving left orphan")
	}
	down, err := os.ReadFile("../../db/migrations_postgre/20261110000001_group_content_integrity.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	execSQL(t, conn, string(down))
	install(t, conn)
}

func TestCleanupDryRunApplyAndRetry(t *testing.T) {
	conn, _ := testDB(t)
	dir := t.TempDir()
	old := time.Now().Add(-30 * 24 * time.Hour)
	for _, name := range []string{"live.webp", "stale.webp", "embedded.webp", "unregistered.webp", "recent.webp", "missing-reference.webp"} {
		file := filepath.Join(dir, name)
		if err := os.WriteFile(file, []byte("image"), 0600); err != nil {
			t.Fatal(err)
		}
		if name != "recent.webp" {
			if err := os.Chtimes(file, old, old); err != nil {
				t.Fatal(err)
			}
		}
	}
	execSQL(t, conn, `INSERT INTO assets(asset_id,url,is_used,created_at) VALUES
 (1,'/uploads/live.webp',false,now()-interval '30 days'),
 (2,'/uploads/stale.webp',true,now()-interval '30 days'),
 (3,'/uploads/embedded.webp',false,now()-interval '30 days'),
 (4,'/uploads/recent.webp',false,now()-interval '30 days'),
 (5,'/uploads/already-missing.webp',true,now()-interval '30 days');
 INSERT INTO kuis VALUES (1,1);
 INSERT INTO artikel VALUES ('<img src="/uploads/embedded.webp">'), ('{"scene":"/uploads/missing-reference.webp"}');
 INSERT INTO group_contents(content_type,content_id) VALUES ('kuis',99),('kuis',1);`)
	var output bytes.Buffer
	opts := Options{UploadDir: dir, MinAge: 7 * 24 * time.Hour, Output: &output}
	if err := Run(context.Background(), conn, opts); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "stale.webp")); err != nil {
		t.Fatal("dry-run removed file")
	}
	if count(t, conn, "SELECT count(*) FROM group_contents") != 2 {
		t.Fatal("dry-run mutated DB")
	}
	opts.Apply = true
	for i := 0; i < 2; i++ {
		if err := Run(context.Background(), conn, opts); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"live.webp", "embedded.webp", "recent.webp", "missing-reference.webp"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("protected %s removed: %v", name, err)
		}
	}
	for _, name := range []string{"stale.webp", "unregistered.webp"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("orphan %s retained", name)
		}
	}
	if count(t, conn, "SELECT count(*) FROM assets WHERE deleted_at IS NOT NULL") != 2 {
		t.Fatal("wrong assets marked deleted")
	}
	if count(t, conn, "SELECT count(*) FROM group_contents") != 1 {
		t.Fatal("wrong relation cleanup")
	}
}

func TestProductionSchemaMigrations(t *testing.T) {
	conn, _ := testDB(t)
	// This database belongs exclusively to this test.
	execSQL(t, conn, "DROP SCHEMA public CASCADE; CREATE SCHEMA public")
	files, err := filepath.Glob("../../db/migrations_postgre/*.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no migrations found")
	}
	for _, file := range files {
		// Historical migration replaces an existing FK without dropping it.
		// Normalize only this disposable fixture; do not modify deployed history.
		if filepath.Base(file) == "20260511000001_add_fk_member_social_links.up.sql" {
			execSQL(t, conn, "ALTER TABLE member_social_links DROP CONSTRAINT member_social_links_member_id_fkey")
		}
		sql, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := conn.Exec(context.Background(), string(sql)); err != nil {
			t.Fatalf("migration %s: %v", filepath.Base(file), err)
		}
	}
	if err := Run(context.Background(), conn, Options{UploadDir: t.TempDir(), MinAge: 7 * 24 * time.Hour, Apply: true}); err != nil {
		t.Fatal(err)
	}
}

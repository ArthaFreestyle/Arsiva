// Package maintenance contains offline, explicit maintenance operations.
package maintenance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Options struct {
	UploadDir string
	MinAge    time.Duration
	Apply     bool
	Output    io.Writer
}

type candidate struct {
	url       string
	ids       []int
	protected bool
}

// LocalPath accepts only canonical local upload URLs, never arbitrary paths.
func LocalPath(url string) (string, error) {
	if !strings.HasPrefix(url, "/uploads/") {
		return "", fmt.Errorf("not a local upload URL: %q", url)
	}
	rel := strings.TrimPrefix(url, "/uploads/")
	if rel == "" || rel == "." || strings.ContainsAny(rel, "\\:%?#\x00") || path.Clean(rel) != rel || !fs.ValidPath(rel) {
		return "", fmt.Errorf("unsafe upload URL: %q", url)
	}
	return rel, nil
}

// Run requires writers (API, workers, uploads) to be stopped during apply.
// SQL locks protect database references; they cannot stop an in-flight upload
// from writing a file before registering its assets row.
func Run(ctx context.Context, conn *pgx.Conn, opts Options) error {
	if opts.MinAge < 0 {
		return fmt.Errorf("min-age must not be negative")
	}
	if opts.Output == nil {
		opts.Output = io.Discard
	}
	root, err := os.OpenRoot(opts.UploadDir)
	if err != nil {
		return err
	}
	defer root.Close()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(ctx, "SET LOCAL lock_timeout = '10s'"); err != nil {
		return err
	}

	// Inspect every public application table, including embedded URLs in JSON
	// or rich text. Discover FK references instead of trusting assets.is_used.
	rows, err := tx.Query(ctx, `SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename`)
	if err != nil {
		return err
	}
	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return err
	}
	if opts.Apply {
		names := make([]string, len(tables))
		for i, table := range tables {
			names[i] = pgx.Identifier{"public", table}.Sanitize()
		}
		if len(names) == 0 {
			return fmt.Errorf("no public application tables found")
		}
		if _, err = tx.Exec(ctx, "LOCK TABLE "+strings.Join(names, ", ")+" IN SHARE ROW EXCLUSIVE MODE"); err != nil {
			return err
		}
	}
	cutoff := time.Now().Add(-opts.MinAge)
	candidates := map[string]*candidate{}
	rows, err = tx.Query(ctx, `SELECT asset_id, url, created_at FROM public.assets`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var url string
		var created time.Time
		if err = rows.Scan(&id, &url, &created); err != nil {
			rows.Close()
			return err
		}
		if _, err = LocalPath(url); err != nil {
			fmt.Fprintf(opts.Output, "SKIP %s\n", url)
			continue
		}
		c := candidates[url]
		if c == nil {
			c = &candidate{url: url}
			candidates[url] = c
		}
		c.ids = append(c.ids, id)
		c.protected = c.protected || !created.Before(cutoff)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	err = fs.WalkDir(root.FS(), ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		url := "/uploads/" + filepath.ToSlash(name)
		if _, err := LocalPath(url); err != nil {
			return nil
		}
		c := candidates[url]
		if c == nil {
			c = &candidate{url: url}
			candidates[url] = c
		}
		c.protected = c.protected || !info.ModTime().Before(cutoff)
		return nil
	})
	if err != nil {
		return err
	}

	rows, err = tx.Query(ctx, `SELECT ns.nspname, cl.relname, att.attname
 FROM pg_constraint fk
 JOIN pg_class cl ON cl.oid = fk.conrelid
 JOIN pg_namespace ns ON ns.oid = cl.relnamespace
 JOIN pg_attribute att ON att.attrelid = fk.conrelid AND att.attnum = fk.conkey[1]
 WHERE fk.contype = 'f' AND fk.confrelid = 'public.assets'::regclass`)
	if err != nil {
		return err
	}
	type reference struct{ schema, table, column string }
	var refs []reference
	for rows.Next() {
		var ref reference
		if err = rows.Scan(&ref.schema, &ref.table, &ref.column); err != nil {
			rows.Close()
			return err
		}
		refs = append(refs, ref)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, ref := range refs {
		query := `SELECT DISTINCT a.url FROM public.assets a JOIN ` + pgx.Identifier{ref.schema, ref.table}.Sanitize() + ` t ON t.` + pgx.Identifier{ref.column}.Sanitize() + ` = a.asset_id`
		rows, err = tx.Query(ctx, query)
		if err != nil {
			return err
		}
		for rows.Next() {
			var url string
			if err = rows.Scan(&url); err != nil {
				rows.Close()
				return err
			}
			if c := candidates[url]; c != nil {
				c.protected = true
			}
		}
		if err = rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
	}
	// JSON decoding also handles escaped slashes and nested scene/article data.
	for _, table := range tables {
		if table == "assets" {
			continue
		}
		rows, err = tx.Query(ctx, `SELECT to_jsonb(t) FROM `+pgx.Identifier{"public", table}.Sanitize()+` t`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var raw []byte
			if err = rows.Scan(&raw); err != nil {
				rows.Close()
				return err
			}
			var value any
			if err = json.Unmarshal(raw, &value); err != nil {
				rows.Close()
				return err
			}
			protectURLs(value, candidates)
		}
		if err = rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
	}
	// Finish database validation before making any irreversible file changes.
	predicate := `(content_type = 'kuis' AND NOT EXISTS (SELECT 1 FROM kuis WHERE kuis_id = gc.content_id))
 OR (content_type = 'cerita' AND NOT EXISTS (SELECT 1 FROM cerita_interaktif WHERE cerita_id = gc.content_id))
 OR (content_type = 'puzzle' AND NOT EXISTS (SELECT 1 FROM puzzles WHERE puzzle_id = gc.content_id))`
	var relations int
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM group_contents gc WHERE `+predicate).Scan(&relations); err != nil {
		return err
	}
	var urls []string
	for url, c := range candidates {
		if !c.protected {
			urls = append(urls, url)
		}
	}
	sort.Strings(urls)
	for _, url := range urls {
		rel, _ := LocalPath(url)
		// Reject symlinks in every component, including an assets-only path.
		if err = checkPath(root, rel); err != nil {
			return err
		}
	}
	for _, url := range urls {
		rel, _ := LocalPath(url)
		fmt.Fprintf(opts.Output, "ORPHAN %s\n", url)
		if !opts.Apply {
			continue
		}
		if err = root.Remove(rel); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", url, err)
		}
		if len(candidates[url].ids) > 0 {
			if _, err = tx.Exec(ctx, `UPDATE public.assets SET deleted_at = COALESCE(deleted_at, NOW()), is_used = false WHERE asset_id = ANY($1)`, candidates[url].ids); err != nil {
				return err
			}
		}
	}
	fmt.Fprintf(opts.Output, "Orphan files/assets: %d; orphan group relations: %d; apply: %t\n", len(urls), relations, opts.Apply)
	if !opts.Apply {
		return nil
	}
	if _, err = tx.Exec(ctx, `DELETE FROM group_contents gc WHERE `+predicate); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func protectURLs(value any, candidates map[string]*candidate) {
	switch v := value.(type) {
	case string:
		if !strings.Contains(v, "/uploads/") {
			return
		}
		for url, c := range candidates {
			if !c.protected && strings.Contains(v, url) {
				c.protected = true
			}
		}
	case []any:
		for _, item := range v {
			protectURLs(item, candidates)
		}
	case map[string]any:
		for _, item := range v {
			protectURLs(item, candidates)
		}
	}
}

func checkPath(root *os.Root, rel string) error {
	parts := strings.Split(rel, "/")
	for i := range parts {
		info, err := root.Lstat(strings.Join(parts[:i+1], "/"))
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing symlink: %s", rel)
		}
		if i == len(parts)-1 && !info.Mode().IsRegular() {
			return fmt.Errorf("refusing non-regular file: %s", rel)
		}
	}
	return nil
}

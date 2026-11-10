package maintenance

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalPath(t *testing.T) {
	for _, url := range []string{"/uploads/../secret", "/uploads/a/../../secret", "/uploads//secret", "/uploads/.", "/uploads/", "/uploads/a\\b", "/uploads/C:/secret", "https://example.org/file", "/uploads/%2e%2e/secret"} {
		if _, err := LocalPath(url); err == nil {
			t.Errorf("accepted unsafe path %q", url)
		}
	}
	if got, err := LocalPath("/uploads/2026/04/image.webp"); err != nil || got != "2026/04/image.webp" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestProtectEmbeddedURLs(t *testing.T) {
	candidates := map[string]*candidate{
		"/uploads/live.webp":   {url: "/uploads/live.webp"},
		"/uploads/orphan.webp": {url: "/uploads/orphan.webp"},
	}
	protectURLs(map[string]any{"scene": []any{map[string]any{"html": `<img src="https://arsiva.id/uploads/live.webp">`}}}, candidates)
	if !candidates["/uploads/live.webp"].protected || candidates["/uploads/orphan.webp"].protected {
		t.Fatal("embedded URL protection incorrect")
	}
}

func TestCheckPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "folder"), 0700); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := checkPath(root, "folder"); err == nil {
		t.Fatal("accepted directory as a file")
	}
	if err := checkPath(root, "missing/file.webp"); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(dir, "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := checkPath(root, "link/file.webp"); err == nil {
		t.Fatal("accepted symlink parent")
	}
}

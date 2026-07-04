package devreload

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/puppe1990/cais/pkg/cais/devlog"
)

// Paths lists filesystem locations that should trigger a browser reload in development.
type Paths struct {
	TemplatesDir string
	CSSPath      string
}

// Stamp returns a value that changes when the server restarts or watched files change.
func Stamp(paths Paths, startedAt int64) string {
	return fmt.Sprintf("%d-%d-%d", startedAt, treeModTime(paths.TemplatesDir, ".html"), fileModTime(paths.CSSPath))
}

type registrar interface {
	Get(pattern string, handler http.HandlerFunc)
}

// Register mounts GET /dev/reload for development hot-reload polling (localhost only).
func Register(r registrar, env string, paths Paths, startedAt int64) {
	if env != "development" {
		r.Get("/dev/reload", func(w http.ResponseWriter, req *http.Request) {
			http.NotFound(w, req)
		})
		return
	}

	handler := devlog.LocalOnly(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprint(w, Stamp(paths, startedAt))
	}))
	r.Get("/dev/reload", handler.ServeHTTP)
}

func fileModTime(path string) int64 {
	if path == "" {
		return 0
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.ModTime().UnixNano()
}

func treeModTime(dir, ext string) int64 {
	if dir == "" {
		return 0
	}
	var latest int64
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ext {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if m := info.ModTime().UnixNano(); m > latest {
			latest = m
		}
		return nil
	})
	return latest
}

package devreload

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestStamp_includesServerStart(t *testing.T) {
	started := int64(1700000000000000000)
	s1 := Stamp(Paths{}, started)
	s2 := Stamp(Paths{}, started+1)
	if s1 == s2 {
		t.Fatalf("stamp should change when server start changes: %q", s1)
	}
}

func TestStamp_changesWhenTemplateModified(t *testing.T) {
	dir := t.TempDir()
	templates := filepath.Join(dir, "templates")
	if err := os.MkdirAll(templates, 0o755); err != nil {
		t.Fatal(err)
	}
	page := filepath.Join(templates, "home.html")
	if err := os.WriteFile(page, []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	paths := Paths{TemplatesDir: templates}
	started := time.Now().UnixNano()
	s1 := Stamp(paths, started)

	time.Sleep(20 * time.Millisecond)
	if err := os.WriteFile(page, []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	s2 := Stamp(paths, started)
	if s1 == s2 {
		t.Fatalf("stamp should change when template changes: %q", s1)
	}
}

func TestStamp_changesWhenCSSModified(t *testing.T) {
	dir := t.TempDir()
	css := filepath.Join(dir, "styles.css")
	if err := os.WriteFile(css, []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	paths := Paths{CSSPath: css}
	started := time.Now().UnixNano()
	s1 := Stamp(paths, started)

	time.Sleep(20 * time.Millisecond)
	if err := os.WriteFile(css, []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	s2 := Stamp(paths, started)
	if s1 == s2 {
		t.Fatalf("stamp should change when css changes: %q", s1)
	}
}

func TestRegister_development_returnsStamp(t *testing.T) {
	dir := t.TempDir()
	templates := filepath.Join(dir, "templates")
	if err := os.MkdirAll(templates, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templates, "a.html"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := cais.NewRouter()
	Register(r, "development", Paths{TemplatesDir: templates}, 42)

	req := httptest.NewRequest(http.MethodGet, "/dev/reload", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if rr.Body.Len() == 0 {
		t.Fatal("expected non-empty stamp")
	}
	if cc := rr.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", cc)
	}
}

func TestRegister_production_returns404(t *testing.T) {
	r := cais.NewRouter()
	Register(r, "production", Paths{}, 42)

	req := httptest.NewRequest(http.MethodGet, "/dev/reload", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}

func TestRegister_nonLoopback_returns403(t *testing.T) {
	r := cais.NewRouter()
	Register(r, "development", Paths{}, 42)

	req := httptest.NewRequest(http.MethodGet, "/dev/reload", nil)
	req.RemoteAddr = "203.0.113.1:12345"
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
}

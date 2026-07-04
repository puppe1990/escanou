package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"

	"github.com/puppe1990/escanou/internal/store"
)

func setupTestApp(t *testing.T) *App {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("go.mod not found")
		}
		wd = parent
	}
	catalog := i18n.DefaultCatalog()
	renderer, err := cais.NewRendererFromDir(filepath.Join(wd, "web", "templates"), catalog)
	if err != nil {
		t.Fatal(err)
	}
	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	cfg := cais.Config{Port: ":0", DBPath: ":memory:", Env: "test"}
	a, err := New(cfg, Deps{
		Renderer:  renderer,
		Store:     s,
		StaticDir: filepath.Join(wd, "web", "static"),
		Site:      meta.SiteFrom("Escanou", ""),
		Catalog:   catalog,
	})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestApp_Supermarket_requiresAuth(t *testing.T) {
	a := setupTestApp(t)
	h := a.Handler()

	for _, path := range []string{"/", "/feed", "/map", "/achievements", "/nfce"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != http.StatusSeeOther {
				t.Fatalf("status = %d, want 303", rr.Code)
			}
			if rr.Header().Get("Location") != "/login" {
				t.Fatalf("Location = %q, want /login", rr.Header().Get("Location"))
			}
		})
	}
}

func TestApp_Supermarket_HTMX_requiresAuth(t *testing.T) {
	a := setupTestApp(t)
	h := a.Handler()

	req := httptest.NewRequest(http.MethodGet, "/feed", nil)
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
	if rr.Header().Get("HX-Redirect") != "/login" {
		t.Fatalf("HX-Redirect = %q, want /login", rr.Header().Get("HX-Redirect"))
	}
}

func TestApp_Login_public(t *testing.T) {
	a := setupTestApp(t)
	h := a.Handler()

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /login status = %d, want 200", rr.Code)
	}
}

func TestApp_DevReload_availableInDevelopment(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("go.mod not found")
		}
		wd = parent
	}
	catalog := i18n.DefaultCatalog()
	templatesDir := filepath.Join(wd, "web", "templates")
	renderer, err := cais.NewRendererFromDir(templatesDir, catalog)
	if err != nil {
		t.Fatal(err)
	}
	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	staticDir := filepath.Join(wd, "web", "static")
	a, err := New(cais.Config{Port: ":0", DBPath: ":memory:", Env: "development"}, Deps{
		Renderer:     renderer,
		Store:        s,
		StaticDir:    staticDir,
		TemplatesDir: templatesDir,
		Site:         meta.SiteFrom("Escanou", ""),
		Catalog:      catalog,
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/dev/reload", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()
	a.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /dev/reload status = %d, want 200", rr.Code)
	}
	if rr.Body.Len() == 0 {
		t.Fatal("expected reload stamp body")
	}
}

package handlers

import (
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	caisi18n "github.com/puppe1990/cais/pkg/cais/i18n"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/testutil"

	appi18n "github.com/puppe1990/mercado/internal/i18n"
	"github.com/puppe1990/mercado/internal/store"
)

func setupTestRenderer(t *testing.T) *cais.Renderer {
	t.Helper()
	return testutil.NewRenderer(t)
}

func setupTestStore(t *testing.T) store.Store {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func testSite() meta.Site {
	return meta.Site{AppName: "mercado", AppURL: "https://example.com"}
}

func testCatalog() *caisi18n.Catalog {
	return appi18n.DefaultCatalog()
}

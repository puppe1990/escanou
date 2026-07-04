package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestHomeHandler_Returns200(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHomeHandler_ContainsWelcome(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "on Cais!") {
		t.Errorf("body missing welcome message, got: %s", rr.Body.String())
	}
}

func TestHomeHandler_ContentType(t *testing.T) {
	h := NewHomeHandler(setupTestRenderer(t), testSite(), testCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

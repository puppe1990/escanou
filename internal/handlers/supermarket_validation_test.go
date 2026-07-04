package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/barcode"
	"github.com/puppe1990/cais/pkg/cais/session"
)

func TestSupermarket_LookupPost_invalidBarcode(t *testing.T) {
	s := setupTestStore(t)
	h := NewSupermarketHandler(setupTestRenderer(t), s, testSite(), cais.Config{Env: "development"})

	form := url.Values{"barcode": {"1234567890123"}}
	req := httptest.NewRequest(http.MethodPost, "/scan/lookup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req = session.WithUserID(req, 1)

	rr := httptest.NewRecorder()
	h.LookupPost(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Código inválido") {
		t.Errorf("body = %s", rr.Body.String())
	}
}

func TestSupermarket_LookupPost_offUnavailable(t *testing.T) {
	offSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "offline", http.StatusServiceUnavailable)
	}))
	t.Cleanup(offSrv.Close)

	s := setupTestStore(t)
	h := NewSupermarketHandler(setupTestRenderer(t), s, testSite(), cais.Config{Env: "development"})
	h.barcode = &barcode.Client{BaseURL: offSrv.URL, HTTP: offSrv.Client()}

	// Valid EAN not in local DB
	form := url.Values{"barcode": {"5901234123457"}}
	req := httptest.NewRequest(http.MethodPost, "/scan/lookup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req = session.WithUserID(req, 1)

	rr := httptest.NewRecorder()
	h.LookupPost(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "informe o nome") {
		t.Errorf("body = %s", body)
	}
	if !strings.Contains(body, `name="name"`) {
		t.Error("should show name form when OFF unavailable")
	}
}

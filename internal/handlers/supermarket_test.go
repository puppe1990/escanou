package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/session"
)

func TestSupermarketHandler_pagesRender(t *testing.T) {
	h := NewSupermarketHandler(setupTestRenderer(t), setupTestStore(t), testSite(), cais.Config{Env: "development"})

	for _, tc := range []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
		want string
	}{
		{"scan", h.Scan, "Abrir câmera"},
		{"map", h.Map, "Mapa de Ofertas"},
		{"feed", h.Feed, "Feed Comunitário"},
		{"achievements", h.Achievements, "Conquistas"},
		{"nfce", h.NFCe, "Importar Nota Fiscal"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = session.WithUserID(req, 1)
			rr := httptest.NewRecorder()
			tc.fn(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d", rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tc.want) {
				t.Errorf("body missing %q", tc.want)
			}
			if !strings.Contains(rr.Body.String(), "Beta Colaborativo") {
				t.Error("body missing supermarket layout branding")
			}
		})
	}
}

func TestSupermarketHandler_scan_lookupSpinner(t *testing.T) {
	h := NewSupermarketHandler(setupTestRenderer(t), setupTestStore(t), testSite(), cais.Config{Env: "development"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = session.WithUserID(req, 1)
	rr := httptest.NewRecorder()
	h.Scan(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `id="scan-lookup-spinner"`) {
		t.Error("scan page should include lookup spinner")
	}
	if !strings.Contains(body, "Buscando produto") {
		t.Error("scan page should show lookup loading message")
	}
	if !strings.Contains(body, `hx-indicator="#scan-lookup-spinner"`) {
		t.Error("lookup form should reference spinner indicator")
	}
}

func TestSupermarketHandler_scan_noDemoRapido(t *testing.T) {
	h := NewSupermarketHandler(setupTestRenderer(t), setupTestStore(t), testSite(), cais.Config{Env: "development"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = session.WithUserID(req, 1)
	rr := httptest.NewRecorder()
	h.Scan(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	body := rr.Body.String()
	if strings.Contains(body, "Demo Rápido") {
		t.Error("scan page should not contain Demo Rápido section")
	}
}

func TestSupermarketHandler_scan_emptyHistory(t *testing.T) {
	h := NewSupermarketHandler(setupTestRenderer(t), setupTestStore(t), testSite(), cais.Config{Env: "development"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = session.WithUserID(req, 99)
	rr := httptest.NewRecorder()
	h.Scan(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Nenhum preço registrado ainda") {
		t.Error("scan page should show empty history message for user without reports")
	}
}

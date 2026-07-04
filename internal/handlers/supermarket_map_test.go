package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/session"
)

func TestSupermarketHandler_map_rendersLeafletMap(t *testing.T) {
	s := setupTestStore(t)
	if err := s.SeedMercadoDemo(); err != nil {
		t.Fatal(err)
	}
	h := NewSupermarketHandler(setupTestRenderer(t), s, testSite(), cais.Config{Env: "development"})

	req := httptest.NewRequest(http.MethodGet, "/map", nil)
	req = session.WithUserID(req, 1)
	rr := httptest.NewRecorder()
	h.Map(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	body := rr.Body.String()
	if strings.Contains(body, "requer integração") {
		t.Error("map should not show placeholder integration message")
	}
	if !strings.Contains(body, `id="mercado-map"`) {
		t.Error("map page should include mercado-map container")
	}
	if !strings.Contains(body, "mercado-map-data") {
		t.Error("map page should embed marker JSON")
	}
	if !strings.Contains(body, "Atacadão Tatuapé") || !strings.Contains(body, "-23.54") {
		t.Error("map markers should include seeded supermarkets with coordinates")
	}
}

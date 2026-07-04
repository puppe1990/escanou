package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
)

func TestDashboardHandler_Returns200(t *testing.T) {
	h := NewDashboardHandler(setupTestRenderer(t), setupTestStore(t), testSite(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestDashboardHandler_ContainsDashboard(t *testing.T) {
	h := NewDashboardHandler(setupTestRenderer(t), setupTestStore(t), testSite(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "Dashboard") {
		t.Errorf("body missing Dashboard, got: %s", rr.Body.String())
	}
}

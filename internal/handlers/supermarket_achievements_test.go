package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/middleware"
	"github.com/puppe1990/cais/pkg/cais/session"

	"github.com/puppe1990/escanou/internal/store"
)

func TestLevelProgress(t *testing.T) {
	for _, tc := range []struct {
		points  int
		wantXP  int
		wantMax int
		wantPct int
	}{
		{0, 0, 100, 0},
		{35, 35, 100, 35},
		{100, 0, 100, 0},
		{135, 35, 100, 35},
		{250, 50, 100, 50},
	} {
		xp, max, pct := levelProgress(tc.points)
		if xp != tc.wantXP || max != tc.wantMax || pct != tc.wantPct {
			t.Errorf("levelProgress(%d) = (%d,%d,%d) want (%d,%d,%d)",
				tc.points, xp, max, pct, tc.wantXP, tc.wantMax, tc.wantPct)
		}
	}
}

func TestSupermarketHandler_achievements_realProgress(t *testing.T) {
	s := setupTestStore(t)
	if err := s.SeedMercadoDemo(); err != nil {
		t.Fatal(err)
	}
	st := s.(*store.SQLiteStore)

	userID, err := st.CreateUser("ach@test.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	pid, err := st.CreateProduct("Leite Teste", "9990001112223", "Laticínios")
	if err != nil {
		t.Fatal(err)
	}
	markets, err := st.ListSupermarkets()
	if err != nil || len(markets) == 0 {
		t.Fatalf("markets: %v len=%d", err, len(markets))
	}
	for i := 0; i < 3; i++ {
		if _, err := st.CreatePriceReport(userID, pid, markets[0].ID, 500+i); err != nil {
			t.Fatal(err)
		}
	}

	h := NewSupermarketHandler(setupTestRenderer(t), s, testSite(), cais.Config{Env: "development"})
	handler := middleware.LoadUserStats(s)(http.HandlerFunc(h.Achievements))

	req := httptest.NewRequest(http.MethodGet, "/achievements", nil)
	req = session.WithUserID(req, userID)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	body := rr.Body.String()
	if strings.Contains(body, "20 / 200 XP") {
		t.Error("achievements should not show hardcoded XP")
	}
	if strings.Contains(body, "Desde Jul/2026") {
		t.Error("achievements should not show hardcoded join date")
	}
	if !strings.Contains(body, "30 / 100 XP") {
		t.Errorf("body missing real XP progress, got fragment around XP: %s", excerpt(body, "XP"))
	}
	if !strings.Contains(body, `style="width: 30%"`) && !strings.Contains(body, `width: 30%`) {
		t.Error("XP bar should reflect 30% progress for 30 points")
	}
	if !strings.Contains(body, "medalhas desbloqueadas") {
		t.Error("body should show unlocked badge count label")
	}
}

func excerpt(s, needle string) string {
	i := strings.Index(s, needle)
	if i < 0 {
		return "(not found)"
	}
	start := i - 40
	if start < 0 {
		start = 0
	}
	end := i + 60
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}

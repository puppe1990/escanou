package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/session"

	"github.com/puppe1990/mercado/internal/store"
)

func TestShoppingListTotal(t *testing.T) {
	total := shoppingListTotalCents([]int{549, 399, 100})
	if total != 1048 {
		t.Fatalf("total = %d, want 1048", total)
	}
}

func TestSupermarketHandler_scan_shoppingListWithTotal(t *testing.T) {
	s := setupTestStore(t)
	if err := s.SeedMercadoDemo(); err != nil {
		t.Fatal(err)
	}
	st := s.(*store.SQLiteStore)

	userID, err := st.CreateUser("lista@test.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	pid, err := st.CreateProduct("Arroz", "1112223334445", "Mercearia")
	if err != nil {
		t.Fatal(err)
	}
	markets, err := st.ListSupermarkets()
	if err != nil || len(markets) == 0 {
		t.Fatalf("markets: %v", err)
	}
	if _, err := st.CreatePriceReport(userID, pid, markets[0].ID, 549); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreatePriceReport(userID, pid, markets[0].ID, 399); err != nil {
		t.Fatal(err)
	}

	h := NewSupermarketHandler(setupTestRenderer(t), s, testSite(), cais.Config{Env: "development"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = session.WithUserID(req, userID)
	rr := httptest.NewRecorder()
	h.Scan(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Lista de compras") {
		t.Error("scan page should show shopping list heading")
	}
	if !strings.Contains(body, "R$ 9,48") {
		t.Errorf("body missing list total, got: %s", excerpt(body, "Total"))
	}
	if !strings.Contains(body, "mercado-shopping-total") {
		t.Error("body should include shopping total block")
	}
}
package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/barcode"
	"github.com/puppe1990/cais/pkg/cais/session"
)

func TestSupermarket_LookupPost_unknownBarcode_showsNameForm(t *testing.T) {
	offSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status": 0}`))
	}))
	t.Cleanup(offSrv.Close)

	s := setupTestStore(t)
	h := NewSupermarketHandler(setupTestRenderer(t), s, testSite(), cais.Config{Env: "development"})
	h.barcode = &barcode.Client{BaseURL: offSrv.URL, HTTP: offSrv.Client()}

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
	if !strings.Contains(body, `name="name"`) {
		t.Errorf("partial should offer name field: %s", body)
	}
	if !strings.Contains(body, "5901234123457") {
		t.Error("partial should preserve scanned barcode")
	}
}

func TestSupermarket_ReportPost_HTMX_updatesHistory(t *testing.T) {
	s := setupTestStore(t)
	if err := s.SeedMercadoDemo(); err != nil {
		t.Fatal(err)
	}
	userID, err := s.CreateUser("report@test.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	product, found, err := s.FindProductByBarcode("7896256801011")
	if err != nil || !found {
		t.Fatal("need seeded product")
	}
	markets, err := s.ListSupermarkets()
	if err != nil || len(markets) == 0 {
		t.Fatal("need seeded supermarkets")
	}

	h := NewSupermarketHandler(setupTestRenderer(t), s, testSite(), cais.Config{Env: "development"})

	form := url.Values{
		"product_id":     {formatInt64(product.ID)},
		"supermarket_id": {formatInt64(markets[0].ID)},
		"price":          {"5,49"},
	}
	req := httptest.NewRequest(http.MethodPost, "/scan/report", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req = session.WithUserID(req, userID)

	rr := httptest.NewRecorder()
	h.ReportPost(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if strings.Contains(strings.ToLower(body), "<!doctype html>") {
		t.Error("HTMX response should be partial")
	}
	if !strings.Contains(body, "Leite Tirol") {
		t.Errorf("history should include reported product: %s", body)
	}
	if toast := rr.Header().Get("HX-Trigger"); !strings.Contains(toast, "pts") {
		t.Errorf("HX-Trigger toast = %q", toast)
	}
}

func TestSupermarket_ReportPost_invalidPrice_returns422Partial(t *testing.T) {
	s := setupTestStore(t)
	if err := s.SeedMercadoDemo(); err != nil {
		t.Fatal(err)
	}
	userID, err := s.CreateUser("bad@test.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	product, found, _ := s.FindProductByBarcode("7896256801011")
	if !found {
		t.Fatal("need product")
	}
	markets, _ := s.ListSupermarkets()

	h := NewSupermarketHandler(setupTestRenderer(t), s, testSite(), cais.Config{Env: "development"})

	form := url.Values{
		"product_id":     {formatInt64(product.ID)},
		"supermarket_id": {formatInt64(markets[0].ID)},
		"price":          {""},
	}
	req := httptest.NewRequest(http.MethodPost, "/scan/report", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req = session.WithUserID(req, userID)

	rr := httptest.NewRecorder()
	h.ReportPost(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Preencha supermercado e preço válidos") {
		t.Errorf("body = %s", rr.Body.String())
	}
}

func formatInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}

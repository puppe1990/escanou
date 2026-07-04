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

func TestSupermarket_LookupPost_HTMX_returnsPartial(t *testing.T) {
	s := setupTestStore(t)
	if err := s.SeedMercadoDemo(); err != nil {
		t.Fatal(err)
	}
	h := NewSupermarketHandler(setupTestRenderer(t), s, testSite(), cais.Config{Env: "development"})

	form := url.Values{"barcode": {"7896256801011"}}
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
	if strings.Contains(body, "<!doctype html>") {
		t.Error("HTMX response should be partial, not full page")
	}
	if !strings.Contains(body, "Leite Tirol") {
		t.Errorf("partial missing product name: %s", body)
	}
}

func TestSupermarket_LookupPost_unknownBarcode_usesOFF(t *testing.T) {
	offSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": 1,
			"product": {
				"product_name": "Produto OFF Teste",
				"categories": "Mercearia"
			}
		}`))
	}))
	t.Cleanup(offSrv.Close)

	s := setupTestStore(t)
	h := NewSupermarketHandler(setupTestRenderer(t), s, testSite(), cais.Config{Env: "development"})
	h.barcode = &barcode.Client{BaseURL: offSrv.URL, HTTP: offSrv.Client()}

	form := url.Values{"barcode": {"1234567890123"}}
	req := httptest.NewRequest(http.MethodPost, "/scan/lookup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req = session.WithUserID(req, 1)

	rr := httptest.NewRecorder()
	h.LookupPost(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Produto OFF Teste") {
		t.Errorf("body missing OFF product: %s", rr.Body.String())
	}
	_, found, err := s.FindProductByBarcode("1234567890123")
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Error("product should be persisted after OFF lookup")
	}
}

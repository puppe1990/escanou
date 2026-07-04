package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/session"
)

func TestSupermarketHandler_feed_voteButtons(t *testing.T) {
	s := setupTestStore(t)
	if err := s.SeedMercadoDemo(); err != nil {
		t.Fatal(err)
	}
	userID, err := s.CreateUser("feed-vote@test.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	pid, err := s.CreateProduct("Aveia", "8887776665554", "Mercearia")
	if err != nil {
		t.Fatal(err)
	}
	markets, _ := s.ListSupermarkets()
	if len(markets) == 0 {
		t.Fatal("need supermarkets")
	}
	if _, err := s.CreatePriceReport(userID, pid, markets[0].ID, 600); err != nil {
		t.Fatal(err)
	}
	h := NewSupermarketHandler(setupTestRenderer(t), s, testSite(), cais.Config{Env: "development"})

	req := httptest.NewRequest(http.MethodGet, "/feed", nil)
	req = session.WithUserID(req, 1)
	rr := httptest.NewRecorder()
	h.Feed(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "mercado-vote-up") {
		t.Error("feed should include green upvote button")
	}
	if !strings.Contains(body, "mercado-vote-down") {
		t.Error("feed should include red downvote button")
	}
}
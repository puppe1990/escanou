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

func TestSupermarketHandler_FlagPost_singleDispute_keepsFeedItem(t *testing.T) {
	s := setupTestStore(t)
	if err := s.SeedMercadoDemo(); err != nil {
		t.Fatal(err)
	}
	ownerID, err := s.CreateUser("owner@test.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	voterID, err := s.CreateUser("voter@test.com", "hash2")
	if err != nil {
		t.Fatal(err)
	}
	pid, err := s.CreateProduct("Leite", "1112223334445", "Laticínios")
	if err != nil {
		t.Fatal(err)
	}
	markets, _ := s.ListSupermarkets()
	reportID, err := s.CreatePriceReport(ownerID, pid, markets[0].ID, 499)
	if err != nil {
		t.Fatal(err)
	}

	h := NewSupermarketHandler(setupTestRenderer(t), s, testSite(), cais.Config{Env: "development"})
	req := httptest.NewRequest(http.MethodPost, "/feed/"+formatInt64(reportID)+"/flag", nil)
	req.Header.Set("HX-Request", "true")
	req = session.WithUserID(req, voterID)

	rr := httptest.NewRecorder()
	h.FlagPost(rr, req, reportID)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("HX-Reswap") == "delete" {
		t.Error("single dispute should not delete feed item")
	}
	body := rr.Body.String()
	if !strings.Contains(body, "feed-votes-") {
		t.Errorf("expected vote partial, got: %s", body)
	}
	if !strings.Contains(body, ">1<") {
		t.Errorf("dispute count should be 1: %s", body)
	}
	if !strings.Contains(body, "disabled") {
		t.Error("dispute button should be disabled after voting")
	}

	reports, err := s.ListFeedReports(10, voterID)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 {
		t.Fatalf("feed len = %d, want 1 after single dispute", len(reports))
	}
}

func TestSupermarketHandler_ConfirmPost_returnsVotePartial(t *testing.T) {
	s := setupTestStore(t)
	if err := s.SeedMercadoDemo(); err != nil {
		t.Fatal(err)
	}
	ownerID, err := s.CreateUser("owner2@test.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	voterID, err := s.CreateUser("voter2@test.com", "hash2")
	if err != nil {
		t.Fatal(err)
	}
	pid, err := s.CreateProduct("Pão", "5556667778889", "Padaria")
	if err != nil {
		t.Fatal(err)
	}
	markets, _ := s.ListSupermarkets()
	reportID, err := s.CreatePriceReport(ownerID, pid, markets[0].ID, 350)
	if err != nil {
		t.Fatal(err)
	}

	h := NewSupermarketHandler(setupTestRenderer(t), s, testSite(), cais.Config{Env: "development"})
	req := httptest.NewRequest(http.MethodPost, "/feed/"+formatInt64(reportID)+"/confirm", nil)
	req.Header.Set("HX-Request", "true")
	req = session.WithUserID(req, voterID)

	rr := httptest.NewRecorder()
	h.ConfirmPost(rr, req, reportID)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body: %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if strings.Contains(strings.ToLower(body), "<!doctype html>") {
		t.Error("confirm should return vote partial, not full page")
	}
	if !strings.Contains(body, ">1<") {
		t.Errorf("confirm count should be 1: %s", body)
	}
}
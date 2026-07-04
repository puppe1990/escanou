package store

import (
	"errors"
	"testing"
)

func TestDisputePriceReport_incrementsAndStaysInFeed(t *testing.T) {
	s := openTestStore(t)

	userID, err := s.CreateUser("owner@test.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	voterID, err := s.CreateUser("voter@test.com", "hash2")
	if err != nil {
		t.Fatal(err)
	}
	pid, _ := s.CreateProduct("Item", "111", "Geral")
	markets, _ := s.ListSupermarkets()
	reportID, err := s.CreatePriceReport(userID, pid, markets[0].ID, 500)
	if err != nil {
		t.Fatal(err)
	}

	count, err := s.DisputePriceReport(reportID, voterID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("disputes = %d, want 1", count)
	}

	reports, err := s.ListFeedReports(10, voterID)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 {
		t.Fatalf("feed len = %d, want 1 after single dispute", len(reports))
	}
	if reports[0].Disputes != 1 {
		t.Fatalf("Disputes = %d, want 1", reports[0].Disputes)
	}
	if !reports[0].ViewerDisputed {
		t.Error("ViewerDisputed should be true")
	}

	_, err = s.DisputePriceReport(reportID, voterID)
	if !errors.Is(err, ErrAlreadyDisputed) {
		t.Fatalf("expected ErrAlreadyDisputed, got %v", err)
	}
}
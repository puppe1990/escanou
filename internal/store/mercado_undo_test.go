package store

import (
	"errors"
	"testing"
)

func TestUndoPriceVote_removesConfirmation(t *testing.T) {
	s := openTestStore(t)

	ownerID, _ := s.CreateUser("owner@test.com", "hash")
	voterID, _ := s.CreateUser("voter@test.com", "hash2")
	pid, _ := s.CreateProduct("Leite", "111", "Laticínios")
	markets, _ := s.ListSupermarkets()
	reportID, _ := s.CreatePriceReport(ownerID, pid, markets[0].ID, 500)

	if _, err := s.ConfirmPriceReport(reportID, voterID); err != nil {
		t.Fatal(err)
	}

	confirms, disputes, err := s.UndoPriceVote(reportID, voterID)
	if err != nil {
		t.Fatal(err)
	}
	if confirms != 0 || disputes != 0 {
		t.Fatalf("counts = %d confirms, %d disputes, want 0/0", confirms, disputes)
	}

	reports, err := s.ListFeedReports(10, voterID)
	if err != nil {
		t.Fatal(err)
	}
	if reports[0].ViewerConfirmed || reports[0].ViewerDisputed {
		t.Error("viewer should have no vote after undo")
	}
}

func TestUndoPriceVote_removesDisputeAndUnflags(t *testing.T) {
	s := openTestStore(t)

	ownerID, _ := s.CreateUser("owner2@test.com", "hash")
	voterID, _ := s.CreateUser("voter2@test.com", "hash2")
	other1, _ := s.CreateUser("other1@test.com", "hash3")
	other2, _ := s.CreateUser("other2@test.com", "hash4")
	pid, _ := s.CreateProduct("Pão", "222", "Padaria")
	markets, _ := s.ListSupermarkets()
	reportID, _ := s.CreatePriceReport(ownerID, pid, markets[0].ID, 300)

	for _, uid := range []int64{voterID, other1, other2} {
		if _, err := s.DisputePriceReport(reportID, uid); err != nil {
			t.Fatal(err)
		}
	}

	if reportInFeed(s, voterID, reportID) {
		t.Fatal("flagged report should be hidden from feed")
	}

	if _, _, err := s.UndoPriceVote(reportID, voterID); err != nil {
		t.Fatal(err)
	}

	if !reportInFeed(s, voterID, reportID) {
		t.Fatal("report should return to feed after undo drops below flag threshold")
	}
	var disputes int
	if err := s.db.Raw().QueryRow(`SELECT disputes FROM price_reports WHERE id = ?`, reportID).Scan(&disputes); err != nil {
		t.Fatal(err)
	}
	if disputes != DisputeFlagThreshold-1 {
		t.Fatalf("disputes = %d, want %d", disputes, DisputeFlagThreshold-1)
	}
}

func reportInFeed(s *SQLiteStore, viewerID, reportID int64) bool {
	reports, err := s.ListFeedReports(100, viewerID)
	if err != nil {
		return false
	}
	for _, r := range reports {
		if r.ID == reportID {
			return true
		}
	}
	return false
}

func TestUndoPriceVote_noVote(t *testing.T) {
	s := openTestStore(t)

	ownerID, _ := s.CreateUser("owner3@test.com", "hash")
	voterID, _ := s.CreateUser("voter3@test.com", "hash2")
	pid, _ := s.CreateProduct("Arroz", "333", "Mercearia")
	markets, _ := s.ListSupermarkets()
	reportID, _ := s.CreatePriceReport(ownerID, pid, markets[0].ID, 900)

	_, _, err := s.UndoPriceVote(reportID, voterID)
	if !errors.Is(err, ErrNoVote) {
		t.Fatalf("expected ErrNoVote, got %v", err)
	}
}
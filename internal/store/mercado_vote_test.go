package store

import (
	"errors"
	"testing"
)

func TestVote_mutuallyExclusive_confirmAfterDispute(t *testing.T) {
	s := openTestStore(t)

	ownerID, _ := s.CreateUser("owner@test.com", "hash")
	voterID, _ := s.CreateUser("voter@test.com", "hash2")
	pid, _ := s.CreateProduct("Arroz", "111", "Mercearia")
	markets, _ := s.ListSupermarkets()
	reportID, _ := s.CreatePriceReport(ownerID, pid, markets[0].ID, 2000)

	if _, err := s.DisputePriceReport(reportID, voterID); err != nil {
		t.Fatal(err)
	}
	_, err := s.ConfirmPriceReport(reportID, voterID)
	if !errors.Is(err, ErrOppositeVote) {
		t.Fatalf("expected ErrOppositeVote, got %v", err)
	}
}

func TestVote_mutuallyExclusive_disputeAfterConfirm(t *testing.T) {
	s := openTestStore(t)

	ownerID, _ := s.CreateUser("owner2@test.com", "hash")
	voterID, _ := s.CreateUser("voter2@test.com", "hash2")
	pid, _ := s.CreateProduct("Feijão", "222", "Mercearia")
	markets, _ := s.ListSupermarkets()
	reportID, _ := s.CreatePriceReport(ownerID, pid, markets[0].ID, 800)

	if _, err := s.ConfirmPriceReport(reportID, voterID); err != nil {
		t.Fatal(err)
	}
	_, err := s.DisputePriceReport(reportID, voterID)
	if !errors.Is(err, ErrOppositeVote) {
		t.Fatalf("expected ErrOppositeVote, got %v", err)
	}
}

package store

import (
	"strings"
	"testing"
)

var saoPauloSeedMarkets = []string{
	"Pão de Açúcar Paulista",
	"Extra Penha",
	"Carrefour Tatuapé",
	"Assaí São Miguel",
}

const (
	spLatMin = -23.75
	spLatMax = -23.35
	spLngMin = -46.95
	spLngMax = -46.35
)

func TestSeedMercadoCatalog_saoPauloCapital(t *testing.T) {
	s := openTestStore(t)

	markets, err := s.ListSupermarkets()
	if err != nil {
		t.Fatal(err)
	}
	if len(markets) < len(saoPauloSeedMarkets) {
		t.Fatalf("markets len = %d, want at least %d", len(markets), len(saoPauloSeedMarkets))
	}

	names := make(map[string]bool, len(markets))
	for _, m := range markets {
		names[m.Name] = true
		if m.Lat < spLatMin || m.Lat > spLatMax || m.Lng < spLngMin || m.Lng > spLngMax {
			t.Errorf("market %q coords (%.4f, %.4f) outside São Paulo capital bounds", m.Name, m.Lat, m.Lng)
		}
	}
	for _, want := range saoPauloSeedMarkets {
		if !names[want] {
			t.Errorf("missing seeded supermarket %q", want)
		}
	}
}

func TestSeedMercadoCatalog_noCuritibaMarkets(t *testing.T) {
	s := openTestStore(t)

	markets, err := s.ListSupermarkets()
	if err != nil {
		t.Fatal(err)
	}
	curitibaHints := []string{"Curitiba", "Bacacheri", "Hauer", "Água Verde"}
	for _, m := range markets {
		for _, hint := range curitibaHints {
			if strings.Contains(m.Name, hint) {
				t.Errorf("unexpected Curitiba-era market %q (matched %q)", m.Name, hint)
			}
		}
	}
}

func TestGetOrCreateProfile_defaultCitySaoPaulo(t *testing.T) {
	s := openTestStore(t)

	userID, err := s.CreateUser("sp-user@test.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	profile, err := s.GetOrCreateProfile(userID)
	if err != nil {
		t.Fatal(err)
	}
	if profile.City != "São Paulo" {
		t.Fatalf("city = %q, want São Paulo", profile.City)
	}
}

func TestSeedMercadoDemo_noFakePriceReports(t *testing.T) {
	s, err := NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if err := s.SeedMercadoDemo(); err != nil {
		t.Fatal(err)
	}

	var count int
	if err := s.db.Raw().QueryRow(`SELECT COUNT(*) FROM price_reports`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("seed should not insert fake price reports, got %d", count)
	}
}

func TestSeedMercadoDemoFeedSample_demoCanVote(t *testing.T) {
	s := openTestStore(t)

	var demoID int64
	if err := s.db.Raw().QueryRow(`SELECT id FROM users WHERE email = ?`, "demo@example.com").Scan(&demoID); err != nil {
		t.Fatal(err)
	}

	reports, err := s.ListFeedReports(10, demoID)
	if err != nil {
		t.Fatal(err)
	}
	var votable bool
	for _, r := range reports {
		if r.UserID != demoID {
			votable = true
			if r.ProductName != "Leite Tirol Integro 1L" {
				t.Errorf("sample product = %q, want Leite Tirol", r.ProductName)
			}
			if r.Contributor != "Ana" {
				t.Errorf("contributor = %q, want Ana", r.Contributor)
			}
		}
	}
	if !votable {
		t.Fatal("demo feed should include a report from another user to test voting")
	}

	var anaExists int
	if err := s.db.Raw().QueryRow(`SELECT COUNT(*) FROM users WHERE email = ?`, "ana@example.com").Scan(&anaExists); err != nil {
		t.Fatal(err)
	}
	if anaExists != 1 {
		t.Fatalf("ana@example.com users = %d, want 1", anaExists)
	}
}

func TestSeedMercadoDemo_demoUserCitySaoPaulo(t *testing.T) {
	s := openTestStore(t)

	var demoUserID int64
	err := s.db.Raw().QueryRow(`SELECT id FROM users WHERE email = ?`, "demo@example.com").Scan(&demoUserID)
	if err != nil {
		t.Fatal(err)
	}
	profile, err := s.GetOrCreateProfile(demoUserID)
	if err != nil {
		t.Fatal(err)
	}
	if profile.City != "São Paulo" {
		t.Fatalf("demo city = %q, want São Paulo", profile.City)
	}
}

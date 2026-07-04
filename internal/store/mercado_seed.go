package store

import (
	"database/sql"
	"fmt"

	"github.com/puppe1990/cais/pkg/cais/session"
)

// SeedMercadoDemo inserts catalog data (supermarkets, badges, reference products).
// Does not seed fake price reports — users build real history by scanning.
func (s *SQLiteStore) SeedMercadoDemo() error {
	db := s.db.Raw()
	if err := seedMercadoCatalog(db); err != nil {
		return err
	}
	return seedMercadoDemoProfile(db)
}

func seedMercadoCatalog(db *sql.DB) error {
	products := []struct{ name, barcode, category string }{
		{"Leite Tirol Integro 1L", "7896256801011", "Laticínios"},
		{"Coca-Cola Original 2L", "7894900011517", "Bebidas"},
		{"Arroz Tio João 5kg", "7891081001015", "Mercearia"},
		{"Café Melitta 500g", "7891000000001", "Mercearia"},
	}
	for _, p := range products {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO products (name, barcode, category, source) VALUES (?, ?, ?, ?)`,
			p.name, p.barcode, p.category, "seed",
		); err != nil {
			return fmt.Errorf("seed product: %w", err)
		}
	}
	markets := []struct {
		name, address string
		lat, lng      float64
	}{
		{"Pão de Açúcar Paulista", "Av. Paulista, 2064", -23.5615, -46.6590},
		{"Extra Penha", "Av. Penha de França, 569", -23.5420, -46.5450},
		{"Carrefour Tatuapé", "R. Serra de Bragança, 629 — Tatuapé", -23.5405, -46.5755},
		{"Assaí São Miguel", "Av. Nordestina, 4944", -23.4945, -46.4440},
		{"Extra Belenzinho", "R. Oriente, 234 — Belenzinho", -23.5398, -46.5865},
		{"Assaí Belém", "Av. Alcântara Machado, 664 — Belém", -23.5368, -46.5832},
		{"Dia Belenzinho", "R. José Belenzinho, 289 — Belenzinho", -23.5442, -46.5948},
		{"Extra Tatuapé", "R. Serra de Bragança, 1555 — Tatuapé", -23.5492, -46.5418},
		{"Assaí Tatuapé", "R. Maria Cândida, 1899 — Tatuapé", -23.5496, -46.5392},
	}
	for _, m := range markets {
		if _, err := db.Exec(`
			INSERT INTO supermarkets (name, address, lat, lng)
			SELECT ?, ?, ?, ?
			WHERE NOT EXISTS (SELECT 1 FROM supermarkets WHERE name = ?)`,
			m.name, m.address, m.lat, m.lng, m.name,
		); err != nil {
			return fmt.Errorf("seed supermarket: %w", err)
		}
	}
	badges := []struct {
		slug, name, desc, icon string
		min                    int
	}{
		{"first_scan", "Primeiro Registro", "Registrou seu primeiro preço", "camera", 10},
		{"fiscal", "Fiscal do Povo", "10+ contribuições", "shield", 100},
		{"verifier", "Verificador Oficial", "Confirmou 3+ preços", "check", 6},
		{"nfce", "Auditor Digital", "Importou uma NFC-e", "qr", 500},
	}
	for _, b := range badges {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO badges (slug, name, description, icon, min_points) VALUES (?, ?, ?, ?, ?)`,
			b.slug, b.name, b.desc, b.icon, b.min,
		); err != nil {
			return fmt.Errorf("seed badge: %w", err)
		}
	}
	return nil
}

func seedMercadoDemoProfile(db *sql.DB) error {
	var demoUserID int64
	err := db.QueryRow(`SELECT id FROM users WHERE email = ?`, "demo@example.com").Scan(&demoUserID)
	if err != nil {
		return nil // auth seed may not exist yet
	}
	_, err = db.Exec(
		`INSERT OR IGNORE INTO user_profiles (user_id, display_name, points, city) VALUES (?, ?, 0, ?)`,
		demoUserID, "Demo", defaultCity,
	)
	return err
}

// SeedMercadoDemoFeedSample adds a community price report in development so demo
// can practice voting without a second account.
func (s *SQLiteStore) SeedMercadoDemoFeedSample() error {
	db := s.db.Raw()
	hash, err := session.HashPassword("password")
	if err != nil {
		return err
	}
	if _, err := db.Exec(
		`INSERT OR IGNORE INTO users (email, password_hash) VALUES (?, ?)`,
		"ana@example.com", hash,
	); err != nil {
		return fmt.Errorf("seed ana user: %w", err)
	}
	var anaID int64
	if err := db.QueryRow(`SELECT id FROM users WHERE email = ?`, "ana@example.com").Scan(&anaID); err != nil {
		return fmt.Errorf("find ana user: %w", err)
	}
	if _, err := db.Exec(
		`INSERT OR IGNORE INTO user_profiles (user_id, display_name, points, city) VALUES (?, ?, ?, ?)`,
		anaID, "Ana", 15, defaultCity,
	); err != nil {
		return fmt.Errorf("seed ana profile: %w", err)
	}
	var existing int
	if err := db.QueryRow(`SELECT COUNT(*) FROM price_reports WHERE user_id = ?`, anaID).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	var productID, marketID int64
	if err := db.QueryRow(`SELECT id FROM products WHERE barcode = ?`, "7896256801011").Scan(&productID); err != nil {
		return fmt.Errorf("sample product: %w", err)
	}
	if err := db.QueryRow(`SELECT id FROM supermarkets WHERE name = ?`, "Pão de Açúcar Paulista").Scan(&marketID); err != nil {
		return fmt.Errorf("sample supermarket: %w", err)
	}
	if _, err := s.CreatePriceReport(anaID, productID, marketID, 549); err != nil {
		return fmt.Errorf("sample report: %w", err)
	}
	return nil
}

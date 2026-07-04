package store

import (
	"database/sql"
	"fmt"
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
		{"Carrefour Tatuapé", "R. Serra de Bragança, 629", -23.5405, -46.5755},
		{"Assaí São Miguel", "Av. Nordestina, 4944", -23.4945, -46.4440},
	}
	for _, m := range markets {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO supermarkets (name, address, lat, lng) VALUES (?, ?, ?, ?)`,
			m.name, m.address, m.lat, m.lng,
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

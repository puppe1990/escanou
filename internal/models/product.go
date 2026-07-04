package models

import "time"

const (
	ProductSourceManual        = "manual"
	ProductSourceOpenFoodFacts = "openfoodfacts"
	ProductSourceSeed          = "seed"
)

type Product struct {
	ID           int64
	Name         string
	Barcode      string
	Category     string
	Brand        string
	Quantity     string
	ImageURL     string
	Source       string
	OffFetchedAt *time.Time
	CreatedAt    time.Time
}

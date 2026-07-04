package store

import (
	"testing"
	"time"

	"github.com/puppe1990/cais/pkg/cais/barcode"
)

func TestCreateProductFromOFF_persistsMetadata(t *testing.T) {
	s := openTestStore(t)

	off := barcode.Product{
		Name:     "Leite Integral 1L",
		Barcode:  "5901234123457",
		Category: "Laticínios",
		Brand:    "Tirol",
		Quantity: "1 L",
		ImageURL: "https://images.openfoodfacts.org/front.jpg",
	}
	before := time.Now().Add(-time.Second)
	p, err := s.CreateProductFromOFF(off)
	if err != nil {
		t.Fatal(err)
	}
	after := time.Now().Add(time.Second)

	if p.Source != "openfoodfacts" {
		t.Errorf("Source = %q, want openfoodfacts", p.Source)
	}
	if p.Brand != "Tirol" || p.Quantity != "1 L" || p.ImageURL == "" {
		t.Errorf("metadata not stored: %+v", p)
	}
	if p.OffFetchedAt == nil {
		t.Fatal("OffFetchedAt should be set")
	}
	if p.OffFetchedAt.Before(before) || p.OffFetchedAt.After(after) {
		t.Errorf("OffFetchedAt = %v, want between %v and %v", p.OffFetchedAt, before, after)
	}

	got, found, err := s.FindProductByBarcode("5901234123457")
	if err != nil || !found {
		t.Fatalf("FindProductByBarcode: found=%v err=%v", found, err)
	}
	if got.Brand != "Tirol" || got.ImageURL != off.ImageURL || got.Source != "openfoodfacts" {
		t.Errorf("reload mismatch: %+v", got)
	}
}
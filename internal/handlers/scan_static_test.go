package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanJS_configuresEANBarcodeScanning(t *testing.T) {
	t.Helper()
	root := filepath.Join("..", "..")
	data, err := os.ReadFile(filepath.Join(root, "web", "static", "js", "scan.js"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, needle := range []string{
		"EAN_13",
		"formatsToSupport",
		"useBarCodeDetectorIfSupported",
		"facingMode: \"environment\"",
		"qrbox",
	} {
		if !strings.Contains(body, needle) {
			t.Errorf("scan.js missing %q — barcode scan may fail on mobile", needle)
		}
	}
}
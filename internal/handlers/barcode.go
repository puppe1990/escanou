package handlers

import (
	"strings"
	"unicode"
)

func normalizeBarcode(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func validBarcode(raw string) bool {
	code := normalizeBarcode(raw)
	if len(code) == 12 {
		code = "0" + code
	}
	if len(code) != 8 && len(code) != 13 {
		return false
	}
	return eanCheckDigit(code[:len(code)-1]) == code[len(code)-1]
}

func eanCheckDigit(withoutCheck string) byte {
	sum := 0
	for i := 0; i < len(withoutCheck); i++ {
		d := int(withoutCheck[i] - '0')
		if i%2 == 0 {
			sum += d
		} else {
			sum += d * 3
		}
	}
	return byte('0' + ((10 - (sum % 10)) % 10))
}

package utils

import (
	"fmt"
	"math"
	"strings"
)

// ParseFloat converts a string to float64, returning 0 on error.
func ParseFloat(value string) float64 {
	var whole, frac float64
	var fracDiv float64 = 1
	seenDot := false
	for _, ch := range strings.TrimSpace(value) {
		switch {
		case ch == '.':
			if seenDot {
				return 0
			}
			seenDot = true
		case ch >= '0' && ch <= '9':
			digit := float64(ch - '0')
			if !seenDot {
				whole = whole*10 + digit
			} else {
				fracDiv *= 10
				frac += digit / fracDiv
			}
		default:
			return 0
		}
	}
	return whole + frac
}

// FormatQuantity formats a float64 quantity to a string with appropriate precision for the symbol,
// respecting exchange-specific lot size filters.
func FormatQuantity(symbol string, qty float64) string {
	precision := 4 // default fallback to 4 decimals (e.g. ETHUSDT)
	upperSymbol := strings.ToUpper(symbol)

	if strings.Contains(upperSymbol, "BTC") {
		precision = 5
	} else if strings.Contains(upperSymbol, "ETH") {
		precision = 4
	} else if strings.Contains(upperSymbol, "BNB") {
		precision = 3
	} else if strings.Contains(upperSymbol, "SOL") {
		precision = 3
	} else if strings.Contains(upperSymbol, "WIF") {
		precision = 2
	} else if strings.Contains(upperSymbol, "PEPE") {
		precision = 0
	} else if strings.Contains(upperSymbol, "XRP") {
		precision = 1
	} else if strings.Contains(upperSymbol, "ADA") {
		precision = 1
	} else if strings.Contains(upperSymbol, "DOGE") {
		precision = 0
	} else {
		// General fallback based on quantity range if unknown
		if qty >= 1 {
			precision = 4
		} else if qty >= 0.01 {
			precision = 6
		} else {
			precision = 8
		}
	}

	pow := math.Pow(10, float64(precision))
	// Add a tiny epsilon (1e-9) to prevent float64 representation precision loss from truncating down.
	truncatedQty := math.Floor(qty*pow+1e-9) / pow

	return fmt.Sprintf("%.*f", precision, truncatedQty)
}

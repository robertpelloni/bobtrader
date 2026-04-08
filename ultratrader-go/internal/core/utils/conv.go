package utils

import "strings"

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

package analytics

import (
	"fmt"
	"math"
)

// PatternType defines the type of chart pattern detected.
type PatternType string

const (
	DoubleTop     PatternType = "DOUBLE_TOP"
	DoubleBottom  PatternType = "DOUBLE_BOTTOM"
	HeadShoulders PatternType = "HEAD_AND_SHOULDERS"
)

// PatternResult contains the details of a detected pattern.
type PatternResult struct {
	Type       PatternType
	Confidence float64 // 0.0 to 1.0 based on how clean the pattern looks
	EndIndex   int     // the index in the prices slice where the pattern concludes
}

// PatternRecognizer scans raw price arrays for market anomalies.
type PatternRecognizer struct {
	tolerancePct float64 // Tolerance for matching peak heights (e.g. 0.02 for 2%)
}

// NewPatternRecognizer creates a new recognizer.
func NewPatternRecognizer(tolerancePct float64) *PatternRecognizer {
	if tolerancePct <= 0 {
		tolerancePct = 0.02 // default 2%
	}
	return &PatternRecognizer{
		tolerancePct: tolerancePct,
	}
}

// Scan processes a slice of close prices to detect all known patterns.
func (p *PatternRecognizer) Scan(prices []float64) []PatternResult {
	var results []PatternResult

	if len(prices) < 20 {
		return results
	}

	results = append(results, p.detectDoubleTopsBottoms(prices)...)
	results = append(results, p.detectHeadShoulders(prices)...)

	return results
}

func (p *PatternRecognizer) isPeak(prices []float64, i int, window int) bool {
	if i-window < 0 || i+window >= len(prices) {
		return false
	}
	val := prices[i]
	for j := i - window; j <= i+window; j++ {
		if prices[j] > val {
			return false
		}
	}
	return true
}

func (p *PatternRecognizer) isTrough(prices []float64, i int, window int) bool {
	if i-window < 0 || i+window >= len(prices) {
		return false
	}
	val := prices[i]
	for j := i - window; j <= i+window; j++ {
		if prices[j] < val {
			return false
		}
	}
	return true
}

// findExtremes finds local maxima and minima using a sliding window.
func (p *PatternRecognizer) findExtremes(prices []float64, window int) (peaks []int, troughs []int) {
	for i := window; i < len(prices)-window; i++ {
		if p.isPeak(prices, i, window) {
			peaks = append(peaks, i)
		}
		if p.isTrough(prices, i, window) {
			troughs = append(troughs, i)
		}
	}
	return
}

func (p *PatternRecognizer) detectDoubleTopsBottoms(prices []float64) []PatternResult {
	var results []PatternResult
	peaks, troughs := p.findExtremes(prices, 5)

	// Double Top
	for i := 0; i < len(peaks)-1; i++ {
		p1, p2 := peaks[i], peaks[i+1]

		// Ensure peaks aren't too close or too far
		if p2-p1 < 5 || p2-p1 > 50 {
			continue
		}

		v1, v2 := prices[p1], prices[p2]
		diff := math.Abs(v1-v2) / math.Max(v1, v2)

		if diff <= p.tolerancePct {
			// Find lowest point between the two peaks
			valley := v1
			for j := p1; j <= p2; j++ {
				if prices[j] < valley {
					valley = prices[j]
				}
			}

			// Ensure there was a significant dip between peaks
			drop := (math.Max(v1, v2) - valley) / math.Max(v1, v2)
			if drop > p.tolerancePct*2 {
				// Calculate confidence: cleaner = higher
				confidence := 1.0 - (diff / p.tolerancePct)
				if confidence < 0 {
					confidence = 0.1
				}

				results = append(results, PatternResult{
					Type:       DoubleTop,
					Confidence: confidence,
					EndIndex:   p2,
				})
			}
		}
	}

	// Double Bottom (inverse)
	for i := 0; i < len(troughs)-1; i++ {
		t1, t2 := troughs[i], troughs[i+1]

		if t2-t1 < 5 || t2-t1 > 50 {
			continue
		}

		v1, v2 := prices[t1], prices[t2]
		diff := math.Abs(v1-v2) / math.Max(v1, v2)

		if diff <= p.tolerancePct {
			// Find highest point between troughs
			peak := v1
			for j := t1; j <= t2; j++ {
				if prices[j] > peak {
					peak = prices[j]
				}
			}

			rise := (peak - math.Min(v1, v2)) / math.Min(v1, v2)
			if rise > p.tolerancePct*2 {
				confidence := 1.0 - (diff / p.tolerancePct)
				if confidence < 0 {
					confidence = 0.1
				}

				results = append(results, PatternResult{
					Type:       DoubleBottom,
					Confidence: confidence,
					EndIndex:   t2,
				})
			}
		}
	}

	return results
}

func (p *PatternRecognizer) detectHeadShoulders(prices []float64) []PatternResult {
	var results []PatternResult
	peaks, _ := p.findExtremes(prices, 5)

	if len(peaks) < 3 {
		return results
	}

	// Look for triplet of peaks: Left Shoulder, Head, Right Shoulder
	for i := 0; i < len(peaks)-2; i++ {
		p1, p2, p3 := peaks[i], peaks[i+1], peaks[i+2]
		v1, v2, v3 := prices[p1], prices[p2], prices[p3]

		// Head must be significantly higher than shoulders
		if v2 > v1*(1+p.tolerancePct) && v2 > v3*(1+p.tolerancePct) {
			// Shoulders must be somewhat equal
			diff := math.Abs(v1-v3) / math.Max(v1, v3)
			if diff <= p.tolerancePct {
				confidence := 1.0 - (diff / p.tolerancePct)

				results = append(results, PatternResult{
					Type:       HeadShoulders,
					Confidence: confidence,
					EndIndex:   p3,
				})
			}
		}
	}

	return results
}

func Example() {
	fmt.Println("Pattern recognizer loaded")
}

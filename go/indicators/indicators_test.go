package indicators

import (
	"math"
	"testing"
)

func TestSMA(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Test SMA with period 3
	result := SMA(values, 3)

	// First two values should be 0 (not enough data)
	if result[0] != 0 || result[1] != 0 {
		t.Errorf("Expected first two values to be 0, got %f, %f", result[0], result[1])
	}

	// SMA at index 2: (1+2+3)/3 = 2
	expected := 2.0
	if math.Abs(result[2]-expected) > 0.0001 {
		t.Errorf("Expected SMA[2] = %f, got %f", expected, result[2])
	}

	// SMA at index 9: (8+9+10)/3 = 9
	expected = 9.0
	if math.Abs(result[9]-expected) > 0.0001 {
		t.Errorf("Expected SMA[9] = %f, got %f", expected, result[9])
	}
}

func TestSMAEdgeCases(t *testing.T) {
	// Empty input
	result := SMA([]float64{}, 5)
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input")
	}

	// Period larger than data
	values := []float64{1, 2, 3}
	result = SMA(values, 10)
	if len(result) != 3 {
		t.Errorf("Expected result length 3, got %d", len(result))
	}

	// Period <= 0
	result = SMA(values, 0)
	if len(result) != 0 {
		t.Errorf("Expected empty result for period <= 0")
	}
}

func TestEMA(t *testing.T) {
	values := []float64{10, 11, 12, 13, 14}

	result := EMA(values, 3)

	// First EMA value should equal first price
	if math.Abs(result[0]-10) > 0.0001 {
		t.Errorf("Expected EMA[0] = 10, got %f", result[0])
	}

	// Verify all values are calculated
	for i := 0; i < len(result); i++ {
		if result[i] == 0 && i > 0 {
			t.Errorf("Unexpected zero value at index %d", i)
		}
	}
}

func TestRSI(t *testing.T) {
	// Constant values should give RSI of 50 (equal gains and losses)
	values := []float64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}

	result := RSI(values, 7)

	// First period values should be 0
	for i := 0; i < 7; i++ {
		if result[i] != 0 {
			t.Errorf("Expected RSI[%d] = 0, got %f", i, result[i])
		}
	}

	// With constant prices, RSI should be around 50 after initial period
	// (actually exactly 50 if no changes)
}

func TestRSIRisingPrices(t *testing.T) {
	// Continuously rising prices should give high RSI
	values := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	result := RSI(values, 7)

	// After period, RSI should be high (> 50) for rising prices
	if result[7] < 50 {
		t.Errorf("Expected RSI > 50 for rising prices, got %f", result[7])
	}
}

func TestRSIFallingPrices(t *testing.T) {
	// Continuously falling prices should give low RSI
	values := []float64{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10}

	result := RSI(values, 7)

	// After period, RSI should be low (< 50) for falling prices
	if result[7] > 50 {
		t.Errorf("Expected RSI < 50 for falling prices, got %f", result[7])
	}
}

func TestRSIEdgeCases(t *testing.T) {
	// Empty input
	result := RSI([]float64{}, 5)
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input")
	}

	// Not enough data
	values := []float64{1, 2}
	result = RSI(values, 10)
	if len(result) != 2 {
		t.Errorf("Expected result length 2, got %d", len(result))
	}
}

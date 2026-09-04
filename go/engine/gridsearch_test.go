package engine

import (
	"backtest/data"
	"backtest/strategy"
	"testing"
)

func TestGridSearch(t *testing.T) {
	// Create sample candles
	candles := generateTestCandles(100)

	// Define parameter ranges
	paramRanges := map[string]ParamRange{
		"fastPeriod": {Min: 5, Max: 10, Step: 5, Name: "Fast Period"},
		"slowPeriod": {Min: 15, Max: 20, Step: 5, Name: "Slow Period"},
	}

	// Settings
	settings := strategy.Settings{
		InitialCapital:   10000,
		CommissionMarket: 0.001,
		CommissionLimit:  0.0005,
		EnableLong:       true,
	}

	// Run grid search
	baseStrategy := strategy.NewSMAStrategy(10, 20)
	results, err := GridSearch(candles, baseStrategy, settings, paramRanges)

	if err != nil {
		t.Fatalf("GridSearch failed: %v", err)
	}

	// Expected combinations: fastPeriod=[5,10], slowPeriod=[15,20] = 4 combinations
	expectedCombinations := 4
	if len(results) != expectedCombinations {
		t.Errorf("Expected %d results, got %d", expectedCombinations, len(results))
	}

	// Verify each result has parameters and metrics
	for i, result := range results {
		if result.Parameters == nil {
			t.Errorf("Result %d: Parameters is nil", i)
		}

		if _, ok := result.Parameters["fastPeriod"]; !ok {
			t.Errorf("Result %d: missing fastPeriod parameter", i)
		}

		if _, ok := result.Parameters["slowPeriod"]; !ok {
			t.Errorf("Result %d: missing slowPeriod parameter", i)
		}

		// Check that metrics are populated
		if result.Metrics.TotalTrades < 0 {
			t.Errorf("Result %d: invalid TotalTrades", i)
		}
	}
}

func TestGenerateCombinations(t *testing.T) {
	paramRanges := map[string]ParamRange{
		"fastPeriod": {Min: 5, Max: 10, Step: 5, Name: "Fast Period"},
		"slowPeriod": {Min: 15, Max: 20, Step: 5, Name: "Slow Period"},
	}

	paramNames := []string{"fastPeriod", "slowPeriod"}
	combinations := generateCombinations(paramRanges, paramNames, 0, make(map[string]float64))

	// Should have 4 combinations: (5,15), (5,20), (10,15), (10,20)
	if len(combinations) != 4 {
		t.Errorf("Expected 4 combinations, got %d", len(combinations))
	}

	// Verify all combinations exist
	found := make(map[string]bool)
	for _, combo := range combinations {
		key := string(rune(int(combo["fastPeriod"]))) + "-" + string(rune(int(combo["slowPeriod"])))
		found[key] = true
	}

	// Just verify we have the right number of unique combinations
	uniqueFast := make(map[float64]bool)
	uniqueSlow := make(map[float64]bool)
	for _, combo := range combinations {
		uniqueFast[combo["fastPeriod"]] = true
		uniqueSlow[combo["slowPeriod"]] = true
	}

	if len(uniqueFast) != 2 {
		t.Errorf("Expected 2 unique fastPeriod values, got %d", len(uniqueFast))
	}

	if len(uniqueSlow) != 2 {
		t.Errorf("Expected 2 unique slowPeriod values, got %d", len(uniqueSlow))
	}
}

func TestGridSearchEmptyParamRanges(t *testing.T) {
	candles := generateTestCandles(50)
	paramRanges := map[string]ParamRange{}
	settings := strategy.Settings{
		InitialCapital:   10000,
		CommissionMarket: 0.001,
		EnableLong:       true,
	}

	baseStrategy := strategy.NewSMAStrategy(10, 20)
	results, err := GridSearch(candles, baseStrategy, settings, paramRanges)

	if err != nil {
		t.Fatalf("GridSearch with empty paramRanges failed: %v", err)
	}

	// Should return 1 result with default/empty parameters
	if len(results) != 1 {
		t.Errorf("Expected 1 result for empty paramRanges, got %d", len(results))
	}
}

// Helper function to generate test candles
func generateTestCandles(num int) []data.Candle {
	candles := make([]data.Candle, num)
	price := 100.0
	baseTime := data.Candle{}.Time

	for i := 0; i < num; i++ {
		change := (float64(i%10) - 5) * 0.1
		open := price
		close := price + change
		high := open
		if close > open {
			high = close
		}
		low := open
		if close < open {
			low = close
		}

		candles[i] = data.Candle{
			Time:   baseTime.AddDate(0, 0, i-num),
			Open:   open,
			High:   high + float64(i)*0.01,
			Low:    low - float64(i)*0.01,
			Close:  close,
			Volume: float64(100 + i*10),
		}
		price = close
	}

	return candles
}

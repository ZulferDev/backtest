package strategy

import (
	"backtest/data"
	"testing"
)

func TestSMAStrategyShouldBuy(t *testing.T) {
	strategy := NewSMAStrategy(5, 10)
	
	// Create test SMA data with a crossover
	// Index 4: fast=10, slow=0 (not enough data for slow)
	// Index 5: fast=11, slow=12 (fast < slow, no buy)
	// Index 6: fast=12, slow=12 (crossover happens here - fast was below, now equal/crossing)
	fastSMA := []float64{0, 0, 0, 0, 10, 11, 13, 14, 15, 16}
	slowSMA := []float64{0, 0, 0, 0, 0, 12, 12, 12, 12, 12}
	
	// At index 6, fast (13) crosses above slow (12) - prev: 11<12, curr: 13>12
	if !strategy.ShouldBuy(6, fastSMA, slowSMA) {
		t.Error("Expected ShouldBuy to return true at crossover")
	}
	
	// At index 5, no crossover yet (fast 11 < slow 12)
	if strategy.ShouldBuy(5, fastSMA, slowSMA) {
		t.Error("Expected ShouldBuy to return false before crossover")
	}
}

func TestSMAStrategyShouldSell(t *testing.T) {
	strategy := NewSMAStrategy(5, 10)
	
	// Create test SMA data with a crossunder
	// Index 5: fast=14, slow=12 (fast > slow)
	// Index 6: fast=13, slow=12 (fast > slow, still)
	// Index 7: fast=11, slow=12 (crossunder - prev: 13>12, curr: 11<12)
	fastSMA := []float64{0, 0, 0, 0, 15, 14, 13, 11, 10, 9}
	slowSMA := []float64{0, 0, 0, 0, 0, 12, 12, 12, 12, 12}
	
	// At index 7, fast (11) crosses below slow (12)
	if !strategy.ShouldSell(7, fastSMA, slowSMA) {
		t.Error("Expected ShouldSell to return true at crossunder")
	}
	
	// At index 6, no crossunder yet (fast 13 > slow 12)
	if strategy.ShouldSell(6, fastSMA, slowSMA) {
		t.Error("Expected ShouldSell to return false before crossunder")
	}
}

func TestSMAStrategyCalculateIndicators(t *testing.T) {
	strategy := NewSMAStrategy(3, 5)
	
	candles := []data.Candle{
		{Close: 1}, {Close: 2}, {Close: 3}, {Close: 4}, {Close: 5},
		{Close: 6}, {Close: 7}, {Close: 8}, {Close: 9}, {Close: 10},
	}
	
	fastSMA, slowSMA := strategy.CalculateIndicators(candles)
	
	if len(fastSMA) != 10 {
		t.Errorf("Expected fastSMA length 10, got %d", len(fastSMA))
	}
	
	if len(slowSMA) != 10 {
		t.Errorf("Expected slowSMA length 10, got %d", len(slowSMA))
	}
	
	// Check first valid SMA values
	// Fast SMA (period 3) at index 2: (1+2+3)/3 = 2
	if fastSMA[2] != 2.0 {
		t.Errorf("Expected fastSMA[2] = 2.0, got %f", fastSMA[2])
	}
	
	// Slow SMA (period 5) at index 4: (1+2+3+4+5)/5 = 3
	if slowSMA[4] != 3.0 {
		t.Errorf("Expected slowSMA[4] = 3.0, got %f", slowSMA[4])
	}
}

func TestSMAStrategyGetParameters(t *testing.T) {
	strategy := NewSMAStrategy(10, 20)
	
	params := strategy.GetParameters()
	
	if params["fastPeriod"] != 10 {
		t.Errorf("Expected fastPeriod 10, got %v", params["fastPeriod"])
	}
	
	if params["slowPeriod"] != 20 {
		t.Errorf("Expected slowPeriod 20, got %v", params["slowPeriod"])
	}
}

func TestSMAStrategyGetName(t *testing.T) {
	strategy := NewSMAStrategy(10, 20)
	
	name := strategy.GetName()
	
	if name != "SMA Crossover" {
		t.Errorf("Expected name 'SMA Crossover', got '%s'", name)
	}
}

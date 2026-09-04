package strategy

import (
	"backtest/data"
	"backtest/indicators"
)

// SMAStrategy implements a Simple Moving Average Crossover strategy
type SMAStrategy struct {
	FastPeriod int `json:"fastPeriod"`
	SlowPeriod int `json:"slowPeriod"`
	fastSMA    []float64
	slowSMA    []float64
}

// NewSMAStrategy creates a new SMA Crossover strategy
func NewSMAStrategy(fastPeriod, slowPeriod int) *SMAStrategy {
	return &SMAStrategy{
		FastPeriod: fastPeriod,
		SlowPeriod: slowPeriod,
	}
}

// GetName returns the strategy name
func (s *SMAStrategy) GetName() string {
	return "SMA Crossover"
}

// GetParameters returns current parameter values
func (s *SMAStrategy) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"fastPeriod": s.FastPeriod,
		"slowPeriod": s.SlowPeriod,
	}
}

// OnBar is called for each candle during backtest
func (s *SMAStrategy) OnBar(bar data.Candle, index int, ctx Context) {
	// Calculate indicators on demand (could be pre-calculated for efficiency)
	closePrices := make([]float64, index+1)
	// In a real implementation, we'd have access to historical close prices
	// For now, we'll use the current bar's close as a placeholder
	// The actual implementation in engine will pre-calculate all indicators

	_ = closePrices
	_ = bar

	// Placeholder - actual logic implemented in engine with pre-calculated indicators
	// Entry: Fast SMA crosses above Slow SMA
	// Exit: Fast SMA crosses below Slow SMA
}

// CalculateIndicators pre-calculates indicators for the strategy
func (s *SMAStrategy) CalculateIndicators(candles []data.Candle) (fastSMA, slowSMA []float64) {
	closePrices := make([]float64, len(candles))
	for i, c := range candles {
		closePrices[i] = c.Close
	}

	fastSMA = indicators.SMA(closePrices, s.FastPeriod)
	slowSMA = indicators.SMA(closePrices, s.SlowPeriod)

	return fastSMA, slowSMA
}

// ShouldBuy checks if the strategy should enter a long position
func (s *SMAStrategy) ShouldBuy(index int, fastSMA, slowSMA []float64) bool {
	if index < 2 || index >= len(fastSMA) || index >= len(slowSMA) {
		return false
	}

	// Check for crossover: fast was below slow, now above
	prevFast := fastSMA[index-1]
	prevSlow := slowSMA[index-1]
	currFast := fastSMA[index]
	currSlow := slowSMA[index]

	// Handle zero values (not enough data)
	if prevFast == 0 || prevSlow == 0 || currFast == 0 || currSlow == 0 {
		return false
	}

	return prevFast <= prevSlow && currFast > currSlow
}

// ShouldSell checks if the strategy should exit a long position
func (s *SMAStrategy) ShouldSell(index int, fastSMA, slowSMA []float64) bool {
	if index < 2 || index >= len(fastSMA) || index >= len(slowSMA) {
		return false
	}

	// Check for crossunder: fast was above slow, now below
	prevFast := fastSMA[index-1]
	prevSlow := slowSMA[index-1]
	currFast := fastSMA[index]
	currSlow := slowSMA[index]

	// Handle zero values (not enough data)
	if prevFast == 0 || prevSlow == 0 || currFast == 0 || currSlow == 0 {
		return false
	}

	return prevFast >= prevSlow && currFast < currSlow
}

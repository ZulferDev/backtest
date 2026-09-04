package engine

import (
	"backtest/data"
	"backtest/strategy"
)

// ParamRange represents a range for parameter optimization
type ParamRange struct {
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
	Step float64 `json:"step"`
	Name string  `json:"name"`
}

// GridResult represents the result of a single grid search iteration
type GridResult struct {
	Parameters map[string]float64 `json:"parameters"`
	Metrics    strategy.Metrics   `json:"metrics"`
}

// GridSearch runs backtest for all combinations of parameters
func GridSearch(candles []data.Candle, baseStrategy *strategy.SMAStrategy, settings strategy.Settings, paramRanges map[string]ParamRange) ([]GridResult, error) {
	var results []GridResult

	// Get parameter names
	paramNames := make([]string, 0, len(paramRanges))
	for name := range paramRanges {
		paramNames = append(paramNames, name)
	}

	// Generate all combinations
	combinations := generateCombinations(paramRanges, paramNames, 0, make(map[string]float64))

	// Run backtest for each combination
	for _, params := range combinations {
		// Create strategy with current parameters
		fastPeriod := int(params["fastPeriod"])
		slowPeriod := int(params["slowPeriod"])

		strat := strategy.NewSMAStrategy(fastPeriod, slowPeriod)

		// Create engine
		eng := NewEngine(settings)

		// Pre-calculate indicators
		fastSMA, slowSMA := strat.CalculateIndicators(candles)

		// Wrap strategy
		wrapped := &smaWrapper{
			strategy: strat,
			fastSMA:  fastSMA,
			slowSMA:  slowSMA,
		}

		// Run backtest
		result, err := eng.Run(candles, wrapped)
		if err != nil {
			continue // Skip failed backtests
		}

		gridResult := GridResult{
			Parameters: params,
			Metrics:    result.Metrics,
		}

		results = append(results, gridResult)
	}

	return results, nil
}

// generateCombinations generates all parameter combinations recursively
func generateCombinations(paramRanges map[string]ParamRange, paramNames []string, index int, current map[string]float64) []map[string]float64 {
	if index >= len(paramNames) {
		// Base case: copy current map
		result := make(map[string]float64)
		for k, v := range current {
			result[k] = v
		}
		return []map[string]float64{result}
	}

	name := paramNames[index]
	paramRange := paramRanges[name]

	var allCombinations []map[string]float64

	// Iterate through all values for this parameter
	for value := paramRange.Min; value <= paramRange.Max+0.0001; value += paramRange.Step {
		current[name] = value
		combinations := generateCombinations(paramRanges, paramNames, index+1, current)
		allCombinations = append(allCombinations, combinations...)
	}

	return allCombinations
}

// smaWrapper is defined in wasm/main.go, but we need it here too for engine package
type smaWrapper struct {
	strategy *strategy.SMAStrategy
	fastSMA  []float64
	slowSMA  []float64
}

func (w *smaWrapper) GetName() string {
	return w.strategy.GetName()
}

func (w *smaWrapper) GetParameters() map[string]interface{} {
	return w.strategy.GetParameters()
}

func (w *smaWrapper) OnBar(bar data.Candle, index int, ctx strategy.Context) {
	if w.strategy.ShouldBuy(index, w.fastSMA, w.slowSMA) {
		if ctx.GetPosition() == 0 {
			quantity := (ctx.GetCash() * 0.95) / bar.Close
			ctx.Buy(quantity)
		}
	} else if w.strategy.ShouldSell(index, w.fastSMA, w.slowSMA) {
		if ctx.GetPosition() > 0 {
			ctx.ClosePosition()
		}
	}
}

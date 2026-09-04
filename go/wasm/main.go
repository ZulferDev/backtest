package main

import (
	"encoding/json"
	"syscall/js"

	"backtest/data"
	"backtest/engine"
	"backtest/strategy"
)

// BacktestSettings represents the settings sent from JavaScript
type BacktestSettings struct {
	InitialCapital   float64 `json:"initialCapital"`
	CommissionMarket float64 `json:"commissionMarket"`
	CommissionLimit  float64 `json:"commissionLimit"`
	Slippage         float64 `json:"slippage"`
	FastPeriod       int     `json:"fastPeriod"`
	SlowPeriod       int     `json:"slowPeriod"`
}

func main() {
	// Register Go functions to be called from JavaScript
	js.Global().Set("goParseCSV", js.FuncOf(parseCSV))
	js.Global().Set("goParseJSON", js.FuncOf(parseJSON))
	js.Global().Set("goDetectGaps", js.FuncOf(detectGaps))
	js.Global().Set("goRunBacktest", js.FuncOf(runBacktest))

	// Keep the Go program running
	select {}
}

// parseCSV parses CSV data and returns JSON array of candles
func parseCSV(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return jsonError("Missing arguments: data and delimiter")
	}

	csvData := args[0].String()
	delimiter := rune(args[1].Int())

	candles, err := data.ParseCSV(csvData, delimiter)
	if err != nil {
		return jsonError(err.Error())
	}

	jsonData, err := json.Marshal(candles)
	if err != nil {
		return jsonError(err.Error())
	}

	return string(jsonData)
}

// parseJSON parses JSON data and returns JSON array of candles
func parseJSON(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return jsonError("Missing argument: data")
	}

	jsonData := args[0].String()

	candles, err := data.ParseJSON(jsonData)
	if err != nil {
		return jsonError(err.Error())
	}

	result, err := json.Marshal(candles)
	if err != nil {
		return jsonError(err.Error())
	}

	return string(result)
}

// detectGaps detects gaps in candle data
func detectGaps(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return jsonError("Missing arguments: candlesJSON and intervalMs")
	}

	candlesJSON := args[0].String()
	intervalMs := args[1].Int()

	var candles []data.Candle
	if err := json.Unmarshal([]byte(candlesJSON), &candles); err != nil {
		return jsonError(err.Error())
	}

	gaps, err := data.DetectGaps(candles, intervalMs)
	if err != nil {
		return jsonError(err.Error())
	}

	result, err := json.Marshal(gaps)
	if err != nil {
		return jsonError(err.Error())
	}

	return string(result)
}

// runBacktest runs a backtest with the given settings
func runBacktest(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return jsonError("Missing arguments: candlesJSON, strategyJSON, settingsJSON")
	}

	candlesJSON := args[0].String()
	_ = args[1].String() // strategyJSON - for future use with different strategies
	settingsJSON := args[2].String()

	// Parse candles
	var candles []data.Candle
	if err := json.Unmarshal([]byte(candlesJSON), &candles); err != nil {
		return jsonError("Failed to parse candles: " + err.Error())
	}

	// Parse settings
	var settings BacktestSettings
	if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
		return jsonError("Failed to parse settings: " + err.Error())
	}

	// Create strategy
	strat := strategy.NewSMAStrategy(settings.FastPeriod, settings.SlowPeriod)

	// Pre-calculate indicators
	fastSMA, slowSMA := strat.CalculateIndicators(candles)

	// Create engine with settings
	engineSettings := strategy.Settings{
		InitialCapital:   settings.InitialCapital,
		CommissionMarket: settings.CommissionMarket,
		CommissionLimit:  settings.CommissionLimit,
		Slippage:         settings.Slippage,
		EnableLong:       true,
		EnableShort:      false,
	}

	eng := engine.NewEngine(engineSettings)

	// Override OnBar to use pre-calculated indicators
	wrappedStrategy := &smaWrapper{
		strategy: strat,
		fastSMA:  fastSMA,
		slowSMA:  slowSMA,
	}

	result, err := eng.Run(candles, wrappedStrategy)
	if err != nil {
		return jsonError("Backtest failed: " + err.Error())
	}

	// Convert result to JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return jsonError("Failed to serialize result: " + err.Error())
	}

	return string(resultJSON)
}

// smaWrapper wraps SMAStrategy to use pre-calculated indicators
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
	// Use pre-calculated indicators
	if w.strategy.ShouldBuy(index, w.fastSMA, w.slowSMA) {
		// Enter long position
		if ctx.GetPosition() == 0 {
			// Calculate quantity based on available cash and current price
			quantity := (ctx.GetCash() * 0.95) / bar.Close // Use 95% of cash
			ctx.Buy(quantity)
		}
	} else if w.strategy.ShouldSell(index, w.fastSMA, w.slowSMA) {
		// Exit position
		if ctx.GetPosition() > 0 {
			ctx.ClosePosition()
		}
	}
}

func jsonError(msg string) string {
	errorMap := map[string]interface{}{
		"error": msg,
	}
	result, _ := json.Marshal(errorMap)
	return string(result)
}

// Helper function to format errors for JavaScript
func formatError(err error) string {
	errorMap := map[string]string{
		"error": err.Error(),
	}
	result, _ := json.Marshal(errorMap)
	return string(result)
}

// Unused but kept for potential future use
var _ = formatError

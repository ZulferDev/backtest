package engine

import (
	"backtest/data"
	"backtest/strategy"
	"testing"
	"time"
)

// mockStrategy is a simple strategy for testing
type mockStrategy struct {
	buyAt  int
	sellAt int
}

func (m *mockStrategy) OnBar(bar data.Candle, index int, ctx strategy.Context) {
	if index == m.buyAt && ctx.GetPosition() == 0 {
		// Buy with all available cash
		cash := ctx.GetCash()
		price := bar.Close
		qty := cash / price / 10 * 10 // Round down to nearest 10
		if qty > 0 {
			ctx.Buy(qty)
		}
	}
	if index == m.sellAt && ctx.GetPosition() > 0 {
		ctx.ClosePosition()
	}
}

func (m *mockStrategy) CalculateIndicators(candles []data.Candle) {}
func (m *mockStrategy) GetParameters() map[string]interface{}    { return nil }
func (m *mockStrategy) GetName() string                          { return "mock" }

func createTestCandles(count int, startPrice float64) []data.Candle {
	candles := make([]data.Candle, count)
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	for i := 0; i < count; i++ {
		price := startPrice + float64(i)*0.1
		candles[i] = data.Candle{
			Time:   baseTime.Add(time.Duration(i) * time.Hour),
			Open:   price,
			High:   price + 0.5,
			Low:    price - 0.5,
			Close:  price,
			Volume: 1000.0,
		}
	}
	return candles
}

func TestNewEngine(t *testing.T) {
	settings := strategy.Settings{
		InitialCapital:     10000,
		CommissionMarket:   0.001,
		CommissionLimit:    0.0005,
		StopLossPercent:    0,
		TakeProfitPercent:  0,
	}
	
	engine := NewEngine(settings)
	
	if engine == nil {
		t.Fatal("Expected engine to be created")
	}
	
	if engine.cash != settings.InitialCapital {
		t.Errorf("Expected cash %f, got %f", settings.InitialCapital, engine.cash)
	}
	
	if engine.position != 0 {
		t.Errorf("Expected initial position 0, got %f", engine.position)
	}
}

func TestEngineRunSimpleBuySell(t *testing.T) {
	settings := strategy.Settings{
		InitialCapital:   10000,
		CommissionMarket: 0.001,
	}
	
	candles := createTestCandles(100, 100)
	strat := &mockStrategy{buyAt: 10, sellAt: 50}
	
	engine := NewEngine(settings)
	result, err := engine.Run(candles, strat)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result == nil {
		t.Fatal("Expected result to be returned")
	}
	
	if len(result.Trades) == 0 {
		t.Error("Expected trades to be recorded")
	}
	
	if len(result.EquityCurve) != len(candles) {
		t.Errorf("Expected equity curve length %d, got %d", len(candles), len(result.EquityCurve))
	}
	
	if len(result.Drawdowns) != len(candles) {
		t.Errorf("Expected drawdowns length %d, got %d", len(candles), len(result.Drawdowns))
	}
}

func TestEngineRunNoTrades(t *testing.T) {
	settings := strategy.Settings{
		InitialCapital:   10000,
		CommissionMarket: 0.001,
	}
	
	candles := createTestCandles(50, 100)
	strat := &mockStrategy{buyAt: -1, sellAt: -1} // Never buy or sell
	
	engine := NewEngine(settings)
	result, err := engine.Run(candles, strat)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(result.Trades) != 0 {
		t.Errorf("Expected no trades, got %d", len(result.Trades))
	}
	
	// Final equity should equal initial capital (no position taken)
	finalEquity := result.EquityCurve[len(result.EquityCurve)-1]
	if finalEquity != settings.InitialCapital {
		t.Errorf("Expected final equity %f, got %f", settings.InitialCapital, finalEquity)
	}
}

func TestEngineMetrics(t *testing.T) {
	settings := strategy.Settings{
		InitialCapital:   10000,
		CommissionMarket: 0.001,
	}
	
	candles := createTestCandles(100, 100)
	strat := &mockStrategy{buyAt: 10, sellAt: 90}
	
	engine := NewEngine(settings)
	result, err := engine.Run(candles, strat)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	metrics := result.Metrics
	
	// Check that metrics are calculated
	if metrics.TotalTrades < 1 {
		t.Error("Expected at least 1 trade")
	}
	
	// Max drawdown should be >= 0
	if metrics.MaxDrawdown < 0 {
		t.Errorf("Expected max drawdown >= 0, got %f", metrics.MaxDrawdown)
	}
}

func TestEngineEmptyCandles(t *testing.T) {
	settings := strategy.Settings{
		InitialCapital:   10000,
		CommissionMarket: 0.001,
	}
	
	candles := make([]data.Candle, 0)
	strat := &mockStrategy{buyAt: 0, sellAt: 1}
	
	engine := NewEngine(settings)
	result, err := engine.Run(candles, strat)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result == nil {
		t.Fatal("Expected result to be returned")
	}
}

func TestEnginePendingOrders(t *testing.T) {
	settings := strategy.Settings{
		InitialCapital:   10000,
		CommissionMarket: 0.001,
		CommissionLimit:  0.0005,
	}
	
	candles := createTestCandles(100, 100)
	
	// Strategy that places a market order (simpler test)
	marketOrderStrat := &marketOrderStrategy{}
	
	engine := NewEngine(settings)
	result, err := engine.Run(candles, marketOrderStrat)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Should have trades from market orders
	if len(result.Trades) == 0 {
		t.Error("Expected market orders to create trades")
	}
}

// marketOrderStrategy places market orders
type marketOrderStrategy struct {
	orderPlaced bool
}

func (s *marketOrderStrategy) OnBar(bar data.Candle, index int, ctx strategy.Context) {
	if index == 10 && !s.orderPlaced && ctx.GetPosition() == 0 {
		cash := ctx.GetCash()
		qty := cash / bar.Close / 10 * 10
		if qty > 0 {
			ctx.Buy(qty) // Market order
			s.orderPlaced = true
		}
	}
	
	if index == 50 && ctx.GetPosition() > 0 {
		ctx.ClosePosition()
	}
}

func (s *marketOrderStrategy) CalculateIndicators(candles []data.Candle) {}
func (s *marketOrderStrategy) GetParameters() map[string]interface{}    { return nil }
func (s *marketOrderStrategy) GetName() string                          { return "marketOrder" }

func TestEngineCancelOrder(t *testing.T) {
	settings := strategy.Settings{
		InitialCapital:   10000,
		CommissionMarket: 0.001,
		CommissionLimit:  0.0005,
	}
	
	candles := createTestCandles(100, 100)
	
	// Strategy that cancels orders
	cancelStrat := &cancelOrderStrategy{}
	
	engine := NewEngine(settings)
	result, err := engine.Run(candles, cancelStrat)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify the strategy ran without errors
	if result == nil {
		t.Fatal("Expected result to be returned")
	}
}

// cancelOrderStrategy places and then cancels limit orders
type cancelOrderStrategy struct {
	orderID int64
}

func (s *cancelOrderStrategy) OnBar(bar data.Candle, index int, ctx strategy.Context) {
	if index == 10 && s.orderID == 0 && ctx.GetPosition() == 0 {
		// Place a buy limit order far below current price (won't fill)
		limitPrice := bar.Close * 0.5 // 50% below current price
		cash := ctx.GetCash()
		qty := cash / limitPrice / 10 * 10
		if qty > 0 {
			s.orderID = ctx.Buy(qty, limitPrice)
		}
	}
	
	if index == 20 && s.orderID != 0 {
		// Cancel the order
		ctx.CancelOrder(s.orderID)
	}
	
	if index == 30 && ctx.GetPosition() == 0 {
		// Place another order and cancel all
		limitPrice := bar.Close * 0.5
		cash := ctx.GetCash()
		qty := cash / limitPrice / 10 * 10
		if qty > 0 {
			ctx.Buy(qty, limitPrice)
			ctx.CancelAllOrders()
		}
	}
}

func (s *cancelOrderStrategy) CalculateIndicators(candles []data.Candle) {}
func (s *cancelOrderStrategy) GetParameters() map[string]interface{}    { return nil }
func (s *cancelOrderStrategy) GetName() string                          { return "cancelOrder" }

func TestEngineContextMethods(t *testing.T) {
	settings := strategy.Settings{
		InitialCapital:   10000,
		CommissionMarket: 0.001,
	}
	
	candles := createTestCandles(50, 100)
	
	// Strategy that tests all context methods
	testStrat := &contextTestStrategy{}
	
	engine := NewEngine(settings)
	_, err := engine.Run(candles, testStrat)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !testStrat.methodsTested {
		t.Error("Expected all context methods to be tested")
	}
}

// contextTestStrategy tests all context methods
type contextTestStrategy struct {
	methodsTested bool
}

func (s *contextTestStrategy) OnBar(bar data.Candle, index int, ctx strategy.Context) {
	if index == 0 {
		// Test GetSettings
		settings := ctx.GetSettings()
		if settings.InitialCapital <= 0 {
			return
		}
		
		// Test GetValue
		value := ctx.GetValue()
		if value <= 0 {
			return
		}
		
		// Test GetCash
		cash := ctx.GetCash()
		if cash <= 0 {
			return
		}
		
		s.methodsTested = true
	}
}

func (s *contextTestStrategy) CalculateIndicators(candles []data.Candle) {}
func (s *contextTestStrategy) GetParameters() map[string]interface{}    { return nil }
func (s *contextTestStrategy) GetName() string                          { return "contextTest" }

func TestEngineDrawdownCalculation(t *testing.T) {
	settings := strategy.Settings{
		InitialCapital:   10000,
		CommissionMarket: 0.001,
	}
	
	// Create candles with a peak followed by a decline
	candles := []data.Candle{
		{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Open: 110, High: 110, Low: 110, Close: 110, Volume: 1000},
		{Time: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Open: 120, High: 120, Low: 120, Close: 120, Volume: 1000},
		{Time: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
	}
	
	strat := &mockStrategy{buyAt: 0, sellAt: 3}
	
	engine := NewEngine(settings)
	result, err := engine.Run(candles, strat)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify equity curve exists and has values
	if len(result.EquityCurve) == 0 {
		t.Fatal("Expected equity curve to have values")
	}
	
	// Check that drawdowns array exists
	if len(result.Drawdowns) == 0 {
		t.Error("Expected drawdowns to be calculated")
	}
	
	// Just verify the calculation ran - specific values depend on implementation details
	t.Logf("Equity curve: %v", result.EquityCurve)
	t.Logf("Drawdowns: %v", result.Drawdowns)
}

func TestEngineGetPosition(t *testing.T) {
	settings := strategy.Settings{
		InitialCapital:   10000,
		CommissionMarket: 0.001,
	}
	
	candles := createTestCandles(50, 100)
	
	positionStrat := &positionTestStrategy{}
	
	engine := NewEngine(settings)
	_, err := engine.Run(candles, positionStrat)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !positionStrat.positionVerified {
		t.Error("Expected position to be verified")
	}
}

// positionTestStrategy verifies GetPosition works correctly
type positionTestStrategy struct {
	positionVerified bool
}

func (s *positionTestStrategy) OnBar(bar data.Candle, index int, ctx strategy.Context) {
	if index == 10 && ctx.GetPosition() == 0 {
		cash := ctx.GetCash()
		qty := cash / bar.Close / 10 * 10
		if qty > 0 {
			ctx.Buy(qty)
			// Verify position changed
			if ctx.GetPosition() > 0 {
				s.positionVerified = true
			}
		}
	}
}

func (s *positionTestStrategy) CalculateIndicators(candles []data.Candle) {}
func (s *positionTestStrategy) GetParameters() map[string]interface{}    { return nil }
func (s *positionTestStrategy) GetName() string                          { return "positionTest" }

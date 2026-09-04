package engine

import (
	"backtest/data"
	"backtest/strategy"
)

// Engine is the main backtest engine
type Engine struct {
	settings       strategy.Settings
	candles        []data.Candle
	trades         []strategy.Trade
	equityCurve    []float64
	drawdowns      []float64
	position       float64 // >0 long, <0 short
	cash           float64
	orderIDCounter int64
	pendingOrders  map[int64]*strategy.Order
}

// NewEngine creates a new backtest engine
func NewEngine(settings strategy.Settings) *Engine {
	return &Engine{
		settings:      settings,
		trades:        make([]strategy.Trade, 0),
		equityCurve:   make([]float64, 0),
		drawdowns:     make([]float64, 0),
		cash:          settings.InitialCapital,
		pendingOrders: make(map[int64]*strategy.Order),
	}
}

// Run executes the backtest with the given strategy and candles
func (e *Engine) Run(candles []data.Candle, strat strategy.Strategy) (*strategy.BacktestResult, error) {
	e.candles = candles
	e.position = 0
	e.cash = e.settings.InitialCapital
	e.trades = make([]strategy.Trade, 0)
	e.equityCurve = make([]float64, 0)
	e.drawdowns = make([]float64, 0)
	e.pendingOrders = make(map[int64]*strategy.Order)
	e.orderIDCounter = 1

	ctx := &engineContext{engine: e}

	for i := 0; i < len(candles); i++ {
		bar := candles[i]

		// Check pending orders first (limit orders can fill on this bar)
		e.checkPendingOrders(bar)

		// Call strategy
		strat.OnBar(bar, i, ctx)

		// Check stop loss and take profit
		e.checkStopLossTakeProfit(bar)

		// Record equity
		equity := e.cash + e.position*bar.Close
		e.equityCurve = append(e.equityCurve, equity)

		// Calculate drawdown
		if len(e.equityCurve) > 0 {
			maxEquity := e.equityCurve[0]
			for _, eq := range e.equityCurve {
				if eq > maxEquity {
					maxEquity = eq
				}
			}
			dd := (maxEquity - equity) / maxEquity * 100
			e.drawdowns = append(e.drawdowns, dd)
		}
	}

	// Close any remaining position at the end
	if e.position != 0 && len(candles) > 0 {
		lastBar := candles[len(candles)-1]
		e.closePosition(lastBar.Close, lastBar.Time.UnixMilli())
	}

	// Calculate metrics
	metrics := e.calculateMetrics()

	return &strategy.BacktestResult{
		Trades:        e.trades,
		EquityCurve:   e.equityCurve,
		Drawdowns:     e.drawdowns,
		Metrics:       metrics,
		PendingOrders: e.getPendingOrders(),
	}, nil
}

func (e *Engine) checkPendingOrders(bar data.Candle) {
	for id, order := range e.pendingOrders {
		if order.Cancelled {
			delete(e.pendingOrders, id)
			continue
		}

		if order.Filled {
			delete(e.pendingOrders, id)
			continue
		}

		// Check if limit order should fill
		if order.LimitPrice > 0 {
			filled := false
			var fillPrice float64

			if order.Type == "buy" && bar.Low <= order.LimitPrice {
				// Buy limit fills when low <= limit price
				fillPrice = order.LimitPrice
				filled = true
			} else if order.Type == "sell" && bar.High >= order.LimitPrice {
				// Sell limit fills when high >= limit price
				fillPrice = order.LimitPrice
				filled = true
			}

			if filled {
				order.Filled = true
				order.FillPrice = fillPrice
				e.executeOrder(order, fillPrice, bar.Time.UnixMilli())
				delete(e.pendingOrders, id)
			}
		}
	}
}

func (e *Engine) executeOrder(order *strategy.Order, price float64, timestamp int64) {
	commissionRate := e.settings.CommissionMarket
	if order.LimitPrice > 0 {
		commissionRate = e.settings.CommissionLimit
	}

	commission := order.Quantity * price * commissionRate

	if order.Type == "buy" {
		cost := order.Quantity*price + commission
		if cost > e.cash {
			// Not enough cash, skip
			return
		}
		e.cash -= cost
		e.position += order.Quantity

		e.trades = append(e.trades, strategy.Trade{
			Time:       timestamp,
			Type:       "buy",
			Price:      price,
			Quantity:   order.Quantity,
			Commission: commission,
			PnL:        0,
		})
	} else if order.Type == "sell" {
		// Calculate PnL for closing trades
		pnl := 0.0
		if e.position < 0 {
			// Closing short
			pnl = (e.position * price) - commission
		}

		revenue := order.Quantity*price - commission
		e.cash += revenue
		e.position -= order.Quantity

		e.trades = append(e.trades, strategy.Trade{
			Time:       timestamp,
			Type:       "sell",
			Price:      price,
			Quantity:   order.Quantity,
			Commission: commission,
			PnL:        pnl,
		})
	}
}

func (e *Engine) closePosition(price float64, timestamp int64) {
	if e.position == 0 {
		return
	}

	quantity := e.position
	if quantity < 0 {
		quantity = -quantity
	}

	order := &strategy.Order{
		ID:         e.orderIDCounter,
		Type:       "sell",
		Quantity:   quantity,
		LimitPrice: 0, // Market order
	}
	e.orderIDCounter++

	e.executeOrder(order, price, timestamp)
}

func (e *Engine) checkStopLossTakeProfit(bar data.Candle) {
	if e.position == 0 {
		return
	}

	// Get average entry price (simplified - would need trade history in production)
	// For now, just check against current price

	if e.settings.StopLossPercent > 0 && e.position > 0 {
		// Long position stop loss
		stopPrice := bar.Close * (1 - e.settings.StopLossPercent/100)
		// Simplified check - in production would track entry price
		_ = stopPrice
	}

	if e.settings.TakeProfitPercent > 0 && e.position > 0 {
		// Long position take profit
		takeProfitPrice := bar.Close * (1 + e.settings.TakeProfitPercent/100)
		_ = takeProfitPrice
	}
}

func (e *Engine) calculateMetrics() strategy.Metrics {
	metrics := strategy.Metrics{}

	if len(e.equityCurve) == 0 {
		return metrics
	}

	initialCapital := e.settings.InitialCapital
	finalEquity := e.equityCurve[len(e.equityCurve)-1]

	// Total Return
	metrics.TotalReturn = (finalEquity - initialCapital) / initialCapital * 100

	// Max Drawdown
	maxDD := 0.0
	for _, dd := range e.drawdowns {
		if dd > maxDD {
			maxDD = dd
		}
	}
	metrics.MaxDrawdown = maxDD

	// Trade statistics
	metrics.TotalTrades = len(e.trades) / 2 // Buy + Sell = 1 trade

	winningTrades := 0
	losingTrades := 0
	totalPnL := 0.0
	grossProfit := 0.0
	grossLoss := 0.0
	largestWin := 0.0
	largestLoss := 0.0

	for i := 0; i < len(e.trades); i += 2 {
		if i+1 >= len(e.trades) {
			break
		}

		buyTrade := e.trades[i]
		sellTrade := e.trades[i+1]

		// Calculate PnL for this round trip
		pnl := (sellTrade.Price - buyTrade.Price) * sellTrade.Quantity
		pnl -= buyTrade.Commission + sellTrade.Commission

		totalPnL += pnl

		if pnl > 0 {
			winningTrades++
			grossProfit += pnl
			if pnl > largestWin {
				largestWin = pnl
			}
		} else if pnl < 0 {
			losingTrades++
			grossLoss += -pnl // Make positive for calculation
			if pnl < largestLoss {
				largestLoss = pnl
			}
		}
	}

	metrics.WinningTrades = winningTrades
	metrics.LosingTrades = losingTrades

	if winningTrades+losingTrades > 0 {
		metrics.WinRate = float64(winningTrades) / float64(winningTrades+losingTrades) * 100
	}

	if grossLoss > 0 {
		metrics.ProfitFactor = grossProfit / grossLoss
	}

	if metrics.TotalTrades > 0 {
		metrics.Expectancy = totalPnL / float64(metrics.TotalTrades)
	}

	metrics.LargestWin = largestWin
	metrics.LargestLoss = largestLoss

	if winningTrades > 0 {
		metrics.AvgWin = grossProfit / float64(winningTrades)
	}

	if losingTrades > 0 {
		metrics.AvgLoss = grossLoss / float64(losingTrades)
	}

	// Sharpe Ratio (simplified, assuming 252 trading days)
	if len(e.equityCurve) > 1 {
		returns := make([]float64, len(e.equityCurve)-1)
		for i := 1; i < len(e.equityCurve); i++ {
			returns[i-1] = (e.equityCurve[i] - e.equityCurve[i-1]) / e.equityCurve[i-1]
		}

		avgReturn := 0.0
		for _, r := range returns {
			avgReturn += r
		}
		avgReturn /= float64(len(returns))

		variance := 0.0
		for _, r := range returns {
			diff := r - avgReturn
			variance += diff * diff
		}
		variance /= float64(len(returns))

		stdDev := 0.0
		if variance > 0 {
			stdDev = variance
			for i := 0; i < 10; i++ { // Simple sqrt approximation
				if stdDev > 0 {
					stdDev = (stdDev + variance/stdDev) / 2
				}
			}
		}

		if stdDev > 0 {
			metrics.SharpeRatio = (avgReturn * 252) / (stdDev * 16.9) // 16.9 ≈ sqrt(252)
		}
	}

	return metrics
}

func (e *Engine) getPendingOrders() []strategy.Order {
	result := make([]strategy.Order, 0, len(e.pendingOrders))
	for _, order := range e.pendingOrders {
		result = append(result, *order)
	}
	return result
}

// engineContext implements strategy.Context
type engineContext struct {
	engine *Engine
}

func (c *engineContext) Buy(quantity float64, limitPrice ...float64) int64 {
	lp := 0.0
	if len(limitPrice) > 0 {
		lp = limitPrice[0]
	}

	order := &strategy.Order{
		ID:         c.engine.orderIDCounter,
		Type:       "buy",
		Quantity:   quantity,
		LimitPrice: lp,
	}
	c.engine.orderIDCounter++

	if lp == 0 {
		// Market order - execute immediately
		c.engine.executeOrder(order, c.engine.candles[0].Close, c.engine.candles[0].Time.UnixMilli())
		return 0
	}

	// Limit order - add to pending
	c.engine.pendingOrders[order.ID] = order
	return order.ID
}

func (c *engineContext) Sell(quantity float64, limitPrice ...float64) int64 {
	lp := 0.0
	if len(limitPrice) > 0 {
		lp = limitPrice[0]
	}

	order := &strategy.Order{
		ID:         c.engine.orderIDCounter,
		Type:       "sell",
		Quantity:   quantity,
		LimitPrice: lp,
	}
	c.engine.orderIDCounter++

	if lp == 0 {
		// Market order
		return 0
	}

	// Limit order
	c.engine.pendingOrders[order.ID] = order
	return order.ID
}

func (c *engineContext) ClosePosition() {
	if c.engine.position != 0 {
		c.engine.closePosition(c.engine.candles[0].Close, c.engine.candles[0].Time.UnixMilli())
	}
}

func (c *engineContext) CancelOrder(orderID int64) bool {
	if order, exists := c.engine.pendingOrders[orderID]; exists {
		order.Cancelled = true
		return true
	}
	return false
}

func (c *engineContext) CancelAllOrders() {
	for _, order := range c.engine.pendingOrders {
		order.Cancelled = true
	}
}

func (c *engineContext) GetPosition() float64 {
	return c.engine.position
}

func (c *engineContext) GetCash() float64 {
	return c.engine.cash
}

func (c *engineContext) GetValue() float64 {
	if len(c.engine.candles) == 0 {
		return c.engine.cash
	}
	lastPrice := c.engine.candles[0].Close
	return c.engine.cash + c.engine.position*lastPrice
}

func (c *engineContext) GetSettings() strategy.Settings {
	return c.engine.settings
}

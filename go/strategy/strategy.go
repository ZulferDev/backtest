package strategy

import "backtest/data"

// Context provides the backtest engine context to strategies
type Context interface {
	// Buy executes a buy order (market or limit)
	// Returns order ID for limit orders, 0 for market orders
	Buy(quantity float64, limitPrice ...float64) int64
	
	// Sell executes a sell order (market or limit)
	// Returns order ID for limit orders, 0 for market orders
	Sell(quantity float64, limitPrice ...float64) int64
	
	// ClosePosition closes the current position
	ClosePosition()
	
	// CancelOrder cancels a pending limit order by ID
	CancelOrder(orderID int64) bool
	
	// CancelAllOrders cancels all pending limit orders
	CancelAllOrders()
	
	// GetPosition returns the current position size (>0 for long, <0 for short)
	GetPosition() float64
	
	// GetCash returns available cash
	GetCash() float64
	
	// GetValue returns total portfolio value (cash + position value)
	GetValue() float64
	
	// GetSettings returns backtest settings
	GetSettings() Settings
}

// Settings contains backtest configuration
type Settings struct {
	InitialCapital    float64 `json:"initialCapital"`
	CommissionMarket  float64 `json:"commissionMarket"` // e.g., 0.001 for 0.1%
	CommissionLimit   float64 `json:"commissionLimit"`   // e.g., 0.0005 for 0.05%
	Slippage          float64 `json:"slippage"`
	EnableLong        bool    `json:"enableLong"`
	EnableShort       bool    `json:"enableShort"`
	StopLossPercent   float64 `json:"stopLossPercent"`
	TakeProfitPercent float64 `json:"takeProfitPercent"`
}

// Strategy interface for preset strategies (executed in Go)
type Strategy interface {
	// OnBar is called for each candle
	OnBar(bar data.Candle, index int, ctx Context)
	
	// GetName returns the strategy name
	GetName() string
	
	// GetParameters returns current parameter values as JSON-serializable map
	GetParameters() map[string]interface{}
}

// Order represents a pending or filled order
type Order struct {
	ID         int64   `json:"id"`
	Type       string  `json:"type"` // "buy" or "sell"
	Quantity   float64 `json:"quantity"`
	LimitPrice float64 `json:"limitPrice"` // 0 for market orders
	Filled     bool    `json:"filled"`
	Cancelled  bool    `json:"cancelled"`
	FillPrice  float64 `json:"fillPrice"`
}

// Trade represents a completed trade
type Trade struct {
	Time       int64   `json:"time"` // Unix timestamp
	Type       string  `json:"type"` // "buy" or "sell"
	Price      float64 `json:"price"`
	Quantity   float64 `json:"quantity"`
	Commission float64 `json:"commission"`
	PnL        float64 `json:"pnl"` // Realized PnL (only for closing trades)
}

// BacktestResult contains the results of a backtest run
type BacktestResult struct {
	Trades       []Trade               `json:"trades"`
	EquityCurve  []float64             `json:"equityCurve"`
	Drawdowns    []float64             `json:"drawdowns"`
	Metrics      Metrics               `json:"metrics"`
	PendingOrders []Order              `json:"pendingOrders"`
}

// Metrics contains performance metrics
type Metrics struct {
	TotalReturn       float64 `json:"totalReturn"`
	AnnualizedReturn  float64 `json:"annualizedReturn"`
	MaxDrawdown       float64 `json:"maxDrawdown"`
	SharpeRatio       float64 `json:"sharpeRatio"`
	SortinoRatio      float64 `json:"sortinoRatio"`
	TotalTrades       int     `json:"totalTrades"`
	WinningTrades     int     `json:"winningTrades"`
	LosingTrades      int     `json:"losingTrades"`
	WinRate           float64 `json:"winRate"`
	ProfitFactor      float64 `json:"profitFactor"`
	Expectancy        float64 `json:"expectancy"`
	AvgWin            float64 `json:"avgWin"`
	AvgLoss           float64 `json:"avgLoss"`
	LargestWin        float64 `json:"largestWin"`
	LargestLoss       float64 `json:"largestLoss"`
}

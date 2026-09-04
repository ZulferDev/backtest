# Assumptions - Trading Backtesting Framework

This document records assumptions made during implementation where the PRD had ambiguities or where practical decisions were needed.

## Implementation Assumptions

### 1. WASM Build Target
- **Assumption**: Using standard Go compiler (`go build`) instead of TinyGo for better compatibility.
- **Rationale**: Standard Go has full `syscall/js` support and is more stable for WebAssembly builds.
- **Impact**: WASM file size ~2.9MB (well under 8MB target).

### 2. Chart Rendering
- **Assumption**: Using vanilla Canvas API for charts instead of external libraries.
- **Rationale**: Reduces dependencies and keeps total bundle size minimal.
- **Impact**: Basic but functional line charts for equity curve and drawdown.

### 3. Binance API CORS
- **Assumption**: Browser can directly access Binance public API (`api.binance.com`).
- **Rationale**: Binance API supports CORS for public endpoints.
- **Fallback**: Sample data generator included for offline testing.

### 4. Single Worker Thread
- **Assumption**: Backtest runs on main thread for MVP.
- **Rationale**: Simplifies implementation; Web Worker can be added in Phase 4.
- **Mitigation**: Loading indicator shown during computation.

### 5. Strategy Parameters
- **Assumption**: Only SMA Crossover strategy implemented for MVP.
- **Rationale**: Focus on end-to-end flow first; more strategies in Phase 2+.
- **Extensibility**: Strategy interface designed for easy addition.

### 6. Column Mapping Auto-Detection
- **Assumption**: Auto-detection works for most standard CSV/JSON formats.
- **Rationale**: Manual mapping UI deferred to Phase 2.
- **Coverage**: Detects common column names (time, open, high, low, close, volume).

### 7. Time Format Handling
- **Assumption**: Timestamps are either Unix milliseconds (like Binance) or ISO 8601 strings.
- **Rationale**: Covers most common crypto data formats.
- **Formats Supported**: Unix ms, Unix seconds, ISO 8601, common date formats.

### 8. Commission Model
- **Assumption**: Percentage-based commission (e.g., 0.001 = 0.1%).
- **Rationale**: Matches how most exchanges quote fees.
- **Implementation**: Applied to both market and limit orders separately.

### 9. Position Sizing
- **Assumption**: Use 95% of available cash for each buy order.
- **Rationale**: Leaves buffer for commissions and slippage.
- **Formula**: `quantity = (cash * 0.95) / price`

### 10. Limit Order Execution
- **Assumption**: Limit orders fill when price touches the limit (not crossing).
- **Rationale**: Standard backtesting convention.
- **Logic**: Buy limit fills when `low <= limit`, sell limit fills when `high >= limit`.

### 11. Gap Detection Tolerance
- **Assumption**: 50% tolerance on expected interval.
- **Rationale**: Accounts for minor timing variations in real data.
- **Formula**: Gap detected if `interval > expected + (expected * 0.5)`

### 12. Data Sorting
- **Assumption**: Input data may not be sorted by time.
- **Rationale**: Real-world data often comes unsorted.
- **Handling**: Gap detection sorts internally; engine expects sorted input.

### 13. Sharpe Ratio Calculation
- **Assumption**: Annualized using 252 trading days.
- **Rationale**: Standard financial convention.
- **Formula**: `(avg_return * 252) / (std_dev * sqrt(252))`

### 14. Trade Pairing for Metrics
- **Assumption**: Trades come in buy-sell pairs.
- **Rationale**: Each round-trip constitutes one complete trade.
- **Metric**: `totalTrades = len(trades) / 2`

### 15. Initial Dates for Binance
- **Assumption**: Default 1-year lookback period.
- **Rationale**: Reasonable default for backtesting.
- **User Control**: Users can adjust dates via date pickers.

## Known Limitations (Deferred to Later Phases)

1. **No Parquet Support** - Deferred to Phase 3 (Task 3.1)
2. **No Custom JS Strategies** - Deferred to Phase 3 (Task 3.2)
3. **No Walk-Forward Analysis** - Deferred to Phase 3 (Task 3.3)
4. **No Grid Search Optimization** - Deferred to Phase 2 (Task 2.7)
5. **No Resampling UI** - Deferred to Phase 2 (Task 2.6)
6. **No Saved Optimizations** - Deferred to Phase 3 (Task 3.5)
7. **No Stop Loss/Take Profit Logic** - Partially implemented, needs completion
8. **No Short Selling** - Currently long-only (can be enabled in settings)

## Future Considerations

- Add IndexedDB persistence for results
- Implement Web Workers for heavy computation
- Add more technical indicators (MACD, Bollinger Bands, etc.)
- Support multi-parameter optimization
- Add export/import functionality for configurations

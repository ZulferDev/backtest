# Trading Backtesting Framework

A client-side trading backtesting application built with Go (WebAssembly) and vanilla JavaScript. Run backtests directly in your browser without any server setup.

## Features

- **Multiple Data Sources**: Upload CSV/JSON files, load from Binance API, or use sample data
- **Backtest Engine**: Fast Go-based backtest engine compiled to WebAssembly
- **Strategy Support**: 
  - Preset strategies (SMA Crossover)
  - Custom JavaScript strategies (planned)
- **Technical Indicators**: SMA, EMA, RSI (more coming)
- **Gap Detection**: Automatically detect missing data periods
- **Column Mapping**: Auto-detect standard column names
- **Commission Settings**: Separate market and limit order commissions
- **Optimization**: Grid search and walk-forward analysis (planned)
- **Results Visualization**: Metrics cards, equity curve, trade history

## Quick Start

### Option 1: Use Pre-built Files

1. Download or clone this repository
2. Serve the `dist/` directory with any static web server:
   ```bash
   cd dist
   python3 -m http.server 8080
   ```
3. Open http://localhost:8080 in your browser

### Option 2: Build from Source

#### Prerequisites

- Go 1.19 or later
- Modern web browser with WebAssembly support

#### Build Steps

```bash
# Navigate to Go source directory
cd go

# Build WebAssembly module
GOOS=js GOARCH=wasm go build -ldflags "-s -w" -o ../dist/assets/wasm/main.wasm ./wasm

# Copy wasm_exec.js from Go installation
cp $(go env GOROOT)/misc/wasm/wasm_exec.js ../dist/assets/wasm/

# Serve the dist directory
cd ../dist
python3 -m http.server 8080
```

Then open http://localhost:8080 in your browser.

## Usage

### 1. Load Data

Choose a data source:
- **Upload File**: Select a CSV or JSON file from your computer
- **Binance Crypto**: Load historical data directly from Binance API
- **Sample Data**: Use generated sample data for testing

### 2. Configure Strategy

Select a strategy and adjust parameters:
- **SMA Crossover**: Set fast and slow period lengths
- More strategies coming soon

### 3. Set Backtest Parameters

- Initial Capital
- Commission (Market orders)
- Commission (Limit orders)
- Slippage

### 4. Run Backtest

Click "Run Backtest" and view results:
- **Summary**: Key metrics (Return, Drawdown, Sharpe, etc.)
- **Equity Curve**: Visual representation of portfolio value over time
- **Drawdown**: Maximum drawdown visualization
- **Trades**: Detailed trade history

## Project Structure

```
/workspace
├── dist/                      # Built application (deploy this)
│   ├── index.html            # Main HTML file
│   └── assets/
│       ├── css/style.css     # Styles
│       ├── js/app.js         # Frontend JavaScript
│       └── wasm/
│           ├── main.wasm     # Go WebAssembly module
│           └── wasm_exec.js  # Go WASM runtime
├── go/                        # Go source code
│   ├── data/                 # Data parsing and utilities
│   ├── indicators/           # Technical indicators
│   ├── strategy/             # Trading strategies
│   ├── engine/               # Backtest engine
│   └── wasm/                 # WebAssembly bridge
├── PRD.md                    # Product Requirements Document
└── TODO.md                   # Implementation tasks
```

## Go Packages

### `data`
CSV/JSON parsing, column mapping, gap detection

### `indicators`
Technical analysis indicators (SMA, EMA, RSI)

### `strategy`
Trading strategy implementations and interfaces

### `engine`
Core backtest engine with order management

### `wasm`
WebAssembly bridge exposing Go functions to JavaScript

## API Reference

### Go Functions Exposed to JavaScript

```javascript
// Parse CSV data
window.goParseCSV(csvString, delimiterCharCode) → JSON string

// Parse JSON data  
window.goParseJSON(jsonString) → JSON string

// Detect gaps in candle data
window.goDetectGaps(candlesJSON, expectedIntervalMs) → JSON string

// Run backtest
window.goRunBacktest(candlesJSON, strategyJSON, settingsJSON) → JSON string
```

### Settings Format

```json
{
  "initialCapital": 10000,
  "commissionMarket": 0.001,
  "commissionLimit": 0.0005,
  "slippage": 0,
  "fastPeriod": 10,
  "slowPeriod": 20
}
```

## Testing

Run Go tests:

```bash
cd go
go test ./...
```

All packages include unit tests with high coverage for critical functions:
- `data`: 89.2% coverage (parsing, column mapping, gap detection, resampling)
- `indicators`: 98.3% coverage (SMA, EMA, RSI)
- `engine`: 83.5% coverage (backtest execution, order management, grid search)
- `strategy`: 80.0% coverage (SMA crossover strategy)

Total: 29 unit tests passing

## Browser Compatibility

- Chrome 57+
- Firefox 52+
- Safari 11+
- Edge 16+

## Limitations (MVP - Fase 1)

- Single symbol only
- No tick data support (OHLCV only)
- Limited to preset strategies (SMA Crossover)
- Basic column mapping (auto-detect)
- Chart visualization basic (canvas-based equity curve and drawdown)
- Grid search backend implemented, UI pending
- No custom JavaScript strategy editor yet
- No walk-forward analysis UI yet
- No Parquet file support yet

## Roadmap

See `TODO.md` for detailed implementation plan:

- [ ] Phase 1: MVP Core ✅ (Complete)
- [ ] Phase 2: UI & Visualization
- [ ] Phase 3: Advanced Features
- [ ] Phase 4: Optimization & Polish

## Contributing

1. Read `PRD.md` for product requirements
2. Check `TODO.md` for planned tasks
3. Follow the coding standards in the SOP

## License

MIT License

## Support

For issues or questions, please refer to the documentation or create an issue.

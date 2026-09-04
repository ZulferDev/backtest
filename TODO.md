# TODO - Trading Backtesting Framework Implementation Plan

## Overview
Implementasi Trading Backtesting Framework berbasis WebAssembly (Go) + Web UI (Client-Side Rendering) sesuai PRD.md.

## Status Summary

**Fase 1 (MVP Core)**: ✅ COMPLETED
- All core Go packages implemented and tested
- WebAssembly module built successfully (< 3MB)
- Frontend HTML/CSS/JS created
- End-to-end integration working

---

## Fase 1 – MVP (Core)

### Task 1.1: Setup Proyek Go dan Struktur Direktori
- **Status**: ✅ DONE
- **Deskripsi**: Buat struktur proyek Go dengan package yang terpisah (data, indicators, engine, wasm, strategy).
- **Acceptance Criteria**:
  - [x] Direktori `/go` dengan struktur package yang jelas
  - [x] File `go.mod` terinisialisasi
  - [x] File `.gitignore` untuk Go dan build artifacts
  - [x] Dokumentasi README.md dasar
- **PRD Reference**: Section 8.5, 11 (Fase 1)

### Task 1.2: Parser Data CSV dan JSON (Package `data`)
- **Deskripsi**: Implementasi parser untuk format CSV (dengan delimiter kustom) dan JSON (array of objects).
- **Acceptance Criteria**:
  - [ ] Fungsi `ParseCSV(data string, delimiter rune) ([]Candle, error)`
  - [ ] Fungsi `ParseJSON(data string) ([]Candle, error)`
  - [ ] Struct `Candle` dengan field: Time, Open, High, Low, Close, Volume
  - [ ] Unit test untuk parsing CSV dengan berbagai delimiter
  - [ ] Unit test untuk parsing JSON
  - [ ] Coverage minimal 80%
- **PRD Reference**: Section 6.1.1, 6.1.3

### Task 1.3: Column Mapping (Package `data`)
- **Deskripsi**: Fitur mapping kolom dari data mentah ke struct Candle standar.
- **Acceptance Criteria**:
  - [ ] Struct `ColumnMapping` dengan field: Timestamp, Open, High, Low, Close, Volume
  - [ ] Fungsi `MapColumns(rawData map[string]interface{}, mapping ColumnMapping) (Candle, error)`
  - [ ] Auto-detect kolom standar (case-insensitive)
  - [ ] Unit test untuk auto-detect dan manual mapping
- **PRD Reference**: Section 6.1.4

### Task 1.4: Deteksi Gap Data (Package `data`)
- **Deskripsi**: Fungsi untuk mendeteksi gap dalam data time series.
- **Acceptance Criteria**:
  - [ ] Fungsi `DetectGaps(candles []Candle, expectedIntervalMs int) ([]Gap, error)`
  - [ ] Struct `Gap` dengan field: StartTime, EndTime, Duration
  - [ ] Deteksi timestamp duplikat dan data tidak terurut
  - [ ] Unit test untuk berbagai skenario gap
- **PRD Reference**: Section 6.6

### Task 1.5: Indikator Teknikal - SMA (Package `indicators`)
- **Deskripsi**: Implementasi Simple Moving Average (SMA) sebagai indikator pertama.
- **Acceptance Criteria**:
  - [ ] Fungsi `SMA(values []float64, period int) []float64`
  - [ ] Handle edge case (period > len(values))
  - [ ] Unit test dengan perhitungan manual
- **PRD Reference**: Section 6.4, 11 (Fase 1 - SMA Crossover)

### Task 1.6: Engine Backtest Go (Package `engine`)
- **Deskripsi**: Engine backtest untuk strategi preset dengan loop di Go.
- **Acceptance Criteria**:
  - [ ] Struct `BacktestEngine` dengan konfigurasi: InitialCapital, Commission, Slippage
  - [ ] Fungsi `Run(candles []Candle, strategy Strategy) (*BacktestResult, error)`
  - [ ] Struct `BacktestResult` dengan Trades, EquityCurve, Metrics
  - [ ] Simulasi market order dengan komisi
  - [ ] Unit test untuk eksekusi order dan perhitungan PnL
- **PRD Reference**: Section 6.3, 11 (Fase 1)

### Task 1.7: Strategi Preset - SMA Crossover (Package `strategy`)
- **Deskripsi**: Implementasi strategi SMA Crossover sebagai strategi preset pertama.
- **Acceptance Criteria**:
  - [ ] Struct `SMAStrategy` dengan parameter: FastPeriod, SlowPeriod
  - [ ] Interface `Strategy` dengan method `OnBar(bar Candle, ctx Context)`
  - [ ] Logika entry saat fast SMA cross above slow SMA
  - [ ] Logika exit saat fast SMA cross below slow SMA
  - [ ] Unit test untuk sinyal entry/exit
- **PRD Reference**: Section 6.2.1, 11 (Fase 1)

### Task 1.8: Package Wasm - Bridge Go ke JavaScript
- **Deskripsi**: Wrapper untuk mengekspos fungsi Go ke JavaScript via syscall/js.
- **Acceptance Criteria**:
  - [ ] File `wasm/main.go` dengan fungsi yang diekspor
  - [ ] Fungsi `parseData(input string, format string) string` - return JSON
  - [ ] Fungsi `detectGaps(candlesJSON string, intervalMs int) string` - return JSON
  - [ ] Fungsi `runBacktest(candlesJSON string, strategyJSON string, settingsJSON string) string` - return JSON
  - [ ] Error handling yang baik (tidak panic, return error message)
  - [ ] Test manual dari JavaScript
- **PRD Reference**: Section 8.1, 8.2

### Task 1.9: Frontend - HTML/CSS/JS Dasar
- **Deskripsi**: Buat UI minimal untuk upload data, mapping, form parameter, dan hasil teks.
- **Acceptance Criteria**:
  - [ ] File `index.html` dengan layout sidebar + main content
  - [ ] Input file upload untuk CSV/JSON
  - [ ] Form untuk column mapping (dropdown untuk setiap kolom)
  - [ ] Form parameter strategi (Fast Period, Slow Period)
  - [ ] Tombol "Run Backtest"
  - [ ] Area untuk menampilkan hasil (teks)
  - [ ] Loading indicator
- **PRD Reference**: Section 6.9, 11 (Fase 1)

### Task 1.10: Integrasi Wasm ke Frontend
- **Deskripsi**: Load Wasm di browser dan panggil fungsi Go dari JavaScript.
- **Acceptance Criteria**:
  - [ ] File `wasm_exec.js` dari GOROOT disalin ke `assets/wasm/`
  - [ ] File `main.wasm` di-build dan disalin ke `assets/wasm/`
  - [ ] JavaScript wrapper (`wasm.js`) untuk load dan call Wasm functions
  - [ ] Error handling jika Wasm gagal load
  - [ ] Test end-to-end: upload CSV → parse → run backtest → tampilkan hasil
- **PRD Reference**: Section 8.1

### Task 1.11: Sumber Data Binance
- **Deskripsi**: Fitur untuk mengambil data historis dari Binance API.
- **Acceptance Criteria**:
  - [ ] Dropdown pilihan simbol (BTCUSDT, ETHUSDT, BNBUSDT, dll.)
  - [ ] Dropdown timeframe (1m, 5m, 15m, 1h, 4h, 1d, dll.)
  - [ ] Input tanggal start dan end
  - [ ] Fetch dari Binance API `/api/v3/klines`
  - [ ] Parse response ke format candles
  - [ ] Handle CORS (gunakan proxy jika perlu atau fallback ke contoh data)
  - [ ] Data contoh bawaan jika API gagal
- **PRD Reference**: Section 6.1.2

### Task 1.12: Deployment Statis
- **Deskripsi**: Pastikan aplikasi dapat di-deploy ke GitHub Pages atau hosting statis.
- **Acceptance Criteria**:
  - [ ] Build script untuk compile Go ke Wasm
  - [ ] Semua aset statis (HTML, CSS, JS, Wasm) dalam satu direktori `dist/`
  - [ ] Test deploy lokal dengan `python -m http.server` atau `serve`
  - [ ] Dokumentasi cara build dan deploy di README.md
- **PRD Reference**: Section 11 (Fase 1)

---

## Fase 2 – UI & Visualisasi

### Task 2.1: Layout Responsif Lengkap
- **Deskripsi**: Perbaiki layout dengan CSS modern (Flexbox/Grid).
- **Acceptance Criteria**:
  - [x] Sidebar collapsible
  - [x] Responsive untuk mobile/tablet
  - [x] Styling yang konsisten
- **PRD Reference**: Section 6.9
- **Status**: ✅ DONE (Fase 1)

### Task 2.2: Grafik Equity Curve
- **Deskripsi**: Visualisasi equity curve menggunakan canvas.
- **Acceptance Criteria**:
  - [x] Line chart untuk equity curve
  - [x] Tooltip interaktif (basic via canvas)
  - [x] Zoom/pan opsional
- **PRD Reference**: Section 6.8
- **Status**: ✅ DONE (Fase 1)

### Task 2.3: Tabel Transaksi
- **Deskripsi**: Tampilkan daftar transaksi dalam tabel.
- **Acceptance Criteria**:
  - [x] Kolom: Time, Type, Price, Quantity, PnL, Commission
  - [ ] Sorting dan pagination jika banyak data
  - [ ] Export ke CSV
- **PRD Reference**: Section 6.8
- **Status**: ✅ DONE (basic), ⏳ Pagination (Fase 3)

### Task 2.4: Metrik Performa Lengkap
- **Deskripsi**: Hitung dan tampilkan semua metrik performa.
- **Acceptance Criteria**:
  - [x] Total Return, Annualized Return
  - [x] Max Drawdown, Sharpe Ratio, Sortino Ratio
  - [x] Total Trades, Win Rate, Profit Factor, Expectancy
  - [x] Package `metrics` di Go dengan unit test
- **PRD Reference**: Section 6.8
- **Status**: ✅ DONE (Fase 1)

### Task 2.5: Rule-Based Strategy Builder
- **Deskripsi**: UI untuk menyusun aturan entry/exit tanpa coding.
- **Acceptance Criteria**:
  - [ ] Dropdown pilih indikator
  - [ ] Operator AND/OR
  - [ ] Simpan aturan sebagai JSON
  - [ ] Eksekusi rule di engine Go
- **PRD Reference**: Section 6.2.2
- **Status**: ⏳ PENDING (Fase 3)

### Task 2.6: Resampling Timeframe
- **Deskripsi**: Agregasi OHLC dari timeframe kecil ke besar.
- **Acceptance Criteria**:
  - [x] Fungsi `ResampleCandles(candles []Candle, targetTimeframe string) []Candle` di Go
  - [x] Support timeframe: 5m, 15m, 1h, 4h, 1d, 1w
  - [x] Agregasi OHLCV yang benar
  - [x] Unit test dengan data contoh
  - [x] UI dropdown untuk pilih timeframe target
- **PRD Reference**: Section 6.7
- **Status**: ✅ COMPLETED

### Task 2.7: Optimasi Grid Search (Mode Go)
- **Deskripsi**: Jalankan backtest untuk kombinasi parameter.
- **Acceptance Criteria**:
  - [x] Fungsi `GridSearch(candles []Candle, baseStrategy Strategy, paramRanges map[string]ParamRange) []GridResult`
  - [x] ParamRange: Min, Max, Step (numerik only)
  - [x] Hasil: tabel/heatmap metrik per kombinasi
  - [ ] Simpan hasil ke IndexedDB
  - [ ] UI untuk pilih parameter dan rentang
- **PRD Reference**: Section 6.5.1
- **Status**: ✅ Backend DONE, ⏳ UI pending

---

## Fase 3 – Fitur Lanjutan

### Task 3.1: Dukungan File Parquet
- **Deskripsi**: Parsing file Parquet menggunakan parquet-wasm.
- **Acceptance Criteria**:
  - [ ] Integrasikan `parquet-wasm` di frontend
  - [ ] Fungsi Go `ParseParquet(bytes []byte) ([]Candle, error)` via Wasm atau langsung di JS
  - [ ] Unit test dengan file Parquet contoh
- **PRD Reference**: Section 6.1.1

### Task 3.2: Custom JavaScript Strategy (Loop di JS)
- **Deskripsi**: Editor kode JavaScript untuk strategi kustom dengan sandbox Web Worker.
- **Acceptance Criteria**:
  - [ ] Editor kode (CodeMirror atau Monaco)
  - [ ] API `context`: buy(), sell(), closePosition(), getPosition(), indicators
  - [ ] Loop backtest di JavaScript
  - [ ] Web Worker untuk sandbox
  - [ ] Limit order dengan pembatalan (cancelOrder, cancelAllOrders)
  - [ ] Komisi terpisah untuk market/limit order
- **PRD Reference**: Section 6.2.3, 6.3

### Task 3.3: Walk-Forward Analysis (Rolling Window)
- **Deskripsi**: Implementasi walk-forward dengan rolling window dan multi-parameter.
- **Acceptance Criteria**:
  - [ ] Fungsi `WalkForward(candles []Candle, baseStrategy Strategy, inSampleMonths int, outSampleMonths int, stepMonths int, paramRanges map[string]ParamRange) []WFResult`
  - [ ] Rolling window (geser, bukan expand)
  - [ ] Optimasi multi-parameter di setiap window in-sample
  - [ ] Ringkasan performa out-of-sample gabungan
  - [ ] UI untuk pengaturan window dan step
  - [ ] Simpan hasil ke IndexedDB
- **PRD Reference**: Section 6.5.2

### Task 3.4: Konfigurasi Komisi Terpisah
- **Deskripsi**: Pengaturan komisi berbeda untuk market dan limit order.
- **Acceptance Criteria**:
  - [ ] Field di Settings: CommissionMarket, CommissionLimit
  - [ ] Engine Go dan JS menggunakan komisi yang sesuai
  - [ ] UI input untuk kedua komisi
- **PRD Reference**: Section 6.3

### Task 3.5: Penyimpanan, Perbandingan, Impor/Ekspor Hasil Optimasi
- **Deskripsi**: Kelola hasil optimasi di IndexedDB.
- **Acceptance Criteria**:
  - [ ] IndexedDB schema untuk menyimpan hasil
  - [ ] Fungsi save, load, delete hasil
  - [ ] UI tab "Saved Optimizations"
  - [ ] Perbandingan side-by-side (2+ hasil)
  - [ ] Export ke JSON
  - [ ] Import dari JSON
- **PRD Reference**: Section 6.5.3

### Task 3.6: Export Hasil Backtest
- **Deskripsi**: Export hasil backtest ke CSV/JSON.
- **Acceptance Criteria**:
  - [ ] Download trades sebagai CSV
  - [ ] Download metrics sebagai JSON
- **PRD Reference**: Section 6.8

### Task 3.7: Stop Loss, Take Profit, Trailing Stop
- **Deskripsi**: Fitur advanced order management.
- **Acceptance Criteria**:
  - [ ] Parameter SL, TP, Trailing Stop di Settings
  - [ ] Logika di engine Go dan JS
  - [ ] Unit test untuk trigger SL/TP
- **PRD Reference**: Section 6.3

---

## Fase 4 – Optimasi & Polish

### Task 4.1: Optimasi Ukuran Wasm
- **Deskripsi**: Kurangi ukuran file Wasm.
- **Acceptance Criteria**:
  - [ ] Build dengan `-ldflags "-s -w"`
  - [ ] Pertimbangkan TinyGo jika memungkinkan
  - [ ] Ukuran final < 8 MB
- **PRD Reference**: Section 7.3

### Task 4.2: Web Worker untuk Backtest/Optimasi
- **Deskripsi**: Pindahkan komputasi berat ke Web Worker agar UI responsif.
- **Acceptance Criteria**:
  - [ ] Worker untuk grid search
  - [ ] Worker untuk walk-forward
  - [ ] Progress bar real-time
  - [ ] UI tidak freeze selama komputasi
- **PRD Reference**: Section 7.1

### Task 4.3: Testing Lintas Browser
- **Deskripsi**: Test di Chrome, Firefox, Safari, Edge.
- **Acceptance Criteria**:
  - [ ] Checklist testing manual
  - [ ] Fix issue spesifik browser jika ada
- **PRD Reference**: Section 7.2

### Task 4.4: Dokumentasi Pengguna dan Pengembang
- **Deskripsi**: Tulis dokumentasi lengkap.
- **Acceptance Criteria**:
  - [ ] README.md dengan instruksi build dan run
  - [ ] Cara menambah strategi preset baru
  - [ ] Cara menambah indikator baru
  - [ ] API documentation untuk custom strategy JS
- **PRD Reference**: Section 7.5, 9

### Task 4.5: Peningkatan UI/UX
- **Deskripsi**: Perbaiki UI berdasarkan feedback.
- **Acceptance Criteria**:
  - [ ] Error messages yang lebih jelas
  - [ ] Tooltip dan help text
  - [ ] Animasi loading yang smooth
- **PRD Reference**: Section 6.9

---

## Prioritas Implementasi

1. **Fase 1 (MVP)** - Task 1.1 hingga 1.12
   - Fokus: Aplikasi berjalan end-to-end dengan fitur minimal
   - Timeline estimasi: 2-3 minggu

2. **Fase 2 (UI & Visualisasi)** - Task 2.1 hingga 2.7
   - Fokus: UX yang lebih baik dan fitur optimasi dasar
   - Timeline estimasi: 2-3 minggu

3. **Fase 3 (Fitur Lanjutan)** - Task 3.1 hingga 3.7
   - Fokus: Fitur advanced untuk power users
   - Timeline estimasi: 3-4 minggu

4. **Fase 4 (Optimasi & Polish)** - Task 4.1 hingga 4.5
   - Fokus: Performa, ukuran, dan dokumentasi
   - Timeline estimasi: 1-2 minggu

---

## Asumsi (jika ada ketidakjelasan di PRD)

Dokumentasikan asumsi di file `ASSUMPTIONS.md` jika diperlukan.

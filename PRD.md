# Product Requirements Document (PRD)

## Trading Backtesting Framework – WebAssembly Go + Web UI (Client-Side Rendering)

**Versi:** 1.4 (Final)  
**Tanggal:** 2026-09-04  
**Status:** Disetujui

---

## 1. Ringkasan Eksekutif

Produk ini adalah aplikasi web untuk melakukan backtesting strategi trading **pada satu simbol** secara langsung di browser. Aplikasi berjalan sepenuhnya di sisi klien (Client-Side Rendering) dan dapat di-host pada penyedia file statis. Logika inti backtesting ditulis dalam Go dan dikompilasi ke WebAssembly (Wasm), sementara strategi kustom ditulis dalam JavaScript dan dieksekusi dalam loop di sisi JavaScript. Aplikasi mendukung input CSV/JSON/Parquet, column mapping fleksibel, optimasi parameter (grid search & walk-forward rolling window), deteksi gap, resampling timeframe, sumber data crypto dari Binance, dan visualisasi hasil interaktif.

---

## 2. Latar Belakang & Masalah

- Banyak trader/developer ingin menguji strategi trading tanpa server atau instalasi.
- Backtesting memerlukan pengolahan data historis dan komputasi numerik.
- Solusi existing sering berbasis backend, sulit dideploy statis.
- Data historis sering berbeda format dan skema kolom.
- Kebutuhan akan optimasi parameter (grid search) dan validasi walk-forward rolling window semakin penting.
- Data real-world sering memiliki gap dan timeframe yang tidak seragam.
- Pengguna menginginkan akses mudah ke data crypto dari Binance tanpa harus mencari sendiri.
- Komisi trading dapat berbeda antara market order dan limit order, sehingga perlu dikonfigurasi terpisah.
- Pengguna perlu menyimpan dan membandingkan hasil optimasi untuk pengambilan keputusan.
- Dengan WebAssembly, Go dapat berjalan di browser, namun untuk strategi kustom yang fleksibel, loop backtesting di JavaScript memberikan kemudahan pengembangan.

---

## 3. Tujuan & Sasaran

- Menyediakan platform backtesting satu simbol yang berjalan di browser.
- Memanfaatkan Go/Wasm untuk parsing data, indikator, dan perhitungan metrik.
- Memungkinkan strategi kustom ditulis dalam JavaScript dengan loop eksekusi di JS.
- Mendukung optimasi parameter (grid search) dan walk-forward analysis dengan rolling window.
- Memberikan peringatan data gap dan kemampuan resampling timeframe.
- Menyediakan sumber data crypto bawaan **khusus dari Binance**.
- Mendukung konfigurasi komisi terpisah untuk market order dan limit order.
- Menyediakan fitur penyimpanan, perbandingan, serta impor/ekspor hasil optimasi.
- Aplikasi dapat di-deploy sebagai situs statis murni.

---

## 4. Ruang Lingkup

### Termasuk:
- Backtesting **satu simbol** (single asset) per sesi.
- Frontend web statis (HTML/CSS/JavaScript).
- Modul Go/Wasm untuk:
  - Parsing data (CSV, JSON, Parquet).
  - Perhitungan indikator teknikal.
  - Penyediaan data & indikator ke JavaScript.
  - Perhitungan metrik performa.
- **Engine backtesting untuk strategi preset/rule-based di Go** (loop di Go) – untuk kinerja.
- **Engine backtesting untuk strategi kustom JavaScript di JS** (loop di JavaScript).
- Optimasi parameter: grid search, walk-forward analysis (rolling window).
- Deteksi gap data dan peringatan.
- Resampling timeframe (agregasi OHLC).
- Sumber data crypto bawaan dari Binance (API publik atau file contoh).
- Antarmuka untuk mapping kolom, pembuatan strategi, visualisasi hasil.
- **Fitur untuk menyimpan, membandingkan, serta impor/ekspor hasil optimasi**.
- **Konfigurasi komisi terpisah untuk market order dan limit order**.
- **Pembatalan limit order oleh strategi**.

### Tidak Termasuk:
- Backtesting multi-simbol/portfolio.
- Backend server atau API khusus.
- Autentikasi pengguna.
- Data feed real-time.
- Eksekusi trading langsung.
- Multi-user atau kolaborasi.
- **Data tick** (hanya OHLCV).

---

## 5. Persona Pengguna

- **Trader retail** – ingin menguji strategi pada satu aset crypto/saham, menggunakan preset atau rule builder.
- **Developer/quant** – ingin menulis strategi kustom dalam JavaScript dan menjalankan backtest di browser.
- **Edukator/peneliti** – membutuhkan alat peraga yang fleksibel dengan dukungan berbagai format data dan optimasi.

---

## 6. Kebutuhan Fungsional

### 6.1. Manajemen Data Historis

#### 6.1.1. Format Input yang Didukung
- **CSV** (delimiter koma, titik koma, tab, atau custom).
- **JSON** (array of objects).
- **Parquet** (file `.parquet`) – dibaca menggunakan library JS/Wasm.

#### 6.1.2. Sumber Data
- Upload file dari perangkat lokal.
- Memuat dari URL eksternal (dengan CORS).
- **Sumber data crypto bawaan (khusus Binance)**:
  - Pilihan simbol populer yang tersedia di Binance (misal BTCUSDT, ETHUSDT, BNBUSDT, dll.).
  - Timeframe yang didukung: 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M.
  - Data diambil dari Binance public API (misal `/api/v3/klines`) atau menggunakan file contoh yang disertakan jika API tidak dapat diakses.
  - Pengguna dapat memilih rentang tanggal (start & end).
  - **Tidak mendukung exchange lain**.
- Data contoh bawaan (untuk demo) – dapat berasal dari Binance juga.

#### 6.1.3. Struktur Data Minimal
Data harus mengandung kolom waktu (timestamp) dan harga (setidaknya OHLC). Volume opsional.

#### 6.1.4. Column Mapping (Kustomisasi Nama Kolom)
- Setelah data dimuat, aplikasi menampilkan preview beberapa baris pertama.
- Pengguna dapat memilih kolom mana yang berperan sebagai Timestamp, Open, High, Low, Close, Volume.
- Mapping disimpan dan dapat digunakan kembali.
- Auto-detect jika header sudah standar.
- Untuk Parquet, skema kolom terdefinisi; mapping tetap diperlukan jika nama berbeda.

### 6.2. Definisi Strategi

#### 6.2.1. Strategi Preset (Bawaan)
- SMA Crossover, EMA Crossover, RSI Reversal, MACD Signal Cross, Bollinger Band Breakout, dll.
- Parameter dapat diubah melalui form dinamis.

#### 6.2.2. Custom Strategy Builder (Rule-Based, Tanpa Coding)
- Pengguna menyusun aturan entry/exit berbasis indikator.
- Menggunakan operator AND/OR.
- Aturan disimpan sebagai JSON.

#### 6.2.3. Custom Strategy dengan JavaScript (Untuk Developer)
- Pengguna menulis kode JavaScript yang berisi fungsi `onBar(bar, context)`.
- **Loop backtesting dijalankan di JavaScript**:
  - JavaScript menerima data historis (array of candles) dan objek indikator (hasil perhitungan dari Go).
  - JavaScript melakukan iterasi dari baris pertama hingga terakhir, memanggil `onBar` pada setiap baris.
  - Di dalam `onBar`, pengguna dapat memanggil `context.buy()`, `context.sell()`, `context.closePosition()`, serta `context.cancelOrder()` atau `context.cancelAllOrders()`.
  - JavaScript menangani manajemen posisi, order, komisi, dan pencatatan transaksi.
  - Setelah loop selesai, JavaScript menghitung metrik performa (atau memanggil fungsi Go untuk menghitung metrik dari daftar transaksi).
- API yang tersedia dalam `context`:
  - `bar`: objek candle saat ini (`time`, `open`, `high`, `low`, `close`, `volume`).
  - `index`: indeks baris saat ini.
  - `indicators`: objek berisi fungsi indikator yang sudah dihitung (misal `indicators.sma(period)` mengembalikan array nilai SMA untuk seluruh data, atau `indicators.smaAt(index, period)` untuk nilai pada indeks tertentu).
  - `buy(price?)`, `sell(price?)`, `closePosition(price?)` – untuk order market/limit; jika parameter `price` diberikan maka menjadi limit order, jika tidak maka market order. Fungsi ini mengembalikan ID order (jika limit) yang dapat digunakan untuk pembatalan.
  - `cancelOrder(orderId)` – membatalkan limit order yang belum tereksekusi.
  - `cancelAllOrders()` – membatalkan semua limit order aktif.
  - `getPosition()` mengembalikan posisi saat ini.
  - `settings`: pengaturan backtest (modal awal, komisi market/limit, dll.).
- Strategi JavaScript dieksekusi di lingkungan sandbox (Web Worker) untuk mencegah akses DOM dan menjaga keamanan.

#### 6.2.4. Custom Indicator (Opsional)
- Pengguna dapat mendefinisikan indikator kustom sebagai fungsi JavaScript yang menerima array data dan mengembalikan array nilai.
- Indikator kustom ini dapat digunakan dalam strategi JavaScript.

### 6.3. Engine Backtesting

- **Dua mode eksekusi**:
  1. **Mode Go** (untuk strategi preset dan rule-based): loop backtest dijalankan di Go/Wasm untuk performa maksimal.
  2. **Mode JavaScript** (untuk strategi kustom): loop dijalankan di JavaScript, dengan Go menyediakan data dan indikator.
- Simulasi order:
  - Market order dan limit order.
  - **Limit order dapat dibatalkan oleh strategi** melalui API `context.cancelOrder(orderId)` atau `context.cancelAllOrders()`.
  - Order yang dibatalkan tidak dieksekusi dan tidak mempengaruhi posisi atau biaya.
  - **Komisi terpisah untuk market order dan limit order**:
    - Pengguna dapat mengatur komisi market order (dalam persen atau nilai tetap) dan komisi limit order (dalam persen atau nilai tetap) secara terpisah.
    - Contoh: komisi market 0.1%, limit 0.05%.
  - Spread dapat dikonfigurasi (untuk market order).
  - Posisi long dan short (opsional).
  - Stop loss, take profit, trailing stop (opsional, dapat diimplementasikan di kedua mode).
- Output: daftar transaksi, equity curve, drawdown, statistik.

### 6.4. Indikator Teknikal

- Implementasi di Go untuk kecepatan.
- Indikator: SMA, EMA, WMA, RSI, MACD, Bollinger Bands, Stochastic, ATR, OBV, dll.
- Perhitungan dilakukan di Go dan hasilnya diteruskan ke JavaScript (jika mode JS) atau digunakan langsung oleh engine Go.

### 6.5. Optimasi Parameter & Walk-Forward Analysis

#### 6.5.1. Grid Search
- Pengguna dapat memilih parameter strategi yang ingin dioptimasi.
- **Parameter yang dapat dioptimasi hanya bertipe numerik (integer/float) dengan rentang nilai.**
- **Tidak mendukung parameter bertipe kategorikal** (misal jenis moving average, tipe harga). Jika diperlukan, pengguna dapat membuat beberapa strategi preset terpisah.
- Menentukan rentang nilai untuk setiap parameter (misal periode fast SMA dari 5 sampai 50 step 5).
- Sistem menjalankan backtest untuk setiap kombinasi parameter.
- Hasil ditampilkan dalam bentuk tabel/heatmap yang menunjukkan metrik performa (return, Sharpe, dll.) untuk setiap kombinasi.
- Pengguna dapat mengurutkan dan memilih kombinasi terbaik.
- **Hasil optimasi dapat disimpan** (lihat 6.5.3).

#### 6.5.2. Walk-Forward Analysis (Rolling Window)
- Data dibagi menjadi beberapa jendela (window) berurutan: in-sample (training) dan out-of-sample (testing).
- **Mendukung rolling window**: setelah selesai satu siklus, window bergeser maju sebesar langkah (step) tertentu, bukan menambah data (expanding).
  - Contoh: In-sample = 6 bulan, Out-of-sample = 2 bulan, step = 2 bulan. Setiap siklus window in-sample dan out-of-sample berukuran tetap namun posisinya bergeser.
- **Mendukung optimasi multi-parameter**: grid search di dalam setiap window in-sample dapat mengoptimasi beberapa parameter sekaligus.
- Pengguna dapat memilih lebih dari satu parameter beserta rentangnya untuk optimasi.
- Parameter terbaik dari tiap window in-sample kemudian diuji pada window out-of-sample berikutnya.
- Hasil walk-forward menampilkan ringkasan performa out-of-sample gabungan.
- Pengguna dapat mengatur ukuran window in-sample, out-of-sample, dan step (rolling).
- **Hasil walk-forward dapat disimpan** (lihat 6.5.3).

#### 6.5.3. Penyimpanan, Perbandingan, Impor/Ekspor Hasil Optimasi
- Setelah grid search atau walk-forward selesai, pengguna dapat menyimpan hasilnya dengan nama tertentu.
- Hasil yang disimpan mencakup:
  - Jenis optimasi (grid search / walk-forward).
  - Parameter yang diuji dan rentangnya.
  - Metrik performa terbaik (atau ringkasan out-of-sample untuk WFA).
  - Tanggal dan waktu.
  - Data yang digunakan (simbol, timeframe, rentang tanggal).
  - Strategi yang digunakan.
- Pengguna dapat melihat daftar hasil tersimpan di tab **Saved Optimizations**.
- Fitur **perbandingan**: pengguna dapat memilih dua atau lebih hasil tersimpan untuk dibandingkan secara berdampingan (side-by-side) dalam tabel atau grafik.
- **Ekspor**: hasil optimasi dapat diekspor ke file JSON.
- **Impor**: pengguna dapat mengimpor file JSON hasil optimasi yang sebelumnya diekspor (dari aplikasi ini) untuk ditambahkan ke daftar hasil tersimpan.
- Hasil tersimpan disimpan di IndexedDB agar dapat diakses kembali.

### 6.6. Deteksi Gap Data & Peringatan

- Sebelum backtest, aplikasi memeriksa konsistensi timestamp data.
- Gap didefinisikan sebagai selisih waktu antar baris yang lebih besar dari timeframe yang diharapkan (atau jika data tidak berurutan sesuai timeframe).
- Jika gap terdeteksi, tampilkan peringatan yang menunjukkan jumlah gap, lokasi, dan durasi.
- Pengguna dapat memilih cara menangani gap:
  - Abaikan (lanjutkan backtest).
  - Forward-fill harga sebelumnya.
  - Hentikan dan minta pengguna memperbaiki data.
- Peringatan juga muncul jika data memiliki timestamp duplikat atau tidak terurut.

### 6.7. Resampling Timeframe

- Jika data yang dimuat memiliki timeframe lebih kecil (misal 1 menit) dan pengguna ingin melakukan backtest pada timeframe lebih besar (misal 1 jam), aplikasi menyediakan fitur resampling.
- Pengguna memilih timeframe target (misal 5m, 15m, 1h, 4h, 1d, 1w).
- Metode agregasi OHLC:
  - Open: harga open pertama dari periode.
  - High: harga tertinggi.
  - Low: harga terendah.
  - Close: harga close terakhir.
  - Volume: jumlah volume.
- Resampling dilakukan di sisi Go/Wasm untuk efisiensi, atau di JavaScript jika data kecil.
- Hasil resampling menggantikan data asli untuk backtest selanjutnya.

### 6.8. Hasil & Visualisasi

- Ringkasan metrik: Total Return, Annualized Return, Max Drawdown, Sharpe Ratio, Sortino Ratio, Total Trades, Win Rate, Profit Factor, Expectancy.
- Grafik: Equity Curve, Drawdown, Harga dengan sinyal entry/exit (opsional).
- Tabel transaksi.
- Export hasil ke CSV/JSON.
- Untuk optimasi: tampilan heatmap/tabel hasil grid search, ringkasan walk-forward, dan **tab "Saved Optimizations" untuk melihat, membandingkan, serta impor/ekspor hasil optimasi**.

### 6.9. UI/UX

- **Layout**:
  - **Sidebar kiri**:
    - **Data Source**: Upload, URL, Crypto Data (Binance), Sample.
    - **Data Preview**: info jumlah baris, timeframe, rentang tanggal.
    - **Mapping Columns**: tombol jika perlu.
    - **Resample**: dropdown timeframe target, tombol "Resample".
    - **Strategy**: mode pilihan (Preset/Rule/JS), parameter form, editor kode.
    - **Settings**: modal awal, komisi market, komisi limit, spread, dll.
    - **Optimization**: (collapsible) pilihan parameter grid, pengaturan walk-forward (rolling window size, step), tombol "Run Grid Search", "Run Walk-Forward", "Save Results".
    - **Run Backtest**: tombol utama.
  - **Area utama**:
    - Tabs: Summary, Equity Curve, Drawdown, Trades, Optimization Results, Saved Optimizations.
    - Kartu metrik, grafik, tabel.
    - Notifikasi gap muncul di atas area utama.
    - Tab Saved Optimizations menampilkan daftar hasil tersimpan dengan opsi perbandingan, impor, dan ekspor.
- **Loading indicator** selama backtest/optimasi.
- **Error handling** yang jelas.

### 6.10. Penyimpanan Lokal

- Simpan konfigurasi strategi, mapping, pengaturan, dan hasil backtest terakhir di localStorage/IndexedDB.
- Simpan hasil optimasi di IndexedDB dengan metadata untuk perbandingan.
- Export/import konfigurasi dan hasil.

---

## 7. Kebutuhan Non-Fungsional

### 7.1. Performa
- Backtest mode Go untuk 100.000 baris < 5 detik.
- Backtest mode JavaScript untuk 50.000 baris < 5 detik (tergantung kompleksitas strategi).
- Grid search dengan 100 kombinasi pada data 50.000 baris < 30 detik (dapat dijalankan di Web Worker agar UI tidak freeze).
- Resampling 1 juta baris < 3 detik.

### 7.2. Kompatibilitas
- Browser modern yang mendukung WebAssembly: Chrome, Firefox, Safari, Edge terbaru.
- Dukungan Web Worker.
- Tidak bergantung pada plugin.

### 7.3. Ukuran Aplikasi
- File Wasm utama < 8 MB (dengan optimasi).
- Total aset statis < 15 MB.

### 7.4. Keamanan
- Semua data diproses lokal.
- Eksekusi JavaScript pengguna di Web Worker dengan konteks terbatas (tidak ada akses DOM, `window`, `document`, `fetch`, dll., kecuali API yang disediakan).
- Tidak ada evaluasi kode yang tidak aman di Go.

### 7.5. Maintainability
- Kode Go dipisah dalam paket: `data`, `indicators`, `engine`, `wasm`.
- Frontend terstruktur modular (TypeScript disarankan).
- Dokumentasi pengembang.

---

## 8. Arsitektur Sistem
```
[ Browser ]
   |
   |--- index.html
   |--- assets/
   |      |--- css/
   |      |--- js/
   |      |--- wasm/
   |
   |--- Data Flow:
   |      |--- User upload file / load Binance data
   |      |--- Parsing (CSV/JSON/Parquet) → array of objects
   |      |--- Column mapping → normalized candles
   |      |--- Resampling (optional) → new candles
   |      |--- Gap detection → warning
   |      |
   |      |--- Strategy selection:
   |      |      |--- Preset / Rule-based → Go engine loop
   |      |      |--- Custom JS → JS loop (using Go indicators)
   |      |
   |      |--- Optimization (grid search / walk-forward rolling window):
   |      |      |--- Multiple backtests (Go or JS depending on strategy mode)
   |      |      |--- Save results to IndexedDB
   |      |      |--- Compare / Import / Export saved results
   |      |
   |      |--- Results rendering
```
### 8.1. Interaksi Go ↔ JavaScript

- **Untuk mode Go (preset/rule-based):**  
  JavaScript mengirim data (candles) dan definisi strategi ke Go melalui `runBacktest()`. Go menjalankan loop, mengembalikan hasil (trades, equity, metrics).

- **Untuk mode JavaScript (custom strategy):**  
  1. JavaScript memanggil Go untuk menghitung indikator yang diperlukan (atau semua indikator standar) berdasarkan data candles. Go mengembalikan objek berisi array indikator.
  2. JavaScript menjalankan loop backtesting sendiri, menggunakan array indikator tersebut dan memanggil fungsi `onBar` pengguna.
  3. JavaScript mengelola posisi, order, dan mencatat transaksi (dengan komisi terpisah untuk market/limit, serta pembatalan limit order).
  4. Setelah loop selesai, JavaScript menghitung metrik (atau memanggil Go untuk perhitungan metrik dari array transaksi).
  5. Hasil ditampilkan.

- **Untuk optimasi (grid search / walk-forward):**  
  Jika strategi mode Go, grid search dijalankan di Go (lebih cepat). Jika strategi mode JavaScript, grid search dijalankan di JavaScript dengan Web Worker. Walk-forward analysis mengimplementasikan rolling window dengan menggeser window secara step, dan mendukung optimasi multi-parameter.

- **Resampling:**  
  Dilakukan di Go, menerima array candles dan timeframe target, mengembalikan array candles baru.

- **Deteksi gap:**  
  Dilakukan di Go atau JavaScript, mengembalikan daftar gap.

### 8.2. API Go yang Diekspos ke JavaScript

- `parseParquet(arrayBuffer) string` – mengembalikan JSON array of objects.
- `resampleCandles(candlesJSON string, timeframe string) string` – mengembalikan JSON candles baru.
- `detectGaps(candlesJSON string, expectedIntervalMs int) string` – mengembalikan JSON daftar gap.
- `runBacktest(candlesJSON string, strategyJSON string, settingsJSON string) string` – untuk mode Go.
- `calculateIndicators(candlesJSON string, indicatorsJSON string) string` – mengembalikan JSON berisi array indikator.
- `calculateMetrics(tradesJSON string, initialCapital float64) string` – menghitung metrik dari daftar transaksi (opsional, bisa juga di JS).

---

## 9. Alur Pengguna (User Journey)

1. Buka aplikasi.
2. Pilih sumber data:
   - Upload file (CSV/JSON/Parquet), atau
   - Pilih data crypto dari Binance (simbol, timeframe, rentang tanggal), atau
   - Gunakan data contoh.
3. Aplikasi menampilkan preview data.
4. Lakukan column mapping jika perlu.
5. (Opsional) Resample data ke timeframe lebih besar.
6. Aplikasi memeriksa gap data dan menampilkan peringatan jika ada.
7. Pilih atau buat strategi (preset, rule builder, atau JavaScript editor).
8. Atur parameter backtest (modal awal, komisi market & limit, spread, dll.).
9. (Opsional) Lakukan optimasi parameter:
   - Pilih parameter yang akan dioptimasi (numerik saja).
   - Tentukan rentang nilai.
   - Jalankan grid search.
   - Lihat hasil heatmap/tabel.
   - Simpan hasil optimasi.
   - (Opsional) Jalankan walk-forward analysis dengan rolling window dan multi-parameter.
   - Simpan hasil walk-forward.
10. Klik "Run Backtest" (atau "Run Optimization").
11. Lihat hasil: ringkasan metrik, grafik, tabel transaksi.
12. Simpan/export hasil.
13. Buka tab "Saved Optimizations" untuk melihat, membandingkan, serta impor/ekspor hasil optimasi yang tersimpan.

---

## 10. Desain UI/UX (Ringkasan)

- **Sidebar kiri**:
  - **Data Source**: Upload, URL, Crypto Data (Binance), Sample.
  - **Data Preview**: info jumlah baris, timeframe, rentang tanggal.
  - **Mapping Columns**: tombol jika perlu.
  - **Resample**: dropdown timeframe target, tombol "Resample".
  - **Strategy**: mode pilihan (Preset/Rule/JS), parameter form, editor kode.
  - **Settings**: modal awal, komisi market, komisi limit, spread, dll.
  - **Optimization**: (collapsible) pilihan parameter grid, pengaturan walk-forward (rolling window size, step), tombol "Run Grid Search", "Run Walk-Forward", "Save Results".
  - **Run Backtest**: tombol utama.
- **Area utama**:
  - Tabs: Summary, Equity Curve, Drawdown, Trades, Optimization Results, Saved Optimizations.
  - Kartu metrik, grafik, tabel.
  - Notifikasi gap muncul di atas area utama.
  - Tab Saved Optimizations menampilkan daftar hasil tersimpan dengan opsi perbandingan, impor, dan ekspor.

---

## 11. Milestone & Deliverables

### Fase 1 – MVP (Core)
- Setup proyek Go, kompilasi ke Wasm.
- Parser data CSV dan JSON.
- Engine backtest Go dengan satu strategi preset (SMA crossover).
- Column mapping sederhana.
- UI minimal: upload data, mapping, form parameter, run, tampilkan hasil teks.
- Deteksi gap sederhana (peringatan).
- Sumber data Binance (API atau contoh) untuk satu simbol.
- Deploy statis.

### Fase 2 – UI & Visualisasi
- Layout responsif lengkap.
- Grafik equity curve.
- Tabel transaksi.
- Metrik performa lengkap.
- Rule-based strategy builder.
- Resampling timeframe.
- Optimasi grid search dasar (mode Go) dengan kemampuan simpan hasil.

### Fase 3 – Fitur Lanjutan
- Dukungan file Parquet.
- Custom JavaScript strategy (loop di JS) dengan editor dan sandbox, termasuk pembatalan limit order.
- Walk-forward analysis dengan rolling window dan multi-parameter.
- Konfigurasi komisi terpisah untuk market dan limit order.
- Penyimpanan, perbandingan, impor/ekspor hasil optimasi.
- Export hasil backtest.
- Pengaturan lanjutan (stop loss, take profit, trailing stop).

### Fase 4 – Optimasi & Polish
- Optimasi ukuran Wasm.
- Web Worker untuk backtest/optimasi.
- Testing lintas browser.
- Dokumentasi pengguna dan pengembang.
- Peningkatan UI/UX berdasarkan feedback.

---

## 12. Kriteria Penerimaan

- Aplikasi dapat di-deploy ke GitHub Pages tanpa konfigurasi server.
- Backtest satu simbol dengan data 10.000 baris (CSV) selesai < 3 detik (mode Go).
- Custom strategy JavaScript dapat berjalan pada 50.000 baris < 5 detik.
- File Parquet 100.000 baris dapat dimuat dan diproses.
- Resampling dari 1 menit ke 1 jam bekerja dengan benar.
- Deteksi gap menampilkan peringatan yang akurat.
- Grid search 100 kombinasi pada data 50.000 baris selesai < 30 detik.
- Walk-forward analysis rolling window berjalan dan menampilkan hasil.
- Pengguna dapat memuat data crypto dari Binance (misal BTCUSDT 1h) tanpa upload manual.
- Pengguna dapat mengatur komisi market dan limit secara terpisah, dan perhitungan biaya sesuai.
- Strategi JavaScript dapat membatalkan limit order yang belum tereksekusi.
- Hasil optimasi dapat disimpan, dibandingkan, diekspor, dan diimpor kembali.
- Tidak ada dukungan untuk data tick.
- Hasil yang ditampilkan akurat (dibandingkan dengan referensi).
- UI tidak error pada browser terbaru.

---

## 13. Risiko & Mitigasi

| Risiko | Mitigasi |
|--------|----------|
| Ukuran Wasm besar | Gunakan TinyGo atau optimasi build, hindari dependensi berat. |
| Dukungan Parquet di browser | Gunakan library `parquet-wasm` yang sudah mendukung read di browser. |
| Kinerja custom strategy (loop di JS) | Batasi ukuran data untuk mode JS, gunakan Web Worker, dan sediakan peringatan jika data besar. |
| Grid search memakan waktu lama | Jalankan di Web Worker, batasi jumlah kombinasi, tampilkan progress. |
| Sumber data Binance API mungkin berubah atau CORS | Sediakan fallback ke data contoh, atau gunakan proxy publik (jika diizinkan). |
| Kompleksitas walk-forward rolling window | Implementasikan dengan parameter yang jelas, dokumentasikan keterbatasan. |
| Keamanan eksekusi JS | Sandbox di Web Worker, tanpa akses ke objek global browser. |
| Deteksi gap yang salah | Tentukan interval expected dari data (atau minta pengguna konfirmasi timeframe), gunakan toleransi. |
| Penyimpanan hasil optimasi membengkak di IndexedDB | Batasi ukuran yang disimpan, beri opsi hapus, gunakan kompresi jika perlu. |

---

## 14. Keputusan Desain yang Telah Ditetapkan

- **Tidak mendukung parameter optimasi bertipe kategorikal** – hanya numerik.
- **Walk-forward analysis mendukung optimasi multi-parameter**.
- **Limit order dapat dibatalkan oleh strategi** melalui API yang disediakan.
- **Fitur impor hasil optimasi dari file JSON** didukung untuk membandingkan hasil dari sesi lain.
- **Sumber data crypto hanya dari Binance** (tidak ada exchange lain).
- **Tidak ada dukungan data tick** (hanya OHLCV).

---

**Dokumen ini siap digunakan sebagai panduan pengembangan. Anda dapat memberikannya kepada AI agent untuk menghasilkan task-task implementasi atau bahkan kode awal.**

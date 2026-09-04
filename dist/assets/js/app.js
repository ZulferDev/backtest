/**
 * Trading Backtesting Framework - Main Application JavaScript
 */

// Global state
let wasmReady = false;
let currentCandles = [];
let backtestResult = null;

// Timeframe interval in milliseconds
const TIMEFRAME_INTERVALS = {
    '1m': 60000,
    '3m': 180000,
    '5m': 300000,
    '15m': 900000,
    '30m': 1800000,
    '1h': 3600000,
    '2h': 7200000,
    '4h': 14400000,
    '6h': 21600000,
    '8h': 28800000,
    '12h': 43200000,
    '1d': 86400000,
    '3d': 259200000,
    '1w': 604800000,
    '1M': 2592000000
};

// Initialize application
async function init() {
    showLoading(true);
    
    try {
        await loadWasm();
        setupEventListeners();
        setupDataSourceToggle();
        
        // Set default dates for Binance
        const today = new Date();
        const oneYearAgo = new Date(today.getFullYear() - 1, today.getMonth(), today.getDate());
        document.getElementById('binanceEnd').valueAsDate = today;
        document.getElementById('binanceStart').valueAsDate = oneYearAgo;
        
    } catch (error) {
        console.error('Failed to initialize:', error);
        alert('Failed to initialize application: ' + error.message);
    } finally {
        showLoading(false);
    }
}

// Load WebAssembly module
async function loadWasm() {
    const go = new Go();
    
    try {
        const response = await fetch('assets/wasm/main.wasm');
        if (!response.ok) {
            throw new Error('Failed to load WASM file');
        }
        
        const bytes = await response.arrayBuffer();
        const result = await WebAssembly.instantiate(bytes, go.importObject);
        
        go.run(result.instance);
        wasmReady = true;
        
        console.log('WebAssembly loaded successfully');
        
        // Enable run button if data is loaded
        checkRunButton();
        
    } catch (error) {
        console.error('WASM loading failed:', error);
        throw error;
    }
}

// Setup event listeners
function setupEventListeners() {
    // Data source buttons
    document.getElementById('btnLoadFile').addEventListener('click', handleFileLoad);
    document.getElementById('btnLoadBinance').addEventListener('click', handleBinanceLoad);
    document.getElementById('btnLoadSample').addEventListener('click', handleSampleLoad);
    
    // Resample button
    document.getElementById('btnResample').addEventListener('click', handleResample);
    
    // Run backtest
    document.getElementById('btnRunBacktest').addEventListener('click', runBacktest);
    
    // Tab navigation
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', handleTabClick);
    });
}

// Setup data source toggle
function setupDataSourceToggle() {
    const dataSourceSelect = document.getElementById('dataSource');
    const uploadSection = document.getElementById('uploadSection');
    const binanceSection = document.getElementById('binanceSection');
    const sampleSection = document.getElementById('sampleSection');
    
    dataSourceSelect.addEventListener('change', () => {
        uploadSection.style.display = 'none';
        binanceSection.style.display = 'none';
        sampleSection.style.display = 'none';
        
        switch (dataSourceSelect.value) {
            case 'upload':
                uploadSection.style.display = 'block';
                break;
            case 'binance':
                binanceSection.style.display = 'block';
                break;
            case 'sample':
                sampleSection.style.display = 'block';
                break;
        }
    });
}

// Handle file upload
async function handleFileLoad() {
    const fileInput = document.getElementById('fileInput');
    const delimiterSelect = document.getElementById('delimiter');
    
    if (!fileInput.files.length) {
        alert('Please select a file');
        return;
    }
    
    const file = fileInput.files[0];
    const delimiter = delimiterSelect.value.charCodeAt(0);
    
    try {
        const text = await readFileAsText(file);
        
        let candlesJSON;
        if (file.name.endsWith('.json')) {
            candlesJSON = window.goParseJSON(text);
        } else {
            candlesJSON = window.goParseCSV(text, delimiter);
        }
        
        const result = JSON.parse(candlesJSON);
        
        if (result.error) {
            throw new Error(result.error);
        }
        
        currentCandles = result;
        onDataLoaded(currentCandles);
        
    } catch (error) {
        console.error('File load failed:', error);
        alert('Failed to load file: ' + error.message);
    }
}

// Handle Binance data load
async function handleBinanceLoad() {
    const symbol = document.getElementById('binanceSymbol').value;
    const interval = document.getElementById('binanceInterval').value;
    const startDate = document.getElementById('binanceStart').value;
    const endDate = document.getElementById('binanceEnd').value;
    
    if (!startDate || !endDate) {
        alert('Please select start and end dates');
        return;
    }
    
    try {
        const startTime = new Date(startDate).getTime();
        const endTime = new Date(endDate).getTime();
        const intervalMs = TIMEFRAME_INTERVALS[interval];
        
        // Fetch from Binance API
        const url = `https://api.binance.com/api/v3/klines?symbol=${symbol}&interval=${interval}&startTime=${startTime}&endTime=${endTime}&limit=1000`;
        
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error('Failed to fetch data from Binance');
        }
        
        const klines = await response.json();
        
        // Convert to candle format
        const candles = klines.map(k => ({
            time: new Date(k[0]).toISOString(),
            open: parseFloat(k[1]),
            high: parseFloat(k[2]),
            low: parseFloat(k[3]),
            close: parseFloat(k[4]),
            volume: parseFloat(k[5])
        }));
        
        // Convert to JSON for Go parsing
        const candlesJSON = JSON.stringify(candles);
        const parsedCandles = window.goParseJSON(candlesJSON);
        const result = JSON.parse(parsedCandles);
        
        if (result.error) {
            throw new Error(result.error);
        }
        
        currentCandles = result;
        onDataLoaded(currentCandles);
        
    } catch (error) {
        console.error('Binance load failed:', error);
        alert('Failed to load Binance data: ' + error.message);
    }
}

// Handle sample data generation
function handleSampleLoad() {
    // Generate sample OHLCV data
    const candles = generateSampleData(500);
    currentCandles = candles;
    onDataLoaded(currentCandles);
}

// Generate sample data for testing
function generateSampleData(numBars) {
    const candles = [];
    let price = 100;
    const now = Date.now();
    const interval = TIMEFRAME_INTERVALS['1h'];
    
    for (let i = 0; i < numBars; i++) {
        const time = new Date(now - (numBars - i) * interval);
        const change = (Math.random() - 0.5) * 2;
        const open = price;
        const close = price + change;
        const high = Math.max(open, close) + Math.random();
        const low = Math.min(open, close) - Math.random();
        const volume = Math.random() * 1000 + 100;
        
        candles.push({
            time: time.toISOString(),
            open: parseFloat(open.toFixed(2)),
            high: parseFloat(high.toFixed(2)),
            low: parseFloat(low.toFixed(2)),
            close: parseFloat(close.toFixed(2)),
            volume: parseFloat(volume.toFixed(2))
        });
        
        price = close;
    }
    
    return candles;
}

// Read file as text
function readFileAsText(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = e => resolve(e.target.result);
        reader.onerror = e => reject(e);
        reader.readAsText(file);
    });
}

// Handle data loaded
function onDataLoaded(candles) {
    console.log(`Loaded ${candles.length} candles`);
    
    // Show preview section
    document.getElementById('dataPreviewSection').style.display = 'block';
    
    // Display data info
    const firstCandle = candles[0];
    const lastCandle = candles[candles.length - 1];
    
    document.getElementById('dataInfo').innerHTML = `
        <p><strong>Candles:</strong> ${candles.length}</p>
        <p><strong>From:</strong> ${new Date(firstCandle.time).toLocaleDateString()}</p>
        <p><strong>To:</strong> ${new Date(lastCandle.time).toLocaleDateString()}</p>
        <p><strong>Price Range:</strong> ${getPriceRange(candles)}</p>
    `;
    
    // Check for gaps
    checkForGaps(candles);
    
    // Enable run button
    checkRunButton();
}

// Get price range string
function getPriceRange(candles) {
    let min = Infinity, max = -Infinity;
    for (const c of candles) {
        min = Math.min(min, c.low);
        max = Math.max(max, c.high);
    }
    return `${min.toFixed(2)} - ${max.toFixed(2)}`;
}

// Check for gaps in data
function checkForGaps(candles) {
    const candlesJSON = JSON.stringify(candles);
    
    // Estimate interval from first two candles
    let expectedInterval = TIMEFRAME_INTERVALS['1h']; // Default
    if (candles.length >= 2) {
        const t1 = new Date(candles[0].time).getTime();
        const t2 = new Date(candles[1].time).getTime();
        expectedInterval = t2 - t1;
    }
    
    const gapsJSON = window.goDetectGaps(candlesJSON, expectedInterval);
    const gaps = JSON.parse(gapsJSON);
    
    const warningEl = document.getElementById('gapWarning');
    
    if (gaps && gaps.length > 0) {
        warningEl.style.display = 'block';
        warningEl.innerHTML = `
            <strong>⚠️ Data Gaps Detected:</strong> Found ${gaps.length} gap(s) in the data.
            This may affect backtest results.
        `;
    } else {
        warningEl.style.display = 'none';
    }
}

// Check if run button should be enabled
function checkRunButton() {
    const btn = document.getElementById('btnRunBacktest');
    btn.disabled = !wasmReady || currentCandles.length === 0;
}

// Handle resample
function handleResample() {
    if (currentCandles.length === 0) {
        alert('Please load data first');
        return;
    }
    
    const timeframe = document.getElementById('resampleTimeframe').value;
    
    try {
        const candlesJSON = JSON.stringify(currentCandles);
        const resampledJSON = window.goResampleCandles(candlesJSON, timeframe);
        const resampled = JSON.parse(resampledJSON);
        
        if (resampled.error) {
            throw new Error(resampled.error);
        }
        
        console.log(`Resampled from ${currentCandles.length} to ${resampled.length} candles (${timeframe})`);
        
        currentCandles = resampled;
        onDataLoaded(currentCandles);
        
        alert(`Successfully resampled data to ${timeframe}. Original: ${document.getElementById('dataInfo').innerHTML.match(/Candles:\s*(\d+)/)?.[1] || '?'} → New: ${currentCandles.length} candles`);
        
    } catch (error) {
        console.error('Resample failed:', error);
        alert('Resample failed: ' + error.message);
    }
}

// Run backtest
function runBacktest() {
    if (!wasmReady || currentCandles.length === 0) {
        alert('Please load data first');
        return;
    }
    
    showLoading(true);
    
    try {
        // Get settings from UI
        const settings = {
            initialCapital: parseFloat(document.getElementById('initialCapital').value),
            commissionMarket: parseFloat(document.getElementById('commissionMarket').value),
            commissionLimit: parseFloat(document.getElementById('commissionLimit').value),
            slippage: parseFloat(document.getElementById('slippage').value),
            fastPeriod: parseInt(document.getElementById('fastPeriod').value),
            slowPeriod: parseInt(document.getElementById('slowPeriod').value)
        };
        
        const candlesJSON = JSON.stringify(currentCandles);
        const strategyJSON = JSON.stringify({ type: 'sma_crossover' });
        const settingsJSON = JSON.stringify(settings);
        
        // Call Go backtest function
        const resultJSON = window.goRunBacktest(candlesJSON, strategyJSON, settingsJSON);
        const result = JSON.parse(resultJSON);
        
        if (result.error) {
            throw new Error(result.error);
        }
        
        backtestResult = result;
        displayResults(result);
        
    } catch (error) {
        console.error('Backtest failed:', error);
        alert('Backtest failed: ' + error.message);
    } finally {
        showLoading(false);
    }
}

// Display backtest results
function displayResults(result) {
    displayMetrics(result.metrics);
    displayEquityCurve(result.equityCurve);
    displayDrawdown(result.drawdowns);
    displayTrades(result.trades);
}

// Display metrics cards
function displayMetrics(metrics) {
    const container = document.getElementById('metricsCards');
    
    const metricCards = [
        { label: 'Total Return', value: formatPercent(metrics.totalReturn), positive: metrics.totalReturn >= 0 },
        { label: 'Max Drawdown', value: formatPercent(-metrics.maxDrawdown), positive: false },
        { label: 'Sharpe Ratio', value: metrics.sharpeRatio.toFixed(2), positive: metrics.sharpeRatio > 0 },
        { label: 'Total Trades', value: metrics.totalTrades.toString(), positive: true },
        { label: 'Win Rate', value: formatPercent(metrics.winRate), positive: metrics.winRate > 50 },
        { label: 'Profit Factor', value: metrics.profitFactor.toFixed(2), positive: metrics.profitFactor > 1 },
        { label: 'Expectancy', value: formatNumber(metrics.expectancy), positive: metrics.expectancy >= 0 },
        { label: 'Avg Win', value: formatNumber(metrics.avgWin), positive: true },
        { label: 'Avg Loss', value: formatNumber(-metrics.avgLoss), positive: false }
    ];
    
    container.innerHTML = metricCards.map(m => `
        <div class="metric-card ${m.positive ? 'positive' : 'negative'}">
            <h3>${m.label}</h3>
            <div class="value">${m.value}</div>
        </div>
    `).join('');
}

// Format percentage
function formatPercent(value) {
    if (value === undefined || value === null) return 'N/A';
    return (value >= 0 ? '+' : '') + value.toFixed(2) + '%';
}

// Format number
function formatNumber(value) {
    if (value === undefined || value === null) return 'N/A';
    return value.toFixed(2);
}

// Display equity curve chart
function displayEquityCurve(equityCurve) {
    const canvas = document.getElementById('equityChart');
    const ctx = canvas.getContext('2d');
    
    // Set canvas size
    canvas.width = canvas.offsetWidth;
    canvas.height = 400;
    
    drawLineChart(ctx, equityCurve, '#3498db', 'Equity Curve');
}

// Display drawdown chart
function displayDrawdown(drawdowns) {
    const canvas = document.getElementById('drawdownChart');
    const ctx = canvas.getContext('2d');
    
    // Set canvas size
    canvas.width = canvas.offsetWidth;
    canvas.height = 400;
    
    drawLineChart(ctx, drawdowns, '#e74c3c', 'Drawdown %');
}

// Draw line chart
function drawLineChart(ctx, data, color, label) {
    const width = ctx.canvas.width;
    const height = ctx.canvas.height;
    const padding = 40;
    
    // Clear canvas
    ctx.clearRect(0, 0, width, height);
    
    if (!data || data.length === 0) return;
    
    // Find min/max
    let min = Math.min(...data);
    let max = Math.max(...data);
    
    // Add padding to y-axis
    const range = max - min || 1;
    min -= range * 0.1;
    max += range * 0.1;
    
    // Draw axes
    ctx.strokeStyle = '#e0e0e0';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(padding, padding);
    ctx.lineTo(padding, height - padding);
    ctx.lineTo(width - padding, height - padding);
    ctx.stroke();
    
    // Draw data line
    ctx.strokeStyle = color;
    ctx.lineWidth = 2;
    ctx.beginPath();
    
    const xStep = (width - 2 * padding) / (data.length - 1);
    
    data.forEach((value, i) => {
        const x = padding + i * xStep;
        const y = height - padding - ((value - min) / (max - min)) * (height - 2 * padding);
        
        if (i === 0) {
            ctx.moveTo(x, y);
        } else {
            ctx.lineTo(x, y);
        }
    });
    
    ctx.stroke();
    
    // Draw label
    ctx.fillStyle = '#333';
    ctx.font = '12px sans-serif';
    ctx.fillText(label, padding, height - 10);
}

// Display trades table
function displayTrades(trades) {
    const tbody = document.querySelector('#tradesTable tbody');
    
    if (!trades || trades.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6">No trades executed</td></tr>';
        return;
    }
    
    tbody.innerHTML = trades.map(trade => `
        <tr>
            <td>${new Date(trade.time).toLocaleString()}</td>
            <td style="color: ${trade.type === 'buy' ? '#27ae60' : '#e74c3c'}">${trade.type.toUpperCase()}</td>
            <td>${trade.price.toFixed(2)}</td>
            <td>${trade.quantity.toFixed(4)}</td>
            <td>${trade.commission.toFixed(4)}</td>
            <td style="color: ${trade.pnl >= 0 ? '#27ae60' : '#e74c3c'}">${trade.pnl.toFixed(2)}</td>
        </tr>
    `).join('');
}

// Handle tab click
function handleTabClick(e) {
    const tabId = e.target.dataset.tab;
    
    // Update active tab button
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    e.target.classList.add('active');
    
    // Show corresponding content
    document.querySelectorAll('.tab-content').forEach(content => {
        content.style.display = 'none';
    });
    document.getElementById(`tab-${tabId}`).style.display = 'block';
    
    // Redraw charts if needed (for proper sizing)
    if (tabId === 'equity' && backtestResult) {
        displayEquityCurve(backtestResult.equityCurve);
    } else if (tabId === 'drawdown' && backtestResult) {
        displayDrawdown(backtestResult.drawdowns);
    }
}

// Show/hide loading indicator
function showLoading(show) {
    document.getElementById('loadingIndicator').style.display = show ? 'flex' : 'none';
}

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', init);

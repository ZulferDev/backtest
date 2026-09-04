package data

import (
	"testing"
	"time"
)

func TestParseCSV(t *testing.T) {
	csvData := `timestamp,open,high,low,close,volume
1704067200000,42000.50,42500.00,41800.00,42300.00,1000.5
1704070800000,42300.00,42800.00,42100.00,42600.00,1200.0
1704074400000,42600.00,43000.00,42400.00,42900.00,1100.25`

	candles, err := ParseCSV(csvData, ',')
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(candles) != 3 {
		t.Errorf("Expected 3 candles, got %d", len(candles))
	}

	// Check first candle
	if candles[0].Open != 42000.50 {
		t.Errorf("Expected open 42000.50, got %f", candles[0].Open)
	}
	if candles[0].Close != 42300.00 {
		t.Errorf("Expected close 42300.00, got %f", candles[0].Close)
	}
	if candles[0].Volume != 1000.5 {
		t.Errorf("Expected volume 1000.5, got %f", candles[0].Volume)
	}
}

func TestParseCSVSemicolonDelimiter(t *testing.T) {
	csvData := `timestamp;open;high;low;close;volume
1704067200000;42000.50;42500.00;41800.00;42300.00;1000.5
1704070800000;42300.00;42800.00;42100.00;42600.00;1200.0`

	candles, err := ParseCSV(csvData, ';')
	if err != nil {
		t.Fatalf("ParseCSV with semicolon failed: %v", err)
	}

	if len(candles) != 2 {
		t.Errorf("Expected 2 candles, got %d", len(candles))
	}
}

func TestParseJSON(t *testing.T) {
	jsonData := `[
		{"time": "1704067200000", "open": 42000.50, "high": 42500.00, "low": 41800.00, "close": 42300.00, "volume": 1000.5},
		{"time": "1704070800000", "open": 42300.00, "high": 42800.00, "low": 42100.00, "close": 42600.00, "volume": 1200.0}
	]`

	candles, err := ParseJSON(jsonData)
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}

	if len(candles) != 2 {
		t.Errorf("Expected 2 candles, got %d", len(candles))
	}

	if candles[0].Open != 42000.50 {
		t.Errorf("Expected open 42000.50, got %f", candles[0].Open)
	}
}

func TestDetectColumnMapping(t *testing.T) {
	headers := []string{"open_time", "open", "high", "low", "close", "volume"}
	mapping := detectColumnMapping(headers)

	if mapping.Timestamp != "open_time" {
		t.Errorf("Expected timestamp 'open_time', got '%s'", mapping.Timestamp)
	}
	if mapping.Open != "open" {
		t.Errorf("Expected open 'open', got '%s'", mapping.Open)
	}
	if mapping.Volume != "volume" {
		t.Errorf("Expected volume 'volume', got '%s'", mapping.Volume)
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		input    string
		expected int64 // expected Unix millisecond
	}{
		{"1704067200000", 1704067200000},
		{"1704067200", 1704067200000},
		{"2024-01-01T00:00:00Z", 1704067200000},
	}

	for _, tt := range tests {
		result, err := parseTimestamp(tt.input)
		if err != nil {
			t.Errorf("parseTimestamp(%s) error: %v", tt.input, err)
			continue
		}

		resultMs := result.UnixMilli()
		// Allow some tolerance for timezone differences
		diff := resultMs - tt.expected
		if diff < 0 {
			diff = -diff
		}
		if diff > 3600000 { // 1 hour tolerance
			t.Errorf("parseTimestamp(%s) = %d, expected %d", tt.input, resultMs, tt.expected)
		}
	}
}

func TestDetectGaps(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	hour := 3600000 // 1 hour in milliseconds

	candles := []Candle{
		{Time: baseTime, Open: 42000, High: 42100, Low: 41900, Close: 42050, Volume: 100},
		{Time: baseTime.Add(time.Duration(hour) * time.Millisecond), Open: 42050, High: 42150, Low: 42000, Close: 42100, Volume: 100},
		{Time: baseTime.Add(time.Duration(hour*5) * time.Millisecond), Open: 42100, High: 42200, Low: 42050, Close: 42150, Volume: 100}, // Gap here
	}

	gaps, err := DetectGaps(candles, hour)
	if err != nil {
		t.Fatalf("DetectGaps failed: %v", err)
	}

	if len(gaps) != 1 {
		t.Errorf("Expected 1 gap, got %d", len(gaps))
	} else {
		expectedDuration := int64(hour * 4) // 4 hours gap
		actualDuration := gaps[0].Duration
		if actualDuration != expectedDuration {
			t.Errorf("Expected gap duration %d ms, got %d ms", expectedDuration, actualDuration)
		}
	}
}

func TestDetectGapsNoGap(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	hour := 3600000

	candles := []Candle{
		{Time: baseTime, Open: 42000, Close: 42050, High: 42100, Low: 41900, Volume: 100},
		{Time: baseTime.Add(time.Duration(hour) * time.Millisecond), Open: 42050, Close: 42100, High: 42150, Low: 42000, Volume: 100},
		{Time: baseTime.Add(time.Duration(hour*2) * time.Millisecond), Open: 42100, Close: 42150, High: 42200, Low: 42050, Volume: 100},
	}

	gaps, err := DetectGaps(candles, hour)
	if err != nil {
		t.Fatalf("DetectGaps failed: %v", err)
	}

	if len(gaps) != 0 {
		t.Errorf("Expected 0 gaps, got %d", len(gaps))
	}
}

func TestHasDuplicateTimestamps(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	candlesWithDup := []Candle{
		{Time: baseTime, Open: 42000, Close: 42050, High: 42100, Low: 41900, Volume: 100},
		{Time: baseTime.Add(time.Hour), Open: 42050, Close: 42100, High: 42150, Low: 42000, Volume: 100},
		{Time: baseTime, Open: 42100, Close: 42150, High: 42200, Low: 42050, Volume: 100}, // Duplicate
	}

	if !HasDuplicateTimestamps(candlesWithDup) {
		t.Error("Expected duplicate timestamps detected")
	}

	candlesNoDup := []Candle{
		{Time: baseTime, Open: 42000, Close: 42050, High: 42100, Low: 41900, Volume: 100},
		{Time: baseTime.Add(time.Hour), Open: 42050, Close: 42100, High: 42150, Low: 42000, Volume: 100},
		{Time: baseTime.Add(time.Hour * 2), Open: 42100, Close: 42150, High: 42200, Low: 42050, Volume: 100},
	}

	if HasDuplicateTimestamps(candlesNoDup) {
		t.Error("Did not expect duplicate timestamps")
	}
}

func TestIsSortedByTime(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	sorted := []Candle{
		{Time: baseTime},
		{Time: baseTime.Add(time.Hour)},
		{Time: baseTime.Add(time.Hour * 2)},
	}

	if !IsSortedByTime(sorted) {
		t.Error("Expected sorted candles to be detected as sorted")
	}

	unsorted := []Candle{
		{Time: baseTime},
		{Time: baseTime.Add(time.Hour * 2)},
		{Time: baseTime.Add(time.Hour)},
	}

	if IsSortedByTime(unsorted) {
		t.Error("Expected unsorted candles to be detected as unsorted")
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"42000.50", 42000.50},
		{"42000,50", 42000.50}, // European decimal
		{"1000", 1000.0},
		{"0.001", 0.001},
	}

	for _, tt := range tests {
		result, err := parseFloat(tt.input)
		if err != nil {
			t.Errorf("parseFloat(%s) error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("parseFloat(%s) = %f, expected %f", tt.input, result, tt.expected)
		}
	}
}

func TestResampleCandles(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Create 4 hourly candles
	candles := []Candle{
		{Time: baseTime, Open: 42000, High: 42100, Low: 41900, Close: 42050, Volume: 100},
		{Time: baseTime.Add(time.Hour), Open: 42050, High: 42150, Low: 42000, Close: 42100, Volume: 150},
		{Time: baseTime.Add(time.Hour * 2), Open: 42100, High: 42200, Low: 42050, Close: 42150, Volume: 200},
		{Time: baseTime.Add(time.Hour * 3), Open: 42150, High: 42250, Low: 42100, Close: 42200, Volume: 180},
	}

	// Resample to 2-hour timeframe
	resampled, err := ResampleCandles(candles, "2h")
	if err != nil {
		t.Fatalf("ResampleCandles failed: %v", err)
	}

	if len(resampled) != 2 {
		t.Errorf("Expected 2 resampled candles, got %d", len(resampled))
	} else {
		// First 2-hour candle should have:
		// Open = 42000 (first open)
		// High = 42150 (max of 42100, 42150)
		// Low = 41900 (min of 41900, 42000)
		// Close = 42100 (second close)
		// Volume = 250 (100 + 150)
		if resampled[0].Open != 42000 {
			t.Errorf("Expected first open 42000, got %f", resampled[0].Open)
		}
		if resampled[0].High != 42150 {
			t.Errorf("Expected first high 42150, got %f", resampled[0].High)
		}
		if resampled[0].Low != 41900 {
			t.Errorf("Expected first low 41900, got %f", resampled[0].Low)
		}
		if resampled[0].Close != 42100 {
			t.Errorf("Expected first close 42100, got %f", resampled[0].Close)
		}
		if resampled[0].Volume != 250 {
			t.Errorf("Expected first volume 250, got %f", resampled[0].Volume)
		}

		// Second 2-hour candle
		if resampled[1].Open != 42100 {
			t.Errorf("Expected second open 42100, got %f", resampled[1].Open)
		}
		if resampled[1].Volume != 380 {
			t.Errorf("Expected second volume 380, got %f", resampled[1].Volume)
		}
	}
}

func TestResampleCandlesEmpty(t *testing.T) {
	candles := []Candle{}

	resampled, err := ResampleCandles(candles, "1h")
	if err != nil {
		t.Fatalf("ResampleCandles with empty data failed: %v", err)
	}

	if len(resampled) != 0 {
		t.Errorf("Expected 0 resampled candles for empty input, got %d", len(resampled))
	}
}

func TestResampleCandlesDaily(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Create 3 daily candles
	candles := []Candle{
		{Time: baseTime, Open: 42000, High: 43000, Low: 41000, Close: 42500, Volume: 1000},
		{Time: baseTime.AddDate(0, 0, 1), Open: 42500, High: 43500, Low: 42000, Close: 43000, Volume: 1200},
		{Time: baseTime.AddDate(0, 0, 2), Open: 43000, High: 44000, Low: 42500, Close: 43500, Volume: 1100},
	}

	// Resample to weekly timeframe
	resampled, err := ResampleCandles(candles, "1w")
	if err != nil {
		t.Fatalf("ResampleCandles failed: %v", err)
	}

	// All 3 days should be in the same week bucket
	if len(resampled) != 1 {
		t.Errorf("Expected 1 weekly candle, got %d", len(resampled))
	} else {
		if resampled[0].Open != 42000 {
			t.Errorf("Expected weekly open 42000, got %f", resampled[0].Open)
		}
		if resampled[0].High != 44000 {
			t.Errorf("Expected weekly high 44000, got %f", resampled[0].High)
		}
		if resampled[0].Low != 41000 {
			t.Errorf("Expected weekly low 41000, got %f", resampled[0].Low)
		}
		if resampled[0].Close != 43500 {
			t.Errorf("Expected weekly close 43500, got %f", resampled[0].Close)
		}
		if resampled[0].Volume != 3300 {
			t.Errorf("Expected weekly volume 3300, got %f", resampled[0].Volume)
		}
	}
}

func TestTimeframeToMs(t *testing.T) {
	tests := []struct {
		timeframe string
		expected  int64
		hasError  bool
	}{
		{"1m", 60000, false},
		{"5m", 300000, false},
		{"1h", 3600000, false},
		{"1d", 86400000, false},
		{"1w", 604800000, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		result, err := timeframeToMs(tt.timeframe)
		if tt.hasError && err == nil {
			t.Errorf("timeframeToMs(%s) expected error, got none", tt.timeframe)
		}
		if !tt.hasError && err != nil {
			t.Errorf("timeframeToMs(%s) unexpected error: %v", tt.timeframe, err)
		}
		if result != tt.expected {
			t.Errorf("timeframeToMs(%s) = %d, expected %d", tt.timeframe, result, tt.expected)
		}
	}
}

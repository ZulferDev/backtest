package data

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Candle represents a single OHLCV candle
type Candle struct {
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume float64   `json:"volume"`
}

// ColumnMapping defines the mapping from CSV/JSON columns to Candle fields
type ColumnMapping struct {
	Timestamp string `json:"timestamp"`
	Open      string `json:"open"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Close     string `json:"close"`
	Volume    string `json:"volume"`
}

// Gap represents a time gap in the data
type Gap struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Duration  int64     `json:"duration"` // in milliseconds
}

// ParseCSV parses CSV data into Candle slice
// delimiter can be ',', ';', '\t', or custom character
func ParseCSV(data string, delimiter rune) ([]Candle, error) {
	reader := csv.NewReader(strings.NewReader(data))
	reader.Comma = delimiter
	reader.TrimLeadingSpace = true

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Auto-detect column mapping
	mapping := detectColumnMapping(headers)

	var candles []Candle
	lineNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV line %d: %w", lineNum, err)
		}

		candle, err := mapRecordToCandle(record, headers, mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}

		candles = append(candles, candle)
		lineNum++
	}

	return candles, nil
}

// ParseJSON parses JSON array of objects into Candle slice
func ParseJSON(data string) ([]Candle, error) {
	var rawData []map[string]interface{}

	err := json.Unmarshal([]byte(data), &rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Auto-detect column mapping from first record
	var mapping ColumnMapping
	if len(rawData) > 0 {
		headers := make([]string, 0, len(rawData[0]))
		for k := range rawData[0] {
			headers = append(headers, k)
		}
		mapping = detectColumnMapping(headers)
	}

	var candles []Candle
	for i, record := range rawData {
		candle, err := mapJSONObjectToCandle(record, mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to parse record %d: %w", i, err)
		}
		candles = append(candles, candle)
	}

	return candles, nil
}

// detectColumnMapping auto-detects column names from headers
func detectColumnMapping(headers []string) ColumnMapping {
	mapping := ColumnMapping{}

	for _, h := range headers {
		hLower := strings.ToLower(strings.TrimSpace(h))

		// Timestamp detection (priority for time-related columns)
		if mapping.Timestamp == "" && matchesAny(hLower, []string{"time", "timestamp", "date", "datetime", "ts", "open_time"}) {
			mapping.Timestamp = h
			continue
		}

		// Open detection (exclude time-related)
		if mapping.Open == "" && matchesAny(hLower, []string{"^open$", "^o$", "price_open"}) {
			mapping.Open = h
			continue
		}

		// High detection
		if mapping.High == "" && matchesAny(hLower, []string{"^high$", "^h$", "price_high", "max"}) {
			mapping.High = h
			continue
		}

		// Low detection
		if mapping.Low == "" && matchesAny(hLower, []string{"^low$", "^l$", "price_low", "min"}) {
			mapping.Low = h
			continue
		}

		// Close detection
		if mapping.Close == "" && matchesAny(hLower, []string{"^close$", "^c$", "price_close", "price"}) {
			mapping.Close = h
			continue
		}

		// Volume detection
		if mapping.Volume == "" && matchesAny(hLower, []string{"volume", "vol", "v", "qty", "quantity"}) {
			mapping.Volume = h
			continue
		}
	}

	return mapping
}

// matchesAny checks if str matches any of the patterns (simple substring match)
func matchesAny(str string, patterns []string) bool {
	for _, p := range patterns {
		// Handle regex-like patterns with ^ and $
		if len(p) >= 2 && p[0] == '^' && p[len(p)-1] == '$' {
			// Exact match pattern
			exact := p[1 : len(p)-1]
			if str == exact {
				return true
			}
		} else if str == p || strings.Contains(str, p) {
			return true
		}
	}
	return false
}

// mapRecordToCandle maps CSV record to Candle using headers and mapping
func mapRecordToCandle(record, headers []string, mapping ColumnMapping) (Candle, error) {
	data := make(map[string]string)
	for i, h := range headers {
		if i < len(record) {
			data[h] = record[i]
		}
	}

	return mapStringMapToCandle(data, mapping)
}

// mapJSONObjectToCandle maps JSON object to Candle
func mapJSONObjectToCandle(record map[string]interface{}, mapping ColumnMapping) (Candle, error) {
	data := make(map[string]string)
	for k, v := range record {
		switch val := v.(type) {
		case string:
			data[k] = val
		case float64:
			data[k] = fmt.Sprintf("%v", val)
		case int:
			data[k] = fmt.Sprintf("%d", val)
		default:
			data[k] = fmt.Sprintf("%v", val)
		}
	}

	return mapStringMapToCandle(data, mapping)
}

// mapStringMapToCandle converts string map to Candle
func mapStringMapToCandle(data map[string]string, mapping ColumnMapping) (Candle, error) {
	candle := Candle{}
	var err error

	// Parse timestamp
	if mapping.Timestamp != "" {
		candle.Time, err = parseTimestamp(data[mapping.Timestamp])
		if err != nil {
			return candle, fmt.Errorf("invalid timestamp '%s': %w", data[mapping.Timestamp], err)
		}
	}

	// Parse OHLCV
	if mapping.Open != "" {
		candle.Open, err = parseFloat(data[mapping.Open])
		if err != nil {
			return candle, fmt.Errorf("invalid open price: %w", err)
		}
	}

	if mapping.High != "" {
		candle.High, err = parseFloat(data[mapping.High])
		if err != nil {
			return candle, fmt.Errorf("invalid high price: %w", err)
		}
	}

	if mapping.Low != "" {
		candle.Low, err = parseFloat(data[mapping.Low])
		if err != nil {
			return candle, fmt.Errorf("invalid low price: %w", err)
		}
	}

	if mapping.Close != "" {
		candle.Close, err = parseFloat(data[mapping.Close])
		if err != nil {
			return candle, fmt.Errorf("invalid close price: %w", err)
		}
	}

	if mapping.Volume != "" {
		candle.Volume, err = parseFloat(data[mapping.Volume])
		if err != nil {
			return candle, fmt.Errorf("invalid volume: %w", err)
		}
	}

	return candle, nil
}

// parseTimestamp tries multiple timestamp formats
func parseTimestamp(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	// Try parsing as Unix timestamp (milliseconds)
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		// If timestamp is in milliseconds (like Binance)
		if ts > 1e12 {
			return time.UnixMilli(ts), nil
		}
		// If timestamp is in seconds
		return time.Unix(ts, 0), nil
	}

	// Try common date formats
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02/01/2006 15:04:05",
		"01/02/2006 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized timestamp format")
}

// parseFloat parses a string to float64
func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}

// DetectGaps finds gaps in the candle data
// expectedIntervalMs is the expected interval between candles in milliseconds
func DetectGaps(candles []Candle, expectedIntervalMs int) ([]Gap, error) {
	if len(candles) < 2 {
		return []Gap{}, nil
	}

	// Sort candles by time
	sorted := make([]Candle, len(candles))
	copy(sorted, candles)
	for i := 0; i < len(sorted)-1; i++ {
		if sorted[i].Time.After(sorted[i+1].Time) {
			sorted[i], sorted[i+1] = sorted[i+1], sorted[i]
		}
	}

	var gaps []Gap
	expectedDuration := time.Duration(expectedIntervalMs) * time.Millisecond
	tolerance := expectedDuration / 2 // 50% tolerance

	for i := 0; i < len(sorted)-1; i++ {
		interval := sorted[i+1].Time.Sub(sorted[i].Time)

		// Check for gap (interval significantly larger than expected)
		if interval > expectedDuration+tolerance {
			gap := Gap{
				StartTime: sorted[i].Time,
				EndTime:   sorted[i+1].Time,
				Duration:  int64(interval.Milliseconds()),
			}
			gaps = append(gaps, gap)
		}

		// Check for duplicate or out-of-order timestamps
		if interval <= 0 {
			return nil, fmt.Errorf("duplicate or out-of-order timestamp detected at index %d", i)
		}
	}

	return gaps, nil
}

// HasDuplicateTimestamps checks if there are duplicate timestamps
func HasDuplicateTimestamps(candles []Candle) bool {
	seen := make(map[int64]bool)
	for _, c := range candles {
		ts := c.Time.UnixMilli()
		if seen[ts] {
			return true
		}
		seen[ts] = true
	}
	return false
}

// IsSortedByTime checks if candles are sorted by time
func IsSortedByTime(candles []Candle) bool {
	for i := 0; i < len(candles)-1; i++ {
		if !candles[i].Time.Before(candles[i+1].Time) {
			return false
		}
	}
	return true
}

// ResampleCandles aggregates candles to a higher timeframe
// targetTimeframe can be: 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M
func ResampleCandles(candles []Candle, targetTimeframe string) ([]Candle, error) {
	if len(candles) == 0 {
		return []Candle{}, nil
	}

	targetMs, err := timeframeToMs(targetTimeframe)
	if err != nil {
		return nil, err
	}

	// Sort candles by time
	sorted := make([]Candle, len(candles))
	copy(sorted, candles)
	for i := 0; i < len(sorted)-1; i++ {
		if sorted[i].Time.After(sorted[i+1].Time) {
			sorted[i], sorted[i+1] = sorted[i+1], sorted[i]
		}
	}

	var resampled []Candle
	var current *Candle

	for _, candle := range sorted {
		// Get the start of the target timeframe bucket
		bucketStart := getBucketStart(candle.Time, targetMs)

		if current == nil {
			// Start new bucket
			current = &Candle{
				Time:   time.UnixMilli(bucketStart.UnixMilli()),
				Open:   candle.Open,
				High:   candle.High,
				Low:    candle.Low,
				Close:  candle.Close,
				Volume: candle.Volume,
			}
		} else if candle.Time.UnixMilli() >= (current.Time.UnixMilli() + targetMs) {
			// Save current bucket and start new one
			resampled = append(resampled, *current)
			current = &Candle{
				Time:   time.UnixMilli(bucketStart.UnixMilli()),
				Open:   candle.Open,
				High:   candle.High,
				Low:    candle.Low,
				Close:  candle.Close,
				Volume: candle.Volume,
			}
		} else {
			// Aggregate into current bucket
			if candle.High > current.High {
				current.High = candle.High
			}
			if candle.Low < current.Low {
				current.Low = candle.Low
			}
			current.Close = candle.Close
			current.Volume += candle.Volume
		}
	}

	// Don't forget the last bucket
	if current != nil {
		resampled = append(resampled, *current)
	}

	return resampled, nil
}

// timeframeToMs converts timeframe string to milliseconds
func timeframeToMs(tf string) (int64, error) {
	intervals := map[string]int64{
		"1m":  60000,
		"3m":  180000,
		"5m":  300000,
		"15m": 900000,
		"30m": 1800000,
		"1h":  3600000,
		"2h":  7200000,
		"4h":  14400000,
		"6h":  21600000,
		"8h":  28800000,
		"12h": 43200000,
		"1d":  86400000,
		"3d":  259200000,
		"1w":  604800000,
		"1M":  2592000000, // Approximate
	}

	if ms, ok := intervals[tf]; ok {
		return ms, nil
	}
	return 0, fmt.Errorf("unknown timeframe: %s", tf)
}

// getBucketStart returns the start of the timeframe bucket for a given time
func getBucketStart(t time.Time, intervalMs int64) time.Time {
	ms := t.UnixMilli()
	bucketStart := (ms / intervalMs) * intervalMs
	return time.UnixMilli(bucketStart)
}

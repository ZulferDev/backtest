package indicators

// SMA calculates Simple Moving Average
func SMA(values []float64, period int) []float64 {
	if period <= 0 || len(values) == 0 {
		return []float64{}
	}

	result := make([]float64, len(values))

	// Not enough data for the first period-1 values
	for i := 0; i < period-1 && i < len(values); i++ {
		result[i] = 0
	}

	// Calculate SMA starting from period-1 index
	if len(values) >= period {
		// Calculate sum of first period values
		sum := 0.0
		for i := 0; i < period; i++ {
			sum += values[i]
		}
		result[period-1] = sum / float64(period)

		// Use sliding window for remaining values
		for i := period; i < len(values); i++ {
			sum = sum - values[i-period] + values[i]
			result[i] = sum / float64(period)
		}
	}

	return result
}

// EMA calculates Exponential Moving Average
func EMA(values []float64, period int) []float64 {
	if period <= 0 || len(values) == 0 {
		return []float64{}
	}

	result := make([]float64, len(values))

	// Multiplier for EMA
	multiplier := 2.0 / float64(period+1)

	// First EMA value is the first price
	if len(values) > 0 {
		result[0] = values[0]
	}

	// Calculate EMA for remaining values
	for i := 1; i < len(values); i++ {
		result[i] = (values[i]-result[i-1])*multiplier + result[i-1]
	}

	return result
}

// RSI calculates Relative Strength Index
func RSI(values []float64, period int) []float64 {
	if period <= 0 || len(values) < 2 {
		return []float64{}
	}

	result := make([]float64, len(values))

	// Not enough data for RSI calculation
	if len(values) <= period {
		return result
	}

	// Calculate price changes
	changes := make([]float64, len(values)-1)
	for i := 1; i < len(values); i++ {
		changes[i-1] = values[i] - values[i-1]
	}

	// Separate gains and losses
	gains := make([]float64, len(changes))
	losses := make([]float64, len(changes))
	for i := 0; i < len(changes); i++ {
		if changes[i] > 0 {
			gains[i] = changes[i]
		} else {
			losses[i] = -changes[i]
		}
	}

	// Calculate initial average gain and loss
	avgGain := 0.0
	avgLoss := 0.0
	for i := 0; i < period && i < len(gains); i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// First RSI value (at index period)
	if avgLoss == 0 {
		result[period] = 100
	} else {
		rs := avgGain / avgLoss
		result[period] = 100 - (100 / (1 + rs))
	}

	// Calculate RSI for remaining values using smoothed averages
	for i := period + 1; i < len(values); i++ {
		idx := i - 1 // index in gains/losses arrays
		if idx < len(gains) {
			avgGain = (avgGain*float64(period-1) + gains[idx]) / float64(period)
			avgLoss = (avgLoss*float64(period-1) + losses[idx]) / float64(period)
		}

		if avgLoss == 0 {
			result[i] = 100
		} else {
			rs := avgGain / avgLoss
			result[i] = 100 - (100 / (1 + rs))
		}
	}

	return result
}

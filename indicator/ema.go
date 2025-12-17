package indicator

type EMA struct {
	Period int
	Value  float64
	k      float64
	initialized bool
}

func NewEMA(period int) *EMA {
	return &EMA{
		Period: period,
		k:      2.0 / float64(period+1),
		initialized: false,
	}
}

func (e *EMA) Update(price float64) float64 {
	if !e.initialized {
		e.Value = price
		e.initialized = true
		return e.Value
	}
	// EMA = Price * k + PrevEMA * (1-k)
	return (price * e.k) + (e.Value * (1 - e.k))
}

// UpdateAndCommit updates the EMA and stores the new value (for closed candles)
func (e *EMA) UpdateAndCommit(price float64) float64 {
	newValue := e.Update(price)
	e.Value = newValue
	return newValue
}

// Calculate returns the EMA value for a given price without updating the state
func (e *EMA) Calculate(price float64) float64 {
	if !e.initialized {
		return price
	}
	return (price * e.k) + (e.Value * (1 - e.k))
}

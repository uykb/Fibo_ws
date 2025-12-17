package indicator

type CrossType int

const (
	None CrossType = iota
	GoldenCross
	DeathCross
)

// CheckCrossover determines if a crossover happened given previous and current EMA values.
// This is a stateless check. The caller maintains state.
func CheckCrossover(prevShort, prevLong, currShort, currLong, price float64) CrossType {
	// Golden Cross: Short goes from below Long to above Long
	if prevShort < prevLong && currShort > currLong {
		// Filter: Price must be above Long EMA
		if price > currLong {
			return GoldenCross
		}
	}

	// Death Cross: Short goes from above Long to below Long
	if prevShort > prevLong && currShort < currLong {
		// Filter: Price must be below Long EMA
		if price < currLong {
			return DeathCross
		}
	}

	return None
}

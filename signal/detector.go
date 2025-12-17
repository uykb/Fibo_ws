package signal

import (
	"fmt"
	"sync"
	"time"

	"fibo-monitor/data/kline"
	"fibo-monitor/indicator"

	"go.uber.org/zap"
)

type Signal struct {
	Type       indicator.CrossType
	Symbol     string
	Interval   string
	Price      float64
	ShortEMA   float64
	LongEMA    float64
	Timestamp  time.Time
}

type Detector struct {
	shortPeriod int
	longPeriod  int
	// state: symbol -> interval -> *state
	state  map[string]map[string]*pairState
	mu     sync.Mutex
	logger *zap.Logger
}

type pairState struct {
	ShortEMA *indicator.EMA
	LongEMA  *indicator.EMA
}

func NewDetector(shortPeriod, longPeriod int, logger *zap.Logger) *Detector {
	return &Detector{
		shortPeriod: shortPeriod,
		longPeriod:  longPeriod,
		state:       make(map[string]map[string]*pairState),
		logger:      logger,
	}
}

func (d *Detector) Detect(inChan <-chan kline.KlineEvent) <-chan Signal {
	outChan := make(chan Signal, 100)

	go func() {
		defer close(outChan)
		for event := range inChan {
			d.mu.Lock()
			// Initialize map for symbol if not exists
			if _, ok := d.state[event.Symbol]; !ok {
				d.state[event.Symbol] = make(map[string]*pairState)
			}
			
			// Initialize state for interval if not exists
			if _, ok := d.state[event.Symbol][event.Kline.Interval]; !ok {
				d.state[event.Symbol][event.Kline.Interval] = &pairState{
					ShortEMA: indicator.NewEMA(d.shortPeriod),
					LongEMA:  indicator.NewEMA(d.longPeriod),
				}
			}

			state := d.state[event.Symbol][event.Kline.Interval]
			d.mu.Unlock()

			price, err := event.Kline.GetClosePrice()
			if err != nil {
				d.logger.Error("Invalid price", zap.Error(err))
				continue
			}

			// Calculate current EMAs (temporary for this tick)
			// The stored EMA values are from the *previous closed* candle (or initial).
			// So 'prev' corresponds to the state at the beginning of this candle.
			// 'curr' corresponds to the state right now.
			
			prevShort := state.ShortEMA.Value
			prevLong := state.LongEMA.Value
			
			currShort := state.ShortEMA.Calculate(price)
			currLong := state.LongEMA.Calculate(price)

			// Check crossover
			crossType := indicator.CheckCrossover(prevShort, prevLong, currShort, currLong, price)

			if crossType != indicator.None {
				outChan <- Signal{
					Type:      crossType,
					Symbol:    event.Symbol,
					Interval:  event.Kline.Interval,
					Price:     price,
					ShortEMA:  currShort,
					LongEMA:   currLong,
					Timestamp: time.Now(),
				}
			}

			// If candle is closed, update the settled EMA state
			if event.Kline.IsClosed {
				state.ShortEMA.UpdateAndCommit(price)
				state.LongEMA.UpdateAndCommit(price)
			}
		}
	}()

	return outChan
}

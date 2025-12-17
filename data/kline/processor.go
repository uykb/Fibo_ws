package kline

import (
	"encoding/json"

	"go.uber.org/zap"
)

type Processor struct {
	logger *zap.Logger
}

func NewProcessor(logger *zap.Logger) *Processor {
	return &Processor{
		logger: logger,
	}
}

func (p *Processor) Process(msgChan <-chan []byte) <-chan KlineEvent {
	outChan := make(chan KlineEvent, 100)

	go func() {
		defer close(outChan)
		for msg := range msgChan {
			var event struct {
				Stream string          `json:"stream"`
				Data   json.RawMessage `json:"data"`
			}
			// Handling combined stream format: {"stream":"<streamName>","data":<payload>}
			if err := json.Unmarshal(msg, &event); err != nil {
				// Fallback to direct payload if not combined stream (though we use combined)
				var klineEvent KlineEvent
				if err2 := json.Unmarshal(msg, &klineEvent); err2 == nil {
					outChan <- klineEvent
				} else {
					p.logger.Error("Failed to unmarshal message", zap.Error(err), zap.String("msg", string(msg)))
				}
				continue
			}
			
			var klineEvent KlineEvent
			if err := json.Unmarshal(event.Data, &klineEvent); err != nil {
				p.logger.Error("Failed to unmarshal kline event", zap.Error(err))
				continue
			}
			outChan <- klineEvent
		}
	}()

	return outChan
}

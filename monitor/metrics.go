package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	WebsocketConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "fibo_websocket_connections",
		Help: "Current number of WebSocket connections",
	})

	KlineReceivedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "fibo_kline_received_total",
		Help: "Total number of kline events received",
	})

	SignalsDetectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fibo_signals_detected_total",
		Help: "Total number of signals detected",
	}, []string{"symbol", "interval", "type"})

	WebhookSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fibo_webhook_sent_total",
		Help: "Total number of webhooks sent",
	}, []string{"status"})
)

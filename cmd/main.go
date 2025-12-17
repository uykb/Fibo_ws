package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"fibo-monitor/config"
	"fibo-monitor/data/kline"
	"fibo-monitor/data/websocket"
	"fibo-monitor/monitor"
	"fibo-monitor/notification"
	pkgSignal "fibo-monitor/signal"

	"go.uber.org/zap"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 2. Init Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting Fibo Monitor...")

	// 3. Init Monitor
	monServer := monitor.NewServer(cfg.Monitoring, logger)
	monServer.Start()

	// 4. Init Components
	// Webhook
	webhookSender := notification.NewWebhookSender(cfg.Webhook, cfg.MessageCard, logger)

	// Filter
	sigFilter := pkgSignal.NewFilter(cfg.Signal.DeduplicationWindow, logger)

	// Detector
	detector := pkgSignal.NewDetector(
		cfg.Indicators.EmaShortPeriod,
		cfg.Indicators.EmaLongPeriod,
		logger,
	)

	// Processor
	processor := kline.NewProcessor(logger)

	// WebSocket Client
	wsClient := websocket.NewClient(
		cfg.Binance.WebsocketURL,
		cfg.Binance.ReconnectInterval,
		cfg.Binance.PingInterval,
		logger,
	)

	// 5. Connect Streams
	// Prepare stream names: <symbol>@kline_<interval>
	var streams []string
	for _, s := range cfg.Symbols {
		for _, i := range cfg.Intervals {
			streams = append(streams, fmt.Sprintf("%s@kline_%s", s, i))
		}
	}

	if err := wsClient.Connect(streams); err != nil {
		logger.Fatal("Failed to connect to WebSocket", zap.Error(err))
	}

	// 6. Data Pipeline
	// WS -> MsgChan
	msgChan := wsClient.Messages()

	// MsgChan -> Processor -> KlineChan
	klineChan := processor.Process(msgChan)

	// KlineChan -> Detector -> RawSignalChan
	// We need to tap into klineChan to update metrics? 
	// Or Detector updates metrics? 
	// Ideally Detector updates metrics or we wrap the channel.
	// For simplicity, let's update metrics in the loop or in components.
	// Let's assume components handle their logic.
	
	rawSignalChan := detector.Detect(klineChan)

	// RawSignalChan -> Filter -> FilteredSignalChan
	filteredSignalChan := sigFilter.Run(rawSignalChan)

	// FilteredSignalChan -> Webhook
	go func() {
		for sig := range filteredSignalChan {
			monitor.SignalsDetectedTotal.WithLabelValues(sig.Symbol, sig.Interval, sig.String()).Inc()
			logger.Info("Signal Detected", 
				zap.String("symbol", sig.Symbol),
				zap.String("interval", sig.Interval),
				zap.String("type", sig.String()),
				zap.Float64("price", sig.Price),
			)
			webhookSender.Send(sig)
		}
	}()

	// 7. Wait for shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down...")
	wsClient.Close()
}

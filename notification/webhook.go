package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fibo-monitor/config"
	"fibo-monitor/signal"

	"go.uber.org/zap"
)

type WebhookSender struct {
	config      config.WebhookConfig
	cardBuilder *MessageCard
	client      *http.Client
	logger      *zap.Logger
}

func NewWebhookSender(cfg config.WebhookConfig, cardCfg config.MessageCardConfig, logger *zap.Logger) *WebhookSender {
	return &WebhookSender{
		config:      cfg,
		cardBuilder: NewMessageCard(cardCfg),
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (w *WebhookSender) Send(sig signal.Signal) {
	// 1. Generic Webhook
	if w.config.Enabled {
		// Generic payload? Or card?
		// For now, let's send JSON of signal
		go w.sendGeneric(sig)
	}

	// 2. Lark Webhook
	if w.config.Lark.Enabled {
		go w.sendLark(sig)
	}
}

func (w *WebhookSender) sendGeneric(sig signal.Signal) {
	payload, err := json.Marshal(sig)
	if err != nil {
		w.logger.Error("Failed to marshal signal", zap.Error(err))
		return
	}

	w.performRequest(w.config.URL, payload)
}

func (w *WebhookSender) sendLark(sig signal.Signal) {
	msg := w.cardBuilder.BuildLarkMessage(sig)
	payload, err := json.Marshal(msg)
	if err != nil {
		w.logger.Error("Failed to marshal lark message", zap.Error(err))
		return
	}

	w.performRequest(w.config.Lark.WebhookURL, payload)
}

func (w *WebhookSender) performRequest(url string, payload []byte) {
	for i := 0; i <= w.config.RetryCount; i++ {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
		if err != nil {
			w.logger.Error("Failed to create request", zap.Error(err))
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := w.client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				w.logger.Info("Webhook sent successfully", zap.String("url", url))
				return
			}
			err = fmt.Errorf("status code: %d", resp.StatusCode)
		}

		w.logger.Warn("Webhook failed", zap.String("url", url), zap.Error(err), zap.Int("attempt", i+1))
		
		if i < w.config.RetryCount {
			time.Sleep(w.config.RetryBackoff)
		}
	}
	w.logger.Error("Webhook failed after retries", zap.String("url", url))
}

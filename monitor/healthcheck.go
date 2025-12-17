package monitor

import (
	"fmt"
	"net/http"
	"time"

	"fibo-monitor/config"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Server struct {
	config config.MonitoringConfig
	logger *zap.Logger
}

func NewServer(cfg config.MonitoringConfig, logger *zap.Logger) *Server {
	return &Server{
		config: cfg,
		logger: logger,
	}
}

func (s *Server) Start() {
	if s.config.PrometheusEnabled {
		go s.startPrometheus()
	}
	go s.startHealthCheck()
}

func (s *Server) startPrometheus() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	
	addr := fmt.Sprintf(":%d", s.config.PrometheusPort)
	s.logger.Info("Starting Prometheus server", zap.String("addr", addr))
	
	if err := http.ListenAndServe(addr, mux); err != nil {
		s.logger.Error("Prometheus server failed", zap.Error(err))
	}
}

func (s *Server) startHealthCheck() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := fmt.Sprintf(":%d", s.config.HealthcheckPort)
	s.logger.Info("Starting Health check server", zap.String("addr", addr))

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		s.logger.Error("Health check server failed", zap.Error(err))
	}
}

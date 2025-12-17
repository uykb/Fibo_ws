package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Binance     BinanceConfig     `mapstructure:"binance"`
	Symbols     []string          `mapstructure:"symbols"`
	Intervals   []string          `mapstructure:"intervals"`
	Indicators  IndicatorsConfig  `mapstructure:"indicators"`
	Signal      SignalConfig      `mapstructure:"signal"`
	Webhook     WebhookConfig     `mapstructure:"webhook"`
	MessageCard MessageCardConfig `mapstructure:"message_card"`
	Monitoring  MonitoringConfig  `mapstructure:"monitoring"`
}

type BinanceConfig struct {
	WebsocketURL      string        `mapstructure:"websocket_url"`
	ReconnectInterval time.Duration `mapstructure:"reconnect_interval"`
	PingInterval      time.Duration `mapstructure:"ping_interval"`
}

type IndicatorsConfig struct {
	EmaShortPeriod int `mapstructure:"ema_short_period"`
	EmaLongPeriod  int `mapstructure:"ema_long_period"`
}

type SignalConfig struct {
	DeduplicationWindow time.Duration `mapstructure:"deduplication_window"`
	MinVolume           float64       `mapstructure:"min_volume"`
}

type WebhookConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	URL          string        `mapstructure:"url"`
	Secret       string        `mapstructure:"secret"`
	Timeout      time.Duration `mapstructure:"timeout"`
	RetryCount   int           `mapstructure:"retry_count"`
	RetryBackoff time.Duration `mapstructure:"retry_backoff"`
}

type MessageCardConfig struct {
	Title            string             `mapstructure:"title"`
	ThemeColor       string             `mapstructure:"theme_color"`
	IncludePrice     bool               `mapstructure:"include_price"`
	IncludeEmaValues bool               `mapstructure:"include_ema_values"`
	IncludeTimestamp bool               `mapstructure:"include_timestamp"`
	LarkSpecific     LarkSpecificConfig `mapstructure:"lark_specific"`
}

type LarkSpecificConfig struct {
	AtAll   bool           `mapstructure:"at_all"`
	AtUsers []string       `mapstructure:"at_users"`
	Buttons []ButtonConfig `mapstructure:"buttons"`
}

type ButtonConfig struct {
	Text   string `mapstructure:"text"`
	URL    string `mapstructure:"url"`
	Action string `mapstructure:"action"`
}

type MonitoringConfig struct {
	HealthcheckPort   int    `mapstructure:"healthcheck_port"`
	LogLevel          string `mapstructure:"log_level"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("FIBO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Set default values if not present
	if config.Webhook.Timeout == 0 {
		config.Webhook.Timeout = 10 * time.Second
	}

	return &config, nil
}
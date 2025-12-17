package websocket

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	url               string
	reconnectInterval time.Duration
	pingInterval      time.Duration
	conn              *websocket.Conn
	stopChan          chan struct{}
	msgChan           chan []byte
	logger            *zap.Logger
	mu                sync.Mutex
	isConnected       bool
	streams           []string
}

func NewClient(url string, reconnectInterval, pingInterval time.Duration, logger *zap.Logger) *Client {
	return &Client{
		url:               url,
		reconnectInterval: reconnectInterval,
		pingInterval:      pingInterval,
		stopChan:          make(chan struct{}),
		msgChan:           make(chan []byte, 100),
		logger:            logger,
	}
}

func (c *Client) Connect(streams []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.streams = streams
	return c.connectInternal()
}

func (c *Client) connectInternal() error {
	if c.isConnected {
		return nil
	}

	baseURL := strings.TrimSuffix(c.url, "/")
	if strings.HasSuffix(baseURL, "/ws") {
		baseURL = strings.TrimSuffix(baseURL, "/ws") + "/stream"
	}
	
	streamParams := strings.Join(c.streams, "/")
	fullURL := fmt.Sprintf("%s?streams=%s", baseURL, streamParams)

	c.logger.Info("Connecting to WebSocket", zap.String("url", fullURL))

	conn, _, err := websocket.DefaultDialer.Dial(fullURL, nil)
	if err != nil {
		return err
	}

	c.conn = conn
	c.isConnected = true

	go c.readLoop()
	
	return nil
}

func (c *Client) readLoop() {
	defer func() {
		c.mu.Lock()
		if c.conn != nil {
			c.conn.Close()
		}
		c.isConnected = false
		c.mu.Unlock()
		
		// Don't reconnect if stopped
		select {
		case <-c.stopChan:
			return
		default:
			c.reconnect()
		}
	}()

	for {
		select {
		case <-c.stopChan:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				c.logger.Error("Read error", zap.Error(err))
				return
			}
			c.msgChan <- message
		}
	}
}

func (c *Client) reconnect() {
	c.logger.Info("Attempting to reconnect...")
	for {
		select {
		case <-c.stopChan:
			return
		default:
			time.Sleep(c.reconnectInterval)
			c.mu.Lock()
			err := c.connectInternal()
			c.mu.Unlock()
			if err == nil {
				c.logger.Info("Reconnected successfully")
				return
			}
			c.logger.Error("Reconnection failed", zap.Error(err))
		}
	}
}

func (c *Client) Messages() <-chan []byte {
	return c.msgChan
}

func (c *Client) Close() {
	close(c.stopChan)
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
	}
	c.mu.Unlock()
}

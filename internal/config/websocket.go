// internal/config/websocket.go
package config

import "time"

// WebSocketConfig holds WebSocket-specific configuration
type WebSocketConfig struct {
	Enabled            bool          `mapstructure:"WEBSOCKET_ENABLED"`
	ReadBufferSize     int           `mapstructure:"WEBSOCKET_READ_BUFFER_SIZE"`
	WriteBufferSize    int           `mapstructure:"WEBSOCKET_WRITE_BUFFER_SIZE"`
	MaxMessageSize     int64         `mapstructure:"WEBSOCKET_MAX_MESSAGE_SIZE"`
	HandshakeTimeout   time.Duration `mapstructure:"WEBSOCKET_HANDSHAKE_TIMEOUT"`
	WriteWait          time.Duration `mapstructure:"WEBSOCKET_WRITE_WAIT"`
	PongWait           time.Duration `mapstructure:"WEBSOCKET_PONG_WAIT"`
	PingPeriod         time.Duration `mapstructure:"WEBSOCKET_PING_PERIOD"`
	MaxConnections     int           `mapstructure:"WEBSOCKET_MAX_CONNECTIONS"`
	MessageBufferSize  int           `mapstructure:"WEBSOCKET_MESSAGE_BUFFER_SIZE"`
	EnablePresence     bool          `mapstructure:"WEBSOCKET_ENABLE_PRESENCE"`
	EnableMessageStore bool          `mapstructure:"WEBSOCKET_ENABLE_MESSAGE_STORE"`
	PersistenceEnabled bool          `mapstructure:"WEBSOCKET_PERSISTENCE_ENABLED"`
}

// DefaultWebSocketConfig returns default WebSocket configuration
func DefaultWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		Enabled:            true,
		ReadBufferSize:     1024,
		WriteBufferSize:    1024,
		MaxMessageSize:     512 * 1024, // 512KB
		HandshakeTimeout:   10 * time.Second,
		WriteWait:          10 * time.Second,
		PongWait:           60 * time.Second,
		PingPeriod:         (60 * time.Second * 9) / 10,
		MaxConnections:     10000,
		MessageBufferSize:  256,
		EnablePresence:     true,
		EnableMessageStore: true,
		PersistenceEnabled: true,
	}
}

package handler

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/synycboom/algorand-notification/client"
	"github.com/synycboom/algorand-notification/hub"
)

// Upgrader represents a contract for http upgrade
type Upgrader interface {
	// Upgrade upgrades http connection to websocket connection
	Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error)
}

// Hub represents a contract for websocket hub
type Hub interface {
	// Register registers a client to the hub
	Register(client hub.Client)
}

// ClientFactory represents a contract for client factory
type ClientFactory interface {
	// New creates a new websocket client
	New(conn client.GorillaConnection) *client.Client
}

// Config is a configuration
type Config struct {
	Hub           Hub
	Upgrader      Upgrader
	ClientFactory ClientFactory
}

// Handler is a http handler
type Handler struct {
	conf Config
}

// New creates a new handler
func New(c Config) *Handler {
	return &Handler{
		conf: c,
	}
}


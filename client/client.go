package client

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/atomic"

	"github.com/synycboom/algorand-notification/event"
)

const (
	invalidFormat     = 400
	methodSubscribe   = "SUBSCRIBE"
	methodUnsubscribe = "UNSUBSCRIBE"
)

var (
	validSubscriptionEvents = map[string]struct{}{
		event.NewBlock: {},
	}
)

// Request represents a websocket request payload
type Request struct {
	ID     int      `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

// Response represents a websocket response payload
type Response struct {
	ID     int            `json:"id,omitempty"`
	Error  *ResponseError `json:"error,omitempty"`
	Result interface{}    `json:"result,omitempty"`
}

// ResponseError is a response error
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GorillaConnection represents a contract for gorilla websocket connection
type GorillaConnection interface {
	SetReadDeadline(t time.Time) error
	SetReadLimit(limit int64)
	SetPongHandler(h func(appData string) error)
	ReadMessage() (messageType int, p []byte, err error)
	SetWriteDeadline(t time.Time) error
	WriteMessage(messageType int, data []byte) error
	Close() error
}

// Config represents a factory configuration
type Config struct {
	WriteWaitTimeout   time.Duration
	PongWaitTimeout    time.Duration
	PingInterval       time.Duration
	MaxReadMessageSize int64
	SendBufferSize     int
}

// Factory is a factory for creating websocket clients
type Factory struct {
	conf  Config
	total *atomic.Uint64
}

// NewFactory creates a client factory
func NewFactory(c Config) (*Factory, error) {
	if c.PongWaitTimeout < c.PingInterval {
		return nil, fmt.Errorf("factory: PongWaitTimeout must be greater than PingInterval")
	}

	return &Factory{
		conf:  c,
		total: atomic.NewUint64(0),
	}, nil
}

func (cf *Factory) New(conn GorillaConnection) *Client {
	id := cf.total.Add(1)
	c := &Client{
		conf:           cf.conf,
		closeChan:      make(chan struct{}),
		conn:           conn,
		id:             id,
		isUnregistered: false,
		mu:             sync.Mutex{},
		sendChan:       make(chan []byte, cf.conf.SendBufferSize),
	}

	go c.write()
	go c.read()

	return c
}

// Client represents websocket client
type Client struct {
	closeChan          chan struct{}
	conn               GorillaConnection
	conf               Config
	id                 uint64
	isUnregistered     bool
	mu                 sync.Mutex
	sendChan           chan []byte
	closeHandler       func()
	subscribeHandler   func(params []string)
	unsubscribeHandler func(params []string)
}

// ID returns a client id
func (c *Client) ID() uint64 {
	return c.id
}

// OnClose sets a close handler
func (c *Client) OnClose(h func()) {
	c.closeHandler = h
}

// OnSubscribe sets a subscribing handler
func (c *Client) OnSubscribe(h func(params []string)) {
	c.subscribeHandler = h
}

// OnSubscribe sets a unsubscribing handler
func (c *Client) OnUnsubscribe(h func(params []string)) {
	c.unsubscribeHandler = h
}

// IsClosed returns true if the client was closed
func (c *Client) IsClosed() bool {
	select {
	case <-c.closeChan:
		return true
	default:
	}

	return false
}

// Send sends a message to the peer
func (c *Client) Send(msg []byte) {
	if c.IsClosed() {
		return
	}

	select {
	case <-c.closeChan:
	case c.sendChan <- msg:
	}
}

func (c *Client) Close(code int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsClosed() {
		return
	}

	close(c.closeChan)
	time.Sleep(time.Duration(20) * time.Millisecond)
	close(c.sendChan)

	_ = c.conn.SetWriteDeadline(time.Now().Add(c.conf.WriteWaitTimeout))
	_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, ""))
	_ = c.conn.Close()

	if c.closeHandler != nil {
		c.closeHandler()
	}
}

func (c *Client) read() {
	defer c.Close(websocket.CloseGoingAway)

	logger := c.logger()
	if err := c.conn.SetReadDeadline(time.Now().Add(c.conf.PongWaitTimeout)); err != nil {
		logger.Warn().Err(err).Msg("client: failed to set read deadline")

		return
	}

	c.conn.SetReadLimit(c.conf.MaxReadMessageSize)
	c.conn.SetPongHandler(func(appData string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.conf.PongWaitTimeout)); err != nil {
			logger.Warn().Err(err).Msg("client: failed to set read deadline")

			return err
		}

		return nil
	})

	for {
		messageType, bb, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Warn().Err(err).Msg("client: connection was disconnected")
			}

			return
		}

		if c.IsClosed() {
			return
		}

		if messageType != websocket.TextMessage {
			continue
		}

		var payload Request
		if err := json.Unmarshal(bb, &payload); err != nil {
			response, err := newErrorResponse(0, invalidFormat, "payload is invalid")
			if err != nil {
				logger.Error().Err(err).Msg("client: failed to create an error response")

				return
			}

			c.Send(response)

			continue
		}

		switch payload.Method {
		case methodSubscribe:
			if err := c.validateSubscribing(payload); err != nil {
				res, err := newErrorResponse(payload.ID, invalidFormat, err.Error())
				if err != nil {
					logger.Error().Err(err).Msg("client: failed to create an error response")

					return
				}

				c.Send(res)

				continue
			}

			c.subscribeHandler(payload.Params)
			res, err := newSubscribingResponse(payload.ID)
			if err != nil {
				logger.Error().Err(err).Msg("client: failed to create a subscribing response")

				return
			}

			c.Send(res)
		case methodUnsubscribe:
			if err := c.validateUnSubscribing(payload); err != nil {
				res, err := newErrorResponse(payload.ID, invalidFormat, err.Error())
				if err != nil {
					logger.Error().Err(err).Msg("client: failed to create an error response")

					return
				}

				c.Send(res)

				continue
			}

			c.unsubscribeHandler(payload.Params)
			res, err := newUnsubscribingResponse(payload.ID)
			if err != nil {
				logger.Error().Err(err).Msg("client: failed to create an unsubscribing response")

				return
			}

			c.Send(res)
		}
	}
}

func (c *Client) write() {
	defer c.Close(websocket.CloseGoingAway)

	ticker := time.NewTicker(c.conf.PingInterval)
	defer ticker.Stop()

	logger := c.logger()
	for {
		select {
		case <-ticker.C:
			logger.Debug().Msgf("client: send ping #%d", c.id)
			if err := c.conn.SetWriteDeadline(time.Now().Add(c.conf.WriteWaitTimeout)); err != nil {
				logger.Warn().Err(err).Msg("client: failed to set write deadline")

				return
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Warn().Err(err).Msg("client: failed to sent a ping message")

				return
			}
		case msg, open := <-c.sendChan:
			if !open {
				return
			}

			if err := c.conn.SetWriteDeadline(time.Now().Add(c.conf.WriteWaitTimeout)); err != nil {
				logger.Warn().Err(err).Msg("client: failed to set write deadline")

				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				logger.Warn().Err(err).Msg("client: failed to send a message")

				return
			}
		case <-c.closeChan:
			return
		}
	}
}

func (c *Client) logger() zerolog.Logger {
	return log.With().Fields(map[string]interface{}{
		"client_id": c.ID(),
	}).Logger()
}

func newErrorResponse(id, code int, message string) ([]byte, error) {
	bb, err := json.Marshal(Response{
		ID: id,
		Error: &ResponseError{
			Code:    code,
			Message: message,
		},
	})
	if err != nil {
		return nil, err
	}

	return bb, nil
}

func newUnsubscribeResponse(id int) ([]byte, error) {
	bb, err := json.Marshal(Response{
		ID:     id,
		Result: nil,
	})
	if err != nil {
		return nil, err
	}

	return bb, nil
}

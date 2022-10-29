package hub

import (
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
	"go.uber.org/atomic"

	"github.com/synycboom/algorand-notification/event"
)

// Client represents a contract for websocket client
type Client interface {
	// ID returns a client id
	ID() uint64

	// OnClose sets a close handler
	OnClose(h func())

	// OnSubscribe sets a subscribing handler
	OnSubscribe(h func(params []string))

	// OnSubscribe sets a unsubscribing handler
	OnUnsubscribe(h func(params []string))

	// Send sends a message to the client
	Send(msg []byte)
}

// SubscribeEvent is a subscription detail
type SubscribeEvent struct {
	ClientID uint64
	Types    []string
}

// UnsubscribeEvent is a unsubscription detail
type UnsubscribeEvent struct {
	ClientID uint64
	Types    []string
}

// Hub maintains a set of active clients
type Hub struct {
	closeChan       chan struct{}
	count           atomic.Uint64
	clients         map[uint64]Client
	pool            *ants.Pool
	subscriptions   map[string]map[uint64]struct{}
	subscribeChan   chan SubscribeEvent
	unsubscribeChan chan UnsubscribeEvent
	registerChan    chan Client
	unregisterChan  chan Client
	eventChan       chan *event.Event
}

// New creates a new hub
func New(workerPoolSize int) (*Hub, error) {
	pool, err := ants.NewPool(workerPoolSize)
	if err != nil {
		return nil, err
	}

	return &Hub{
		closeChan:       make(chan struct{}),
		clients:         make(map[uint64]Client),
		pool:            pool,
		subscriptions:   make(map[string]map[uint64]struct{}),
		eventChan:       make(chan *event.Event, 1024),
		subscribeChan:   make(chan SubscribeEvent),
		unsubscribeChan: make(chan UnsubscribeEvent),
		registerChan:    make(chan Client),
		unregisterChan:  make(chan Client),
	}, nil
}

// Close closes a hub (not thread safe)
func (h *Hub) Close() {
	close(h.closeChan)
}

// Register registers a client to the hub
func (h *Hub) Register(c Client) {
	c.OnClose(func() {
		h.Unsubscribe(UnsubscribeEvent{
			ClientID: c.ID(),
			Types:    event.AllEvents,
		})
		h.UnRegister(c)
	})
	c.OnSubscribe(func(params []string) {
		h.Subscribe(SubscribeEvent{
			ClientID: c.ID(),
			Types:    params,
		})
	})
	c.OnUnsubscribe(func(params []string) {
		h.Unsubscribe(UnsubscribeEvent{
			ClientID: c.ID(),
			Types:    params,
		})
	})

	h.registerChan <- c
}

// UnRegister unregisters a client from the hub
func (h *Hub) UnRegister(c Client) {
	h.unregisterChan <- c
}

// Subscribe handle subscribing
func (h *Hub) Subscribe(e SubscribeEvent) {
	h.subscribeChan <- e
}

// Unsubscribe handle unsubscribing
func (h *Hub) Unsubscribe(e UnsubscribeEvent) {
	h.unsubscribeChan <- e
}

// SendEvent sends an event to clients
func (h *Hub) SendEvent(e *event.Event) {
  h.eventChan <- e
}

// Run runs worker to manage clients
func (h *Hub) Run() {
	for {
		select {
		case <-h.closeChan:
			return
		case c := <-h.registerChan:
			h.clients[c.ID()] = c
			h.count.Inc()

			log.Info().Msgf("hub: %v active sessions", h.count.Load())
		case c := <-h.unregisterChan:
			delete(h.clients, c.ID())
			h.count.Dec()

			log.Info().Msgf("hub: %v active sessions", h.count.Load())
		case e := <-h.subscribeChan:
			for _, evtType := range e.Types {
				if _, exist := h.subscriptions[evtType]; !exist {
					h.subscriptions[evtType] = make(map[uint64]struct{})
				}

				h.subscriptions[evtType][e.ClientID] = struct{}{}
			}
		case e := <-h.unsubscribeChan:
			for _, evtType := range e.Types {
				delete(h.subscriptions[evtType], e.ClientID)
				if len(h.subscriptions[evtType]) == 0 {
					delete(h.subscriptions, evtType)
				}
			}
		case event := <-h.eventChan:
			var clients []Client
			var wg sync.WaitGroup
			for clientID := range h.subscriptions[event.Type] {
				clients = append(clients, h.clients[clientID])
			}

			logger := log.With().Fields(map[string]interface{}{
				"event_type":    event.Type,
				"event_payload": string(event.Payload),
			}).Logger()

			for _, client := range clients {
				client := client
				wg.Add(1)

				err := h.pool.Submit(func() {
					defer wg.Done()

					client.Send(event.Payload)
				})
				if err != nil {
					wg.Done()
					logger.Error().Msg("hub: cannot submit task to the worker pool")
				}
			}

			wg.Wait()
		}
	}
}

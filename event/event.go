package event

import (
	"encoding/json"

	"github.com/algorand/go-algorand-sdk/client/algod/models"
)

const (
	// NewBlock is a an event occuring when a new block is mined
	NewBlock = "NEW_BLOCK"
)

var (
	// AllEvents represents all events
	AllEvents = []string{NewBlock}
)

// Event represents an event
type Event struct {
	Type string
	Payload []byte
}

// Parse raw data to an event
func Parse(data []byte) ([]*Event, error) {
  var events []*Event
  var block models.Block
  if err := json.Unmarshal(data, &block); err != nil {
    return nil, err
  }

  events = append(events, &Event{
    Type: NewBlock,
    Payload: data,
  })

  return events, nil
}


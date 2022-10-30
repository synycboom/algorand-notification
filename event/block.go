package event

import (
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

// BlockEvent represents a new block event
type BlockEvent struct {
	EventType string       `json:"eventType"`
	Data      models.Block `json:"data"`
}

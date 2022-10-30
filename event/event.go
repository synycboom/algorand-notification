package event

import (
	"encoding/json"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/iancoleman/strcase"
)

const (
	// NewBlock is the event for a new block
	NewBlock = "NEW_BLOCK"

	// NewPaymentTx is the event for payment transactions
	NewPaymentTx = "NEW_PAYMENT_TX"

	// NewKeyRegistrationTx is the event for key registration transactions
	NewKeyRegistrationTx = "NEW_KEY_REGISTRATION_TX"

	// NewAssetConfigTx is the event for creating, re-configuring, or destroying an asset
	NewAssetConfigTx = "NEW_ASSET_CONFIG_TX"

	// NewAssetTransferTx is the event for transferring assets between accounts (optionally closing)
	NewAssetTransferTx = "NEW_ASSET_TRANSFER_TX"

	// NewAssetFreezeTx is the event for changing the freeze status of an asset
	NewAssetFreezeTx = "NEW_ASSET_FREEZE_TX"

	// NewApplicationCallTx is the event for creating, deleting, and interacting with an application
	NewApplicationCallTx = "NEW_APPLICATION_CALL_TX"

	// NewStateProofTx is the event for recording a state proof
	NewStateProofTx = "NEW_STATE_PROOF_TX"
)

var (
	// AllEvents represents all events
	AllEvents = []string{
		NewBlock,
		NewPaymentTx,
		NewKeyRegistrationTx,
		NewAssetConfigTx,
		NewAssetTransferTx,
		NewAssetFreezeTx,
		NewApplicationCallTx,
		NewStateProofTx,
	}
)

// Event represents an event
type Event struct {
	Type    string
	Payload []byte
}

// Parse raw data to an event
func Parse(data []byte) ([]*Event, error) {
	var events []*Event
	var block models.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, err
	}

	blockEvent := BlockEvent{
		EventType: NewBlock,
		Data:      block,
	}

	payload, err := json.Marshal(blockEvent)
	if err != nil {
		return nil, err
	}

	events = append(events, &Event{
		Type:    NewBlock,
		Payload: convertKeys(payload),
	})

	for _, tx := range block.Transactions {
		txEvent := NewTransactionEvent(tx)
		payload, err := json.Marshal(txEvent)
		if err != nil {
			return nil, err
		}

		events = append(events, &Event{
			Type:    txEvent.EventType,
			Payload: convertKeys(payload),
		})
	}

	return events, nil
}

// convertKeys converts keys to camel case
func convertKeys(data []byte) []byte {
	m := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &m); err != nil {
		return data
	}

	for k, v := range m {
		camelized := strcase.ToLowerCamel(k)
		delete(m, k)
		m[camelized] = convertKeys(v)
	}

	bb, err := json.Marshal(m)
	if err != nil {
		return data
	}

	return bb
}

package event

import (
	"encoding/json"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/types"
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

	events = append(events, &Event{
		Type:    NewBlock,
		Payload: data,
	})

	for _, tx := range block.Transactions {
		txData, err := json.Marshal(tx)
		if err != nil {
			return nil, err
		}

		// TODO: filter unrelated transaction object (it does not omit empty), and rename event type in payload
		switch tx.Type {
		case string(types.PaymentTx):
			events = append(events, &Event{
				Type:    NewPaymentTx,
				Payload: txData,
			})
		case string(types.KeyRegistrationTx):
			events = append(events, &Event{
				Type:    NewKeyRegistrationTx,
				Payload: txData,
			})
		case string(types.AssetConfigTx):
			events = append(events, &Event{
				Type:    NewAssetConfigTx,
				Payload: txData,
			})
		case string(types.AssetTransferTx):
			events = append(events, &Event{
				Type:    NewAssetTransferTx,
				Payload: txData,
			})
		case string(types.AssetFreezeTx):
			events = append(events, &Event{
				Type:    NewAssetFreezeTx,
				Payload: txData,
			})
		case string(types.ApplicationCallTx):
			events = append(events, &Event{
				Type:    NewApplicationCallTx,
				Payload: txData,
			})
		case string(types.StateProofTx):
			events = append(events, &Event{
				Type:    NewStateProofTx,
				Payload: txData,
			})
		}
	}

	return events, nil
}

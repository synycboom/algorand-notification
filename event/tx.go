package event

import (
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/types"
)

// TransactionEventData is an extended version of an original transaction to provide more cleaner version of fields
type TransactionEventData struct {
	models.Transaction

	// ApplicationTransaction fields for application transactions.
	ApplicationTransaction *models.TransactionApplication `json:"application-transaction,omitempty"`

	// AssetConfigTransaction fields for asset allocation, re-configuration, and
	// destruction.
	// A zero value for asset-id indicates asset creation.
	// A zero value for the params indicates asset destruction.
	AssetConfigTransaction *models.TransactionAssetConfig `json:"asset-config-transaction,omitempty"`

	// AssetFreezeTransaction fields for an asset freeze transaction.
	AssetFreezeTransaction *models.TransactionAssetFreeze `json:"asset-freeze-transaction,omitempty"`

	// AssetTransferTransaction fields for an asset transfer transaction.
	AssetTransferTransaction *models.TransactionAssetTransfer `json:"asset-transfer-transaction,omitempty"`

	// KeyregTransaction fields for a keyreg transaction.
	KeyregTransaction *models.TransactionKeyreg `json:"keyreg-transaction,omitempty"`

	// PaymentTransaction fields for a payment transaction.
	PaymentTransaction *models.TransactionPayment `json:"payment-transaction,omitempty"`

	// StateProofTransaction fields for a state proof transaction.
	StateProofTransaction *models.TransactionStateProof `json:"state-proof-transaction,omitempty"`
}

// TransactionEvent is a transaction event
type TransactionEvent struct {
	EventType string               `json:"eventType"`
	Data      TransactionEventData `json:"data"`
}

// NewTransactionEvent creats a tx event from an original tx
func NewTransactionEvent(tx models.Transaction) TransactionEvent {
	var data TransactionEventData
	var eventType string

	data.AuthAddr = tx.AuthAddr
	data.CloseRewards = tx.CloseRewards
	data.ClosingAmount = tx.ClosingAmount
	data.ConfirmedRound = tx.ConfirmedRound
	data.CreatedApplicationIndex = tx.CreatedApplicationIndex
	data.CreatedAssetIndex = tx.CreatedAssetIndex
	data.Fee = tx.Fee
	data.FirstValid = tx.FirstValid
	data.GenesisHash = tx.GenesisHash
	data.GenesisId = tx.GenesisId
	data.GlobalStateDelta = tx.GlobalStateDelta
	data.Group = tx.Group
	data.Id = tx.Id
	data.InnerTxns = tx.InnerTxns
	data.IntraRoundOffset = tx.IntraRoundOffset
	data.LastValid = tx.LastValid
	data.Lease = tx.Lease
	data.LocalStateDelta = tx.LocalStateDelta
	data.Logs = tx.Logs
	data.Note = tx.Note
	data.ReceiverRewards = tx.ReceiverRewards
	data.RekeyTo = tx.RekeyTo
	data.RoundTime = tx.RoundTime
	data.Sender = tx.Sender
	data.SenderRewards = tx.SenderRewards
	data.Signature = tx.Signature
	data.Type = tx.Type

	switch tx.Type {
	case string(types.PaymentTx):
		eventType = NewPaymentTx
		data.PaymentTransaction = &tx.PaymentTransaction
	case string(types.KeyRegistrationTx):
		eventType = NewKeyRegistrationTx
		data.KeyregTransaction = &tx.KeyregTransaction
	case string(types.AssetConfigTx):
		eventType = NewAssetConfigTx
		data.AssetConfigTransaction = &tx.AssetConfigTransaction
	case string(types.AssetTransferTx):
		eventType = NewAssetTransferTx
		data.AssetTransferTransaction = &tx.AssetTransferTransaction
	case string(types.AssetFreezeTx):
		eventType = NewAssetFreezeTx
		data.AssetFreezeTransaction = &tx.AssetFreezeTransaction
	case string(types.ApplicationCallTx):
		eventType = NewApplicationCallTx
		data.ApplicationTransaction = &tx.ApplicationTransaction
	case string(types.StateProofTx):
		eventType = NewStateProofTx
		data.StateProofTransaction = &tx.StateProofTransaction
	}

	return TransactionEvent{
		EventType: eventType,
		Data:      data,
	}
}

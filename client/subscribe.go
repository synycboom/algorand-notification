package client

import (
	"encoding/json"
	"fmt"
)

func (c *Client) validateSubscribing(req Request) error {
	if req.Method != methodSubscribe {
		return fmt.Errorf("invalid method")
	}

	if len(req.Params) == 0 {
		return fmt.Errorf("params are required")
	}

	for _, event := range req.Params {
		if _, ok := validSubscriptionEvents[event]; !ok {
			return fmt.Errorf("invalid params")
		}
	}

	return nil
}

func newSubscribingResponse(id uint64) ([]byte, error) {
	bb, err := json.Marshal(Response{
		ID:     id,
		Result: nil,
	})
	if err != nil {
		return nil, err
	}

	return bb, nil
}

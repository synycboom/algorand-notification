package client

import (
	"encoding/json"
	"fmt"
)

func (c *Client) validateUnSubscribing(req Request) error {
  if req.Method != methodUnsubscribe {
    return fmt.Errorf("invalid method")
  }

  for _, event := range req.Params {
    if _, ok := validSubscriptionEvents[event]; !ok {
      return fmt.Errorf("invalid params")
    }
  }

  return nil
}

func newUnsubscribingResponse(id int) ([]byte, error) {
	bb, err := json.Marshal(Response{
		ID:     id,
		Result: nil,
	})
	if err != nil {
		return nil, err
	}

	return bb, nil
}


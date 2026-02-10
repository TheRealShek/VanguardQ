package validators

import (
	"encoding/json"
	"errors"
)

const MaxPayloadSize = 256 * 1024 // 256KB

func IsValidPayload(p json.RawMessage) error {
	if len(p) == 0 {
		return errors.New("Payload cannot be empty")
	}
	if len(p) > MaxPayloadSize {
		return errors.New("Payload too large")
	}
	if !json.Valid(p) {
		return errors.New("Payload must be a Valid JSON")
	}
	return nil
}

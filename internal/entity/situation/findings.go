package csituation

import "encoding/json"

type Findings struct {
	Status []json.RawMessage `json:"status"`
}

package experiment

import "encoding/json"

type Configurable interface {
	Configure(cfg json.RawMessage) error
}

package experiments

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FieldSource is used to identify the source of the Litmus user data field
type FieldSource string

const (
	// PayloadFieldSource is used to represent the request payload
	PayloadFieldSource FieldSource = "payload"
	// HeaderFieldSource is used to represent the request header
	HeaderFieldSource FieldSource = "header"
)

// GetFieldSource converts the input string to a FieldSource
func GetFieldSource(srcString string) (FieldSource, error) {
	switch strings.ToLower(srcString) {
	case "header":
		return HeaderFieldSource, nil
	case "payload":
		return PayloadFieldSource, nil
	}
	return FieldSource(""), fmt.Errorf("Unknown field source %s", srcString)
}

// UnmarshalJSON unmarshals the data as a string and then creates the
// appropriate FieldSource
func (f *FieldSource) UnmarshalJSON(data []byte) error {
	var s string
	var err error
	if err = json.Unmarshal(data, &s); err != nil {
		return err
	}

	*f, err = GetFieldSource(s)
	return err
}

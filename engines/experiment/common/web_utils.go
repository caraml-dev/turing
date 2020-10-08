package common

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

// GetValueFromRequest parses the request header / payload to retrieve the value
// for the given field
//
// reqHeader - request header
// bodyBytes - request JSON payload
// fieldSrc - source of data, where the given key will be looked in,
//			  one of `PayloadFieldSource` | `HeaderFieldSource`
// field - if `fieldSrc` is `HeaderFieldSource` - name of request header
//		   if `fieldSrc` is `PayloadFieldSource` - json path to the value that should
//         be extracted from the request payload
func GetValueFromRequest(
	reqHeader http.Header,
	bodyBytes []byte,
	fieldSrc FieldSource,
	field string,
) (string, error) {
	var value string
	var err error

	switch fieldSrc {
	case PayloadFieldSource:
		value, err = getValueFromJSONPayload(bodyBytes, field)
	case HeaderFieldSource:
		value = reqHeader.Get(field)
		if value == "" {
			// key not found in header, set error
			err = fmt.Errorf("Field %s not found in the request header", field)
		}
	default:
		err = fmt.Errorf("Unrecognized field source %s", fieldSrc)
	}

	return value, err
}

func getValueFromJSONPayload(body []byte, key string) (string, error) {
	// Retrieve value using JSON path
	value, err := jsonparser.GetString(body, strings.Split(key, ".")...)
	if err != nil {
		return value, errors.Wrapf(err, "Field %s not found in the request payload", key)
	}
	return value, nil
}

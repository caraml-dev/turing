package experiments

import (
	"fmt"
	"net/http"
	"strings"
	"unsafe"

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
	switch fieldSrc {
	case PayloadFieldSource:
		return getValueFromJSONPayload(bodyBytes, field)
	case HeaderFieldSource:
		value := reqHeader.Get(field)
		if value == "" {
			// key not found in header
			return "", fmt.Errorf("Field %s not found in the request header", field)
		}
		return value, nil
	default:
		return "", fmt.Errorf("Unrecognized field source %s", fieldSrc)
	}
}

func getValueFromJSONPayload(body []byte, key string) (string, error) {
	// Retrieve value using JSON path
	value, typez, _, _ := jsonparser.Get(body, strings.Split(key, ".")...)

	switch typez {
	case jsonparser.String, jsonparser.Number, jsonparser.Boolean:
		// See: https://github.com/buger/jsonparser/blob/master/bytes_unsafe.go#L31
		return *(*string)(unsafe.Pointer(&value)), nil
	case jsonparser.Null:
		return "", nil
	case jsonparser.NotExist:
		return "", errors.Errorf("Field %s not found in the request payload: Key path not found", key)
	default:
		return "", errors.Errorf(
			"Field %s can not be parsed as string value, unsupported type: %s", key, typez.String())
	}
}

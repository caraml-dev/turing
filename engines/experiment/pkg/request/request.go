package request

import (
	"fmt"
	"net/http"
	"strings"
	"unsafe"

	"github.com/buger/jsonparser"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

// FieldSource is used to identify the source of the experiment-engine user data field
type FieldSource string

const (
	// PayloadFieldSource is used to represent the request payload
	PayloadFieldSource FieldSource = "payload"
	// HeaderFieldSource is used to represent the request header
	HeaderFieldSource FieldSource = "header"
	// PredictionContextSource is used to represent the prediction_context field in UPI request
	PredictionContextSource FieldSource = "prediction_context"
)

// GetFieldSource converts the input string to a FieldSource
func GetFieldSource(srcString string) (FieldSource, error) {
	switch strings.ToLower(srcString) {
	case "header":
		return HeaderFieldSource, nil
	case "payload":
		return PayloadFieldSource, nil
	case "prediction_context":
		return PredictionContextSource, nil
	}
	return "", fmt.Errorf("Unknown field source %s", srcString)
}

// GetValueFromHTTPRequest parses the request header / payload to retrieve the value
// for the given field
//
// reqHeader - request header
// bodyBytes - request JSON payload
// fieldSrc - source of data, where the given key will be looked in,
//
//	one of `PayloadFieldSource` | `HeaderFieldSource`
//
// field - if `fieldSrc` is `HeaderFieldSource` - name of request header
//
//			   if `fieldSrc` is `PayloadFieldSource` - json path to the value that should
//	        be extracted from the request payload
func GetValueFromHTTPRequest(
	reqHeader http.Header,
	bodyBytes []byte,
	fieldSrc FieldSource,
	field string,
) (string, error) {
	switch fieldSrc {
	case PayloadFieldSource:
		return getValueFromJSONPayload(bodyBytes, field)
	case HeaderFieldSource:
		value := strings.Join(reqHeader.Values(field), ",")
		if value == "" {
			// key not found in header
			return "", fmt.Errorf("Field %s not found in the request header", field)
		}
		return value, nil
	default:
		return "", fmt.Errorf("Unrecognized field source %s", fieldSrc)
	}
}

// GetValueFromUPIRequest retrieve the value from upi request or header depending on the value of `fieldSrc`.
// Valid value of `fieldSrc` are `HeaderFieldSource` and `PredictionContextSource`.
// If `fieldSrc` is `HeaderFieldSource`, then the value will be retrieved from `reqHeader`.
// If `fieldSrc` is `PredictionContextSource`, then the value will be retrieved from
// `prediction_context` field of the upi request `req`.
// Other `fieldSrc` value will produce error.
func GetValueFromUPIRequest(
	reqHeader metadata.MD,
	req *upiv1.PredictValuesRequest,
	fieldSrc FieldSource,
	field string,
) (string, error) {
	switch fieldSrc {
	case HeaderFieldSource:
		values := reqHeader.Get(field)
		if len(values) == 0 {
			// key not found in header
			return "", fmt.Errorf("Field %s not found in the request header", field)
		}

		// return first value to be consistent with `GetValueFromHTTPRequest`
		return values[0], nil
	case PredictionContextSource:
		predContext, err := UPIVariablesToStringMap(req.PredictionContext)
		if err != nil {
			return "", err
		}

		value, exists := predContext[field]
		if !exists {
			return "", fmt.Errorf("Variable %s not found in the prediction context", field)
		}

		return value, nil
	default:
		return "", fmt.Errorf("Unrecognized field source %s", fieldSrc)
	}
}

func getValueFromJSONPayload(body []byte, key string) (string, error) {
	// Retrieve value using JSON path
	value, dataType, _, _ := jsonparser.Get(body, strings.Split(key, ".")...)

	switch dataType {
	case jsonparser.String, jsonparser.Number, jsonparser.Boolean, jsonparser.Array, jsonparser.Object:
		// See: https://github.com/buger/jsonparser/blob/master/bytes_unsafe.go#L31
		return *(*string)(unsafe.Pointer(&value)), nil
	case jsonparser.Null:
		return "", nil
	// Default non exist
	default:
		return "", errors.Errorf("Field %s not found in the request payload: Key path not found", key)
	}
}

// UPIVariablesToStringMap convert slice of upi Variables into map of string
func UPIVariablesToStringMap(vars []*upiv1.Variable) (map[string]string, error) {
	strMap := map[string]string{}
	for _, v := range vars {
		vstr, err := getValueAsString(v)
		if err != nil {
			return nil, err
		}

		strMap[v.Name] = vstr
	}

	return strMap, nil
}

func getValueAsString(v *upiv1.Variable) (string, error) {
	switch v.Type {
	case upiv1.Type_TYPE_DOUBLE:
		return fmt.Sprintf("%f", v.DoubleValue), nil
	case upiv1.Type_TYPE_INTEGER:
		return fmt.Sprintf("%d", v.IntegerValue), nil
	case upiv1.Type_TYPE_STRING:
		return v.StringValue, nil
	default:
		return "", fmt.Errorf("Unknown value type %s", v.Type)
	}
}

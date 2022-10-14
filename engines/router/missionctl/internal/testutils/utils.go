package testutils

import (
	"io"
	"os"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// ReadFile reads a file and returns the byte contents
func ReadFile(filepath string) ([]byte, error) {
	// Open file
	fileObj, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer fileObj.Close()
	// Read contents
	byteValue, err := io.ReadAll(fileObj)
	if err != nil {
		return nil, err
	}
	return byteValue, nil
}

func CompareUpiResponse(x *upiv1.PredictValuesResponse, y *upiv1.PredictValuesResponse) bool {
	return cmp.Equal(x, y,
		cmpopts.IgnoreUnexported(
			upiv1.PredictValuesResponse{},
			upiv1.Table{},
			upiv1.Column{},
			upiv1.Row{},
			upiv1.Value{},
			upiv1.Variable{},
			upiv1.ResponseMetadata{},
		))
}

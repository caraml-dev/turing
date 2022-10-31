package testutils

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/api/turing/models"
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

// GetRouterVersion reads the given file, attempts to convert the contents into
// a router version model object and returns it. The method records a failure to
// the test s if an error is encountered.
func GetRouterVersion(t *testing.T, filePath string) *models.RouterVersion {
	fileBytes, err := ReadFile(filePath)
	require.NoError(t, err)

	// Convert to RouterVersion type
	var routerVersion models.RouterVersion
	err = json.Unmarshal(fileBytes, &routerVersion)
	require.NoError(t, err)

	return &routerVersion
}

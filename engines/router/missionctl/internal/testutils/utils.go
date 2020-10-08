package testutils

import (
	"io/ioutil"
	"os"
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
	byteValue, err := ioutil.ReadAll(fileObj)
	if err != nil {
		return nil, err
	}
	return byteValue, nil
}

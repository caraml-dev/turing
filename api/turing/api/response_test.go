package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	tu "github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestResponses(t *testing.T) {
	// Accepted
	assert.Equal(t, Response{
		code: 202,
		data: "test-data-accepted",
	}, *Accepted("test-data-accepted"))

	// OK
	assert.Equal(t, Response{
		code: 200,
		data: "test-data-ok",
	}, *Ok("test-data-ok"))

	// Created
	assert.Equal(t, Response{
		code: 201,
		data: "test-data-created",
	}, *Created("test-data-created"))

	// Error
	expectedResponse := &Response{
		code: 10,
		data: struct {
			Description string `json:"description"`
			Message     string `json:"error"`
		}{"desc", "msg"},
	}
	assert.Equal(t, expectedResponse, Error(10, "desc", "msg"))

	// Not Found
	expectedResponse = &Response{
		code: 404,
		data: struct {
			Description string `json:"description"`
			Message     string `json:"error"`
		}{"desc-404", "msg-404"},
	}
	assert.Equal(t, expectedResponse, Error(404, "desc-404", "msg-404"))

	// Bad Request
	expectedResponse = &Response{
		code: 400,
		data: struct {
			Description string `json:"description"`
			Message     string `json:"error"`
		}{"desc-400", "msg-400"},
	}
	assert.Equal(t, expectedResponse, Error(400, "desc-400", "msg-400"))

	// Internal Server Error
	expectedResponse = &Response{
		code: 500,
		data: struct {
			Description string `json:"description"`
			Message     string `json:"error"`
		}{"desc-500", "msg-500"},
	}
	assert.Equal(t, expectedResponse, Error(500, "desc-500", "msg-500"))
}

func TestMarshalResponse(t *testing.T) {
	resp := &Response{
		code: 10,
		data: "test-data",
	}

	// Marshal into JSON
	bytes, err := json.Marshal(resp)
	tu.FailOnError(t, err)
	assert.JSONEq(t, `
		{
			"code": 10,
			"data": "test-data"
		}
	`, string(bytes))
}

func TestWriteTo(t *testing.T) {
	// Create test response
	resp := &Response{
		code: 103,
		data: "test-data",
	}
	// Create test response writer
	rr := httptest.NewRecorder()
	// Expected header
	header := http.Header{
		"Content-Type": []string{"application/json; charset=UTF-8"},
	}
	// Validate
	resp.WriteTo(rr)
	resultResponse := rr.Result()
	tu.FailOnNil(t, rr.Body)
	defer resultResponse.Body.Close()
	assert.Equal(t, "\"test-data\"\n", rr.Body.String())
	assert.Equal(t, 103, rr.Code)
	assert.Equal(t, header, resultResponse.Header)
}

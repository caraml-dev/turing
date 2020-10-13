package api

import (
	"encoding/json"
	"net/http"
)

// Response contains the return code and data to return to the caller.
type Response struct {
	code int
	data interface{}
}

// MarshalJSON is a custom marshaler for the unexported fields
func (r *Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code int         `json:"code"`
		Data interface{} `json:"data"`
	}{
		Code: r.code,
		Data: r.data,
	})
}

// WriteTo writes an Response to the provided http.ResponseWriter.
func (r *Response) WriteTo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(r.code)

	if r.data != nil {
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(r.data)
	}
}

// Accepted returns an Response with status code 202.
func Accepted(data interface{}) *Response {
	return &Response{
		code: http.StatusAccepted,
		data: data,
	}
}

// Ok returns an Response with status code 200.
func Ok(data interface{}) *Response {
	return &Response{
		code: http.StatusOK,
		data: data,
	}
}

// Created returns an Response with status code 201.
func Created(data interface{}) *Response {
	return &Response{
		code: http.StatusCreated,
		data: data,
	}
}

// Error returns an Response with the provided status code and error message.
func Error(code int, description string, msg string) *Response {
	return &Response{
		code: code,
		data: struct {
			Description string `json:"description"`
			Message     string `json:"error"`
		}{description, msg},
	}
}

// NotFound returns an Response with the error code 404.
func NotFound(description, msg string) *Response {
	return Error(http.StatusNotFound, description, msg)
}

// BadRequest returns an Response with the error code 400.
func BadRequest(description, msg string) *Response {
	return Error(http.StatusBadRequest, description, msg)
}

// InternalServerError returns an Response with the error code 500.
func InternalServerError(description, msg string) *Response {
	return Error(http.StatusInternalServerError, description, msg)
}

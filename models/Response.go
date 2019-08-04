package models

import "encoding/json"

// RequestErrorCode - special type for error codes
type RequestErrorCode error

func NewRequestErrorCode(text string) RequestErrorCode {
	return &errorString{text}
}

// errorString - implementing error interface for RequestErrorCode
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

// it just returns string but not a JSON struct
func (e *errorString) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.s)
}

// Response - struct for sending payload from server and more info about occurred error
// It behaves like Either Monad: 'Error' field is set if error occurred, otherwise 'Body' contains payload
type Response struct {
	Error RequestErrorCode `json:"error"`
	Body  interface{}      `json:"body"`
}

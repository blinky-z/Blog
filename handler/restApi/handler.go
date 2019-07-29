package restApi

import (
	"encoding/json"
	"net/http"
)

// general error codes
const (
	// TechnicalError - internal server error
	TechnicalError RequestErrorCode = "TECHNICAL_ERROR"
	// BadRequestBody - invalid body
	BadRequestBody RequestErrorCode = "BAD_BODY"
	// NoPermissions - user doesn't permissions to create/update/delete resource
	NoPermissions RequestErrorCode = "NO_PERMISSIONS"
	// InvalidRequest - error code for other errors
	InvalidRequest RequestErrorCode = "INVALID_REQUEST"
	// NoError - no error occurred while handling request. Should not be exposed but only used internally
	NoError RequestErrorCode = ""
)

// RequestErrorCode - special type for error codes
type RequestErrorCode string

// Respond - helper function for responding with only status code
func Respond(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

// Response - struct for sending payload from server and more info about occurred error
// It behaves like Either Monad: 'Error' field is set if error occurred, otherwise 'Body' contains payload
type Response struct {
	Error RequestErrorCode `json:"error"`
	Body  interface{}      `json:"body"`
}

func respondWithJSON(w http.ResponseWriter, code int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(body)
}

// RespondWithError - helper function for responding with error in body
// This function uses special 'Response' struct. See above
func RespondWithError(w http.ResponseWriter, code int, errorCode RequestErrorCode) {
	response := &Response{
		Error: errorCode,
	}
	encodedResponse, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, code, encodedResponse)
}

// RespondWithBody - helper function for responding with payload in body
// This function uses special 'Response' struct. See above
func RespondWithBody(w http.ResponseWriter, code int, payload interface{}) {
	response := &Response{
		Body: payload,
	}
	encodedResponse, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, code, encodedResponse)
}

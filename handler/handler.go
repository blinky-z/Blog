package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	// TechnicalError - server error
	TechnicalError PostErrorCode = "TECHNICAL_ERROR"
	// NoError - no error occurred while handling request
	NoError PostErrorCode = ""
)

// Response - behaves like Either Monad
// 'Error' field is set while error occurred.
// Otherwise 'Body' field is used to return post from database
type Response struct {
	Error PostErrorCode `json:"error"`
	Body  interface{}   `json:"body"`
}

// PostErrorCode - represents error occurred while handling request
type PostErrorCode string

func respond(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

func respondWithJSON(w http.ResponseWriter, code int, body []byte, logError *log.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_, err := w.Write(body)
	if err != nil {
		logError.Print(err)
	}
}

//RespondWithError - write error in response body and respond
func RespondWithError(w http.ResponseWriter, code int, errorCode PostErrorCode, logError *log.Logger) {
	var response Response
	response.Error = errorCode
	encodedResponse, _ := json.Marshal(response)

	respondWithJSON(w, code, encodedResponse, logError)
}

//RespondWithBody - write body in response body and respond
func RespondWithBody(w http.ResponseWriter, code int, payload interface{}, logError *log.Logger) {
	var response Response
	response.Body = payload
	encodedResponse, _ := json.Marshal(response)

	respondWithJSON(w, code, encodedResponse, logError)
}

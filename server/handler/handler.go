package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

var (
	// LogInfo - log for writing+ messages
	LogInfo *log.Logger
	// LogError - log for writing server errors
	LogError *log.Logger

	// Db - database connection. This variable is set by main function
	Db *sql.DB
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

func respondWithJSON(w http.ResponseWriter, code int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_, err := w.Write(body)
	if err != nil {
		LogError.Print(err)
	}
}

func respondWithError(w http.ResponseWriter, code int, errorCode PostErrorCode) {
	var response Response
	response.Error = errorCode
	encodedResponse, _ := json.Marshal(response)

	respondWithJSON(w, code, encodedResponse)
}

func respondWithBody(w http.ResponseWriter, code int, payload interface{}) {
	var response Response
	response.Body = payload
	encodedResponse, _ := json.Marshal(response)

	respondWithJSON(w, code, encodedResponse)
}

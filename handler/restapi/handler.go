package restapi

import (
	"encoding/json"
	"fmt"
	"github.com/blinky-z/Blog/models"
	"net/http"
)

// general error codes
var (
	// TechnicalError - internal server error
	TechnicalError models.RequestErrorCode = models.NewRequestErrorCode("TECHNICAL_ERROR")
	// BadRequestBody - invalid body
	BadRequestBody models.RequestErrorCode = models.NewRequestErrorCode("BAD_BODY")
	// NoPermissions - user doesn't permissions to create/update/delete resource
	NoPermissions models.RequestErrorCode = models.NewRequestErrorCode("NO_PERMISSIONS")
	// InvalidRequest - error code for other errors
	InvalidRequest models.RequestErrorCode = models.NewRequestErrorCode("INVALID_REQUEST")
)

// Respond - helper function for responding with only status code
func Respond(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

func respondWithJSON(w http.ResponseWriter, code int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(body)
}

// RespondWithError - helper function for responding with error in body
// This function uses special 'Response' struct. See above
func RespondWithError(w http.ResponseWriter, code int, errorCode models.RequestErrorCode) {
	response := &models.Response{
		Error: errorCode,
		Body:  nil,
	}
	encodedResponse, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Printf("Built response: %v\n", string(encodedResponse))

	respondWithJSON(w, code, encodedResponse)
}

// RespondWithBody - helper function for responding with payload in body
// This function uses special 'Response' struct. See above
func RespondWithBody(w http.ResponseWriter, code int, payload interface{}) {
	response := &models.Response{
		Error: nil,
		Body:  payload,
	}
	encodedResponse, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Printf("Built response: %v\n", string(encodedResponse))

	respondWithJSON(w, code, encodedResponse)
}

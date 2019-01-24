package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/blinky-z/Blog/server/handler"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

var (
	client = &http.Client{}

	loginUsername string
	loginEmail    string
	loginPassword string

	authToken string
)

type ResponseWithError struct {
	Error handler.PostErrorCode
	Body  interface{}
}

// helpful API for testing

// API for matching status code and error message of responses

// checkErrorResponse - check response that should return error message in response body
func checkErrorResponse(r *http.Response, expectedStatusCode int, expectedErrorMessage handler.PostErrorCode) {
	var response ResponseWithError
	decodeErrorResponse(r.Body, &response)
	if r.StatusCode != expectedStatusCode || response.Error != expectedErrorMessage {
		panic(fmt.Sprintf("Test Error. Received Error code: %d. Error message: %s\n"+
			"Expected Error code: %d. Error message: %s",
			r.StatusCode, response.Error, expectedStatusCode, expectedErrorMessage))
	}
}

// checkNiceResponse - check response that should return only status code
// If received status code does not match expected status code, then get Error Message from response body and print it
func checkNiceResponse(r *http.Response, expectedStatusCode int) {
	if r.StatusCode != expectedStatusCode {
		var response ResponseWithError
		decodeErrorResponse(r.Body, &response)

		panic(fmt.Sprintf("Test Error. Received Error code: %d. Expected Error code: %d\nError message: %s",
			r.StatusCode, expectedStatusCode, response.Error))
	}
}

// -----------

// API for encoding and decoding messages
func encodeMessage(message interface{}) []byte {
	encodedMessage, err := json.Marshal(message)
	if err != nil {
		panic(fmt.Sprintf("Error encoding message.\nMessage: %v\n. Error: %s", message, err))
	}

	return encodedMessage
}

func decodeErrorResponse(responseBody io.ReadCloser, resp *ResponseWithError) {
	bodyBytes, _ := ioutil.ReadAll(responseBody)

	responseBody = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	err := json.NewDecoder(responseBody).Decode(resp)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body.\nBody: %s\n. Error: %s", string(bodyBytes), err))
	}
}

// -----------
// Tests

func TestRunServer(t *testing.T) {
	go RunServer("testConfig", ".")
	for {
		resp, err := http.Get("http://" + Address + "/hc")
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// Register test user
func TestRegisterUser(t *testing.T) {
	loginUsername = uuid.New().String()
	loginEmail = loginUsername + "@gmail.com"
	loginPassword = uuid.New().String() + "Z"

	r := registerUser(loginUsername, loginEmail, loginPassword)

	checkNiceResponse(r, http.StatusOK)
}

// Log In with test user
func TestLoginUserWithEmail(t *testing.T) {
	r := loginUser("", loginEmail, loginPassword)

	bodyBytes, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	checkNiceResponse(r, http.StatusAccepted)

	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	var response ResponseLogIn
	decodeAuthResponse(r.Body, &response)

	authToken = response.Body
}

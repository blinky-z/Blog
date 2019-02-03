package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/blinky-z/Blog/handler/api"
	"github.com/blinky-z/Blog/models"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var (
	client = &http.Client{}

	loginUsername string
	loginEmail    string
	loginPassword string

	authToken string
	ctxCookie *http.Cookie
)

type ResponseWithError struct {
	Error api.PostErrorCode
	Body  interface{}
}

type AdminsConfig struct {
	Admins []models.User `json:"admins"`
}

// helpful API for testing

func setNewAuthData(r *http.Response) {
	var response ResponseLogin
	decodeAuthResponse(r.Body, &response)

	authToken = response.Body
	cookies := r.Cookies()
	for _, currentCookie := range cookies {
		if currentCookie.Name == "Secure-Fgp" {
			ctxCookie = currentCookie
		}
	}
}

// API for matching status code and error message of responses

// checkErrorResponse - check response that should return error message in response body
func checkErrorResponse(r *http.Response, expectedStatusCode int, expectedErrorMessage api.PostErrorCode) {
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

func init() {
	testConfigFile := flag.String("config", "configs/testConfig.json", "server config file for tests")
	flag.Parse()

	loginUsername = uuid.New().String()
	loginEmail = loginUsername + "@gmail.com"
	loginPassword = uuid.New().String() + "Z"

	admins := &AdminsConfig{Admins: []models.User{{Login: loginUsername}}}
	encodedAdmins := encodeMessage(admins)

	testAdminsFile := "configs/testAdmins.json"
	adminsFile, err := os.OpenFile(testAdminsFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		panic("Error opening admins config file")
	}
	_, err = adminsFile.Write(encodedAdmins)
	if err != nil {
		panic("Error writing admin to tests admins config file")
	}

	go RunServer(*testConfigFile, testAdminsFile)
	for {
		resp, err := http.Get("http://" + Address + "/api/hc")
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Register test user
	{
		r := registerUser(loginUsername, loginEmail, loginPassword)

		checkNiceResponse(r, http.StatusOK)
	}

	// Log In with registered test user and save token and fingerprint
	{
		r := loginUser("", loginEmail, loginPassword)

		checkNiceResponse(r, http.StatusAccepted)

		setNewAuthData(r)
	}
}

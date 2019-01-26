package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/blinky-z/Blog/server/handler"
	"github.com/blinky-z/Blog/server/models"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
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
	Error handler.PostErrorCode
	Body  interface{}
}

type AdminsConfig struct {
	Admins []models.User `json:"admins"`
}

// helpful API for testing

func setNewAuthData(r *http.Response) {
	var response ResponseLogIn
	decodeAuthResponse(r.Body, &response)

	authToken = response.Body
	cookies := r.Cookies()
	for _, currentCookie := range cookies {
		if currentCookie.Name == "__Secure-Fgp" {
			ctxCookie = currentCookie
		}
	}
}

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
	loginUsername = uuid.New().String()
	loginEmail = loginUsername + "@gmail.com"
	loginPassword = uuid.New().String() + "Z"

	admins := &AdminsConfig{Admins: []models.User{{Login: loginUsername}}}
	encodedAdmins := encodeMessage(admins)

	viper.SetConfigName("testConfig")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		logError.Fatalf("Fatal error reading config file: %s \n", err)
	}

	adminsConfigFile := viper.GetString("adminsConfigFile")

	f, err := os.OpenFile(adminsConfigFile+".json", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic("Error opening admins config file")
	}
	_, err = f.Write(encodedAdmins)
	if err != nil {
		panic("Error writing new test admin to admins config file")
	}

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
func TestRegisterAdmin(t *testing.T) {
	r := registerUser(loginUsername, loginEmail, loginPassword)

	checkNiceResponse(r, http.StatusOK)
}

// Log In with registered in prev test test user
func TestLoginAdminWithEmail(t *testing.T) {
	r := loginUser("", loginEmail, loginPassword)

	checkNiceResponse(r, http.StatusAccepted)

	setNewAuthData(r)
}

package main

// Authorization system tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/blinky-z/Blog/server/handler"
	"github.com/blinky-z/Blog/server/models"
	"io"
	"net/http"
	"strings"
	"testing"
)

type ResponseLogIn struct {
	Error handler.PostErrorCode
	Body  string
}

// -----------

// API for encoding and decoding messages

func decodeAuthResponse(responseBody io.ReadCloser, response *ResponseLogIn) {
	err := json.NewDecoder(responseBody).Decode(&response)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
}

// -----------
// Helpful API for sending authorization http requests

func sendAuthUserMessage(address string, credentials models.User) *http.Response {
	encodedCredentials := encodeMessage(credentials)

	request, err := http.NewRequest("GET", address, bytes.NewReader(encodedCredentials))
	if err != nil {
		panic(fmt.Sprintf("Can not create request. Error: %s", err))
	}
	request.Header.Set("Content-Type", "application/json")

	r, err := client.Do(request)
	if err != nil {
		panic(fmt.Sprintf("Can not send request. Error: %s", err))
	}

	return r
}

func registerUser(login, email, password string) *http.Response {
	registrationCredentials := models.User{Login: login, Email: email, Password: password}

	return sendAuthUserMessage("http://"+Address+"/user/register", registrationCredentials)
}

func loginUser(login, email, password string) *http.Response {
	loginCredentials := models.User{Login: login, Email: email, Password: password}

	return sendAuthUserMessage("http://"+Address+"/user/login", loginCredentials)
}

// -----------
// tests

// Registration tests

func TestRegisterAlreadyRegisteredUser(t *testing.T) {
	username := strings.Repeat("a", handler.MaxLoginLen*2)

	r := registerUser(username, loginEmail, loginPassword)

	checkErrorResponse(r, http.StatusBadRequest, handler.InvalidLogin)
}

func TestRegisterUserWithTooLongUsername(t *testing.T) {
	username := strings.Repeat("a", handler.MaxLoginLen*2)

	r := registerUser(username, loginEmail, loginPassword)

	checkErrorResponse(r, http.StatusBadRequest, handler.InvalidLogin)
}

func TestRegisterUserWithTooShortUsername(t *testing.T) {
	username := strings.Repeat("a", handler.MinLoginLen/2)

	r := registerUser(username, loginEmail, loginPassword)

	checkErrorResponse(r, http.StatusBadRequest, handler.InvalidLogin)
}

func TestRegisterUserWithTooLongPassword(t *testing.T) {
	password := strings.Repeat("A", handler.MaxPwdLen*2)

	r := registerUser(loginUsername, loginEmail, password)

	checkErrorResponse(r, http.StatusBadRequest, handler.InvalidPassword)
}

func TestRegisterUserWithTooShortPassword(t *testing.T) {
	password := strings.Repeat("A", handler.MinPwdLen/2)

	r := registerUser(loginUsername, loginEmail, password)

	checkErrorResponse(r, http.StatusBadRequest, handler.InvalidPassword)
}

func TestRegisterUserWithIdenticalPassword(t *testing.T) {
	password := strings.Repeat("A", handler.MinPwdLen)
	login := password

	r := registerUser(login, loginEmail, password)

	checkErrorResponse(r, http.StatusBadRequest, handler.InvalidPassword)
}

func TestRegisterUserWithEmptyUsername(t *testing.T) {
	r := registerUser("", loginEmail, loginPassword)

	checkErrorResponse(r, http.StatusBadRequest, handler.IncompleteCredentials)
}

func TestRegisterUserWithEmptyEmail(t *testing.T) {
	r := registerUser(loginUsername, "", loginPassword)

	checkErrorResponse(r, http.StatusBadRequest, handler.IncompleteCredentials)
}

func TestRegisterUserWithEmptyPassword(t *testing.T) {
	r := registerUser(loginUsername, loginEmail, "")

	checkErrorResponse(r, http.StatusBadRequest, handler.IncompleteCredentials)
}

func TestRegisterUserWithBadRequestBody(t *testing.T) {
	message := `{bad request body}`

	request, err := http.NewRequest("GET", "http://"+Address+"/user/register", strings.NewReader(message))
	if err != nil {
		panic(fmt.Sprintf("Can not create request. Error: %s", err))
	}
	request.Header.Set("Content-Type", "application/json")

	r, err := client.Do(request)
	if err != nil {
		panic(fmt.Sprintf("Can not send request. Error: %s", err))
	}

	checkErrorResponse(r, http.StatusBadRequest, handler.BadRequestBody)
}

// Log In tests

func TestLoginUserWithUsername(t *testing.T) {
	r := loginUser(loginUsername, "", loginPassword)

	checkNiceResponse(r, http.StatusAccepted)
}

func TestLoginUserWithWrongUsername(t *testing.T) {
	r := loginUser("abcd", "", loginPassword)

	checkErrorResponse(r, http.StatusUnauthorized, handler.WrongCredentials)
}

func TestLoginUserWithWrongEmail(t *testing.T) {
	r := loginUser("", "abcd@gmail.com", loginPassword)

	checkErrorResponse(r, http.StatusUnauthorized, handler.WrongCredentials)
}

func TestLoginUserWithWrongPassword(t *testing.T) {
	r := loginUser("", loginEmail, "abcd")

	checkErrorResponse(r, http.StatusUnauthorized, handler.WrongCredentials)
}

func TestLoginUserWithEmptyLoginAndEmail(t *testing.T) {
	r := loginUser("", "", "abcd")

	checkErrorResponse(r, http.StatusBadRequest, handler.IncompleteCredentials)
}

func TestLoginUserWithEmptyPassword(t *testing.T) {
	r := loginUser("", loginEmail, "")

	checkErrorResponse(r, http.StatusBadRequest, handler.IncompleteCredentials)
}

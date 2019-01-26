package main

// Authorization system tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/blinky-z/Blog/server/handler"
	"github.com/blinky-z/Blog/server/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
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

func TestRegisterUserWithExistingEmail(t *testing.T) {
	username := uuid.New().String()

	r := registerUser(username, loginEmail, loginPassword)

	checkErrorResponse(r, http.StatusBadRequest, handler.AlreadyRegistered)
}

func TestRegisterUserWithExistingLogin(t *testing.T) {
	email := uuid.New().String() + "@gmail.com"

	r := registerUser(loginUsername, email, loginPassword)

	checkErrorResponse(r, http.StatusBadRequest, handler.AlreadyRegistered)
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

func TestRegisterUserWhereLoginIdenticalPassword(t *testing.T) {
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
	r := loginUser("abcdef", "", loginPassword)

	checkErrorResponse(r, http.StatusUnauthorized, handler.WrongCredentials)
}

func TestLoginUserWithWrongEmail(t *testing.T) {
	r := loginUser("", "abcd@gmail.com", loginPassword)

	checkErrorResponse(r, http.StatusUnauthorized, handler.WrongCredentials)
}

func TestLoginUserWithWrongPassword(t *testing.T) {
	r := loginUser("", loginEmail, "abcde1fZ")

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

// Test JWT tokens
// Test creating, updating, deleting posts with not admin role

func TestRegisterNotAdmin(t *testing.T) {
	loginUsername = uuid.New().String()
	loginEmail = loginUsername + "@gmail.com"
	loginPassword = uuid.New().String() + "Z"

	r := registerUser(loginUsername, loginEmail, loginPassword)

	checkNiceResponse(r, http.StatusOK)
}

func TestLoginNotAdmin(t *testing.T) {
	r := loginUser("", loginEmail, loginPassword)

	checkNiceResponse(r, http.StatusAccepted)

	var response ResponseLogIn
	decodeAuthResponse(r.Body, &response)

	authToken = response.Body
}

func TestCreatePostWithNotAdminUser(t *testing.T) {
	var post models.Post
	post.Title = "Title1"
	post.Content = "Content1 Content2 Content3"

	r := createPost(post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusForbidden, handler.NoPermissions)
}

func TestUpdatePostWithNotAdminUser(t *testing.T) {
	var newPost models.Post
	newPost.Title = "Title1"
	newPost.Content = "Content1 Content2 Content3"

	r := updatePost("1", newPost)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusForbidden, handler.NoPermissions)
}

func TestDeletePostWithNotAdminUser(t *testing.T) {
	r := deletePost("1")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusForbidden, handler.NoPermissions)
}

func TestCreateInvalidJwtToken(t *testing.T) {
	var claims models.TokenClaims
	claims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
	claims.Role = "admin"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("wrongSigningKey"))
	if err != nil {
		t.Fatalf("Error creating invalid JWT Token. Error: %v", err)
		return
	}

	authToken = tokenString
}

func TestCreatePostWithInvalidJwtToken(t *testing.T) {
	var post models.Post
	post.Title = "Title1"
	post.Content = "Content1 Content2 Content3"

	r := createPost(post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusUnauthorized, handler.InvalidToken)
}

func TestUpdatePostWithInvalidJwtToken(t *testing.T) {
	var post models.Post
	post.Title = "Title1"
	post.Content = "Content1 Content2 Content3"

	r := updatePost("1", post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusUnauthorized, handler.InvalidToken)
}

func TestDeletePostWithInvalidJwtToken(t *testing.T) {
	r := deletePost("1")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusUnauthorized, handler.InvalidToken)
}

func TestCreatePostWithMissingToken(t *testing.T) {
	authToken = ""

	var post models.Post
	post.Title = "Title1"
	post.Content = "Content1 Content2 Content3"

	r := createPost(post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusUnauthorized, handler.InvalidToken)
}

func TestCreatePostWithMissingAuthorizationHeader(t *testing.T) {
	var post models.Post
	post.Title = "Title1"
	post.Content = "Content1 Content2 Content3"

	encodedMessage := encodeMessage(post)

	request, err := http.NewRequest("POST", "http://"+Address+"/posts", bytes.NewReader(encodedMessage))
	if err != nil {
		panic(fmt.Sprintf("Can not create request. Error: %s", err))
	}
	request.Header.Set("Content-Type", "application/json")

	r, err := client.Do(request)
	if err != nil {
		panic(fmt.Sprintf("Can not send request. Error: %s", err))
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusUnauthorized, handler.MissingToken)
}

package main

// Authorization system tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/blinky-z/Blog/handler/api"
	"github.com/blinky-z/Blog/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type ResponseLogin struct {
	Error api.PostErrorCode
	Body  string
}

// -----------

// API for encoding and decoding messages

func decodeAuthResponse(responseBody io.ReadCloser, response *ResponseLogin) {
	err := json.NewDecoder(responseBody).Decode(&response)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
}

// -----------
// Helpful API for sending authorization http requests

func sendAuthUserMessage(address string, credentials interface{}) *http.Response {
	encodedCredentials := encodeMessage(credentials)

	request, err := http.NewRequest("POST", address, bytes.NewReader(encodedCredentials))
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
	registrationCredentials := models.RegistrationRequest{Username: login, Email: email, Password: password}

	return sendAuthUserMessage("http://"+Address+"/api/user/register", registrationCredentials)
}

func loginUser(login, email, password string) *http.Response {
	loginCredentials := models.LoginRequest{Username: login, Email: email, Password: password}

	return sendAuthUserMessage("http://"+Address+"/api/user/login", loginCredentials)
}

// -----------
// tests

// Registration tests

func TestRegisterUserWithExistingEmail(t *testing.T) {
	username := uuid.New().String()

	r := registerUser(username, loginEmail, loginPassword)

	checkErrorResponse(r, http.StatusBadRequest, api.AlreadyRegistered)
}

func TestRegisterUserWithExistingLogin(t *testing.T) {
	email := uuid.New().String() + "@gmail.com"

	r := registerUser(loginUsername, email, loginPassword)

	checkErrorResponse(r, http.StatusBadRequest, api.AlreadyRegistered)
}

func TestRegisterUserWithTooLongUsername(t *testing.T) {
	username := strings.Repeat("a", api.MaxLoginLen*2)

	r := registerUser(username, loginEmail, loginPassword)

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidLogin)
}

func TestRegisterUserWithTooShortUsername(t *testing.T) {
	username := strings.Repeat("a", api.MinLoginLen/2)

	r := registerUser(username, loginEmail, loginPassword)

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidLogin)
}

func TestRegisterUserWithTooLongPassword(t *testing.T) {
	password := strings.Repeat("A", api.MaxPwdLen*2)

	r := registerUser(loginUsername, loginEmail, password)

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidPassword)
}

func TestRegisterUserWithTooShortPassword(t *testing.T) {
	password := strings.Repeat("A", api.MinPwdLen/2)

	r := registerUser(loginUsername, loginEmail, password)

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidPassword)
}

func TestRegisterUserWithEmptyUsername(t *testing.T) {
	r := registerUser("", loginEmail, loginPassword)

	checkErrorResponse(r, http.StatusBadRequest, api.IncompleteCredentials)
}

func TestRegisterUserWithEmptyEmail(t *testing.T) {
	r := registerUser(loginUsername, "", loginPassword)

	checkErrorResponse(r, http.StatusBadRequest, api.IncompleteCredentials)
}

func TestRegisterUserWithEmptyPassword(t *testing.T) {
	r := registerUser(loginUsername, loginEmail, "")

	checkErrorResponse(r, http.StatusBadRequest, api.IncompleteCredentials)
}

func TestRegisterUserWithBadRequestBody(t *testing.T) {
	message := `{bad request body}`

	request, err := http.NewRequest("POST", "http://"+Address+"/api/user/register", strings.NewReader(message))
	if err != nil {
		panic(fmt.Sprintf("Can not create request. Error: %s", err))
	}
	request.Header.Set("Content-Type", "application/json")

	r, err := client.Do(request)
	if err != nil {
		panic(fmt.Sprintf("Can not send request. Error: %s", err))
	}

	checkErrorResponse(r, http.StatusBadRequest, api.BadRequestBody)
}

// Log In tests

func TestLoginUserWithUsername(t *testing.T) {
	r := loginUser(loginUsername, "", loginPassword)

	checkNiceResponse(r, http.StatusAccepted)
}

func TestLoginUserWithWrongUsername(t *testing.T) {
	r := loginUser("abcdef", "", loginPassword)

	checkErrorResponse(r, http.StatusUnauthorized, api.WrongCredentials)
}

func TestLoginUserWithWrongEmail(t *testing.T) {
	r := loginUser("", "abcd@gmail.com", loginPassword)

	checkErrorResponse(r, http.StatusUnauthorized, api.WrongCredentials)
}

func TestLoginUserWithWrongPassword(t *testing.T) {
	r := loginUser("", loginEmail, "abcde1fZ")

	checkErrorResponse(r, http.StatusUnauthorized, api.WrongCredentials)
}

func TestLoginUserWithEmptyLoginAndEmail(t *testing.T) {
	r := loginUser("", "", "abcd")

	checkErrorResponse(r, http.StatusBadRequest, api.IncompleteCredentials)
}

func TestLoginUserWithEmptyPassword(t *testing.T) {
	r := loginUser("", loginEmail, "")

	checkErrorResponse(r, http.StatusBadRequest, api.IncompleteCredentials)
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

func TestLoginWithNotAdmin(t *testing.T) {
	r := loginUser("", loginEmail, loginPassword)

	checkNiceResponse(r, http.StatusAccepted)

	setNewAuthData(r)
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

	checkErrorResponse(r, http.StatusForbidden, api.NoPermissions)
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

	checkErrorResponse(r, http.StatusForbidden, api.NoPermissions)
}

func TestDeletePostWithNotAdminUser(t *testing.T) {
	r := deletePost("1")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusForbidden, api.NoPermissions)
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

	checkErrorResponse(r, http.StatusUnauthorized, api.InvalidToken)
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

	checkErrorResponse(r, http.StatusUnauthorized, api.InvalidToken)
}

func TestDeletePostWithInvalidJwtToken(t *testing.T) {
	r := deletePost("1")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusUnauthorized, api.InvalidToken)
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

	checkErrorResponse(r, http.StatusUnauthorized, api.InvalidToken)
}

func TestCreatePostWithMissingAuthorizationHeader(t *testing.T) {
	var post models.Post
	post.Title = "Title1"
	post.Content = "Content1 Content2 Content3"

	encodedMessage := encodeMessage(post)

	request, err := http.NewRequest("POST", "http://"+Address+"/api/posts", bytes.NewReader(encodedMessage))
	if err != nil {
		panic(fmt.Sprintf("Can not create request. Error: %s", err))
	}
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(ctxCookie)

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

	checkErrorResponse(r, http.StatusUnauthorized, api.InvalidToken)
}

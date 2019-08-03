package tests

import (
	"bytes"
	"fmt"
	"github.com/blinky-z/Blog/handler/restApi"
	"github.com/blinky-z/Blog/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"testing"
	"time"
)

// Registration tests

func TestRegisterUserWithExistingEmail(t *testing.T) {
	username := uuid.New().String()

	r := registerUser(username, loginEmail, loginPassword)

	assertErrorResponse(r, http.StatusBadRequest, restApi.UserAlreadyRegistered)
}

func TestRegisterUserWithExistingLogin(t *testing.T) {
	email := uuid.New().String() + "@gmail.com"

	r := registerUser(loginUsername, email, loginPassword)

	assertErrorResponse(r, http.StatusBadRequest, restApi.UserAlreadyRegistered)
}

func TestRegisterUserWithTooLongUsername(t *testing.T) {
	username := strings.Repeat("a", restApi.MaxUsernameLen*2)

	r := registerUser(username, loginEmail, loginPassword)

	assertErrorResponse(r, http.StatusBadRequest, restApi.InvalidUsername)
}

func TestRegisterUserWithTooShortUsername(t *testing.T) {
	username := strings.Repeat("a", restApi.MinUsernameLen/2)

	r := registerUser(username, loginEmail, loginPassword)

	assertErrorResponse(r, http.StatusBadRequest, restApi.InvalidUsername)
}

func TestRegisterUserWithTooLongPassword(t *testing.T) {
	password := strings.Repeat("A", restApi.MaxPwdLen*2)

	r := registerUser(loginUsername, loginEmail, password)

	assertErrorResponse(r, http.StatusBadRequest, restApi.InvalidPassword)
}

func TestRegisterUserWithTooShortPassword(t *testing.T) {
	password := strings.Repeat("A", restApi.MinPwdLen/2)

	r := registerUser(loginUsername, loginEmail, password)

	assertErrorResponse(r, http.StatusBadRequest, restApi.InvalidPassword)
}

func TestRegisterUserWithEmptyUsername(t *testing.T) {
	r := registerUser("", loginEmail, loginPassword)

	assertErrorResponse(r, http.StatusBadRequest, restApi.IncompleteCredentials)
}

func TestRegisterUserWithEmptyEmail(t *testing.T) {
	r := registerUser(loginUsername, "", loginPassword)

	assertErrorResponse(r, http.StatusBadRequest, restApi.IncompleteCredentials)
}

func TestRegisterUserWithEmptyPassword(t *testing.T) {
	r := registerUser(loginUsername, loginEmail, "")

	assertErrorResponse(r, http.StatusBadRequest, restApi.IncompleteCredentials)
}

func TestRegisterUserWithBadRequestBody(t *testing.T) {
	message := `{bad request body}`

	request, err := http.NewRequest("POST", "http://"+address+"/api/user/register", strings.NewReader(message))
	if err != nil {
		panic(fmt.Sprintf("Can not create request. Error: %s", err))
	}
	request.Header.Set("Content-Type", "application/json")

	r, err := client.Do(request)
	if err != nil {
		panic(fmt.Sprintf("Can not send request. Error: %s", err))
	}

	assertErrorResponse(r, http.StatusBadRequest, restApi.BadRequestBody)
}

// Log In tests

func TestLoginUserWithUsername(t *testing.T) {
	r := loginUser(loginUsername, "", loginPassword)

	assertNiceResponse(r, http.StatusOK)
}

func TestLoginUserWithWrongUsername(t *testing.T) {
	r := loginUser("abcdef", "", loginPassword)

	assertErrorResponse(r, http.StatusUnauthorized, restApi.WrongCredentials)
}

func TestLoginUserWithWrongEmail(t *testing.T) {
	r := loginUser("", "abcd@gmail.com", loginPassword)

	assertErrorResponse(r, http.StatusUnauthorized, restApi.WrongCredentials)
}

func TestLoginUserWithWrongPassword(t *testing.T) {
	r := loginUser("", loginEmail, "abcde1fZ")

	assertErrorResponse(r, http.StatusUnauthorized, restApi.WrongCredentials)
}

func TestLoginUserWithEmptyLoginAndEmail(t *testing.T) {
	r := loginUser("", "", "abcd")

	assertErrorResponse(r, http.StatusBadRequest, restApi.IncompleteCredentials)
}

func TestLoginUserWithEmptyPassword(t *testing.T) {
	r := loginUser("", loginEmail, "")

	assertErrorResponse(r, http.StatusBadRequest, restApi.IncompleteCredentials)
}

// Test JWT tokens

func TestRegisterNotAdmin(t *testing.T) {
	loginUsername = uuid.New().String()
	loginEmail = loginUsername + "@gmail.com"
	loginPassword = uuid.New().String() + "Z"

	r := registerUser(loginUsername, loginEmail, loginPassword)

	assertNiceResponse(r, http.StatusOK)
}

func TestInitNotAdminUser(t *testing.T) {
	r := loginUser("", loginEmail, loginPassword)

	assertNiceResponse(r, http.StatusOK)

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

	assertErrorResponse(r, http.StatusForbidden, restApi.NoPermissions)
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

	assertErrorResponse(r, http.StatusForbidden, restApi.NoPermissions)
}

func TestDeletePostWithNotAdminUser(t *testing.T) {
	r := deletePost("1")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	assertErrorResponse(r, http.StatusForbidden, restApi.NoPermissions)
}

func TestInitInvalidJwtToken(t *testing.T) {
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

	assertErrorResponse(r, http.StatusUnauthorized, restApi.InvalidToken)
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

	assertErrorResponse(r, http.StatusUnauthorized, restApi.InvalidToken)
}

func TestDeletePostWithInvalidJwtToken(t *testing.T) {
	r := deletePost("1")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	assertErrorResponse(r, http.StatusUnauthorized, restApi.InvalidToken)
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

	assertErrorResponse(r, http.StatusUnauthorized, restApi.InvalidToken)
}

func TestCreatePostWhenMissingAuthorizationHeader(t *testing.T) {
	var post models.Post
	post.Title = "Title1"
	post.Content = "Content1 Content2 Content3"

	encodedMessage := encodeMessageIntoJson(post)

	request, err := http.NewRequest("POST", "http://"+address+"/api/posts", bytes.NewReader(encodedMessage))
	if err != nil {
		panic(fmt.Sprintf("Can not create request. Error: %s", err))
	}
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(fgpCookie)

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

	assertErrorResponse(r, http.StatusUnauthorized, restApi.InvalidToken)
}

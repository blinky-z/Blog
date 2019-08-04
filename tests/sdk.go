package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/blinky-z/Blog/handler/restapi"
	"github.com/blinky-z/Blog/models"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
)

var (
	client = &http.Client{}

	loginUsername string
	loginEmail    string
	loginPassword string

	authToken string
	fgpCookie *http.Cookie
	address   string
	db        *sql.DB
)

// Response - generic struct for storing deserialized response body
type Response struct {
	Error interface{}
	Body  interface{}
}

// ResponseWithPost - struct for storing returned post
type ResponseWithPost struct {
	Error interface{}
	Body  models.Post
}

// ResponseWithCertainPost - struct for storing returned post with comments
type ResponseWithCertainPost struct {
	Error interface{}
	Body  models.CertainPost
}

// ResponseWithRangeOfPosts - struct for storing returned range of posts
type ResponseWithRangeOfPosts struct {
	Error interface{}
	Body  []models.Post
}

// ResponseWithComment - struct for storing returned comment
type ResponseWithComment struct {
	Error interface{}
	Body  models.Comment
}

// -----------
// Internal helper methods

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func generateRandomAlphanumericString(l int) string {
	b := make([]rune, l)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// decodeResponse - usually you don't want to use this function as it will deserialize JSON into generic
// Response type, but you want to use deserializing functions for concrete types
func decodeResponse(responseBody io.ReadCloser) *Response {
	bodyBytes, _ := ioutil.ReadAll(responseBody)
	responseBodyCopy := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	resp := &Response{}
	err := json.NewDecoder(responseBodyCopy).Decode(resp)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
	return resp
}

// decodeResponseWithPostBody - use this function to deserialize response that contains blog post
func decodeResponseWithPostBody(responseBody io.ReadCloser) *ResponseWithPost {
	bodyBytes, _ := ioutil.ReadAll(responseBody)
	responseBodyCopy := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	resp := &ResponseWithPost{}
	err := json.NewDecoder(responseBodyCopy).Decode(resp)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
	return resp
}

// decodeResponseWithCertainPostBody - use this function to deserialize response that contains blog post with comments
func decodeResponseWithCertainPostBody(responseBody io.ReadCloser) *ResponseWithCertainPost {
	bodyBytes, _ := ioutil.ReadAll(responseBody)
	responseBodyCopy := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	resp := &ResponseWithCertainPost{}
	err := json.NewDecoder(responseBodyCopy).Decode(resp)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
	return resp
}

// decodeResponseWithRangeOfPostsBody - use this function to deserialize response that contains range of blog posts
func decodeResponseWithRangeOfPostsBody(responseBody io.ReadCloser) *ResponseWithRangeOfPosts {
	bodyBytes, _ := ioutil.ReadAll(responseBody)
	responseBodyCopy := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	resp := &ResponseWithRangeOfPosts{}
	err := json.NewDecoder(responseBodyCopy).Decode(resp)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
	return resp
}

// decodeResponseWithPostBody - use this function to deserialize response that contains comment
func decodeResponseWithCommentBody(responseBody io.ReadCloser) *ResponseWithComment {
	bodyBytes, _ := ioutil.ReadAll(responseBody)
	responseBodyCopy := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	resp := &ResponseWithComment{}
	err := json.NewDecoder(responseBodyCopy).Decode(resp)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
	return resp
}

// sendMessage - generic function for sending a request
// @method - supports "GET", "POST", "PUT", "DELETE"
// @address - http address to send request to. Example of address: "localhost:8080"
// @message - raw message to send in body. It will be serialized into JSON
// This function automatically sets required auth data for POST, PUT and DELETE methods
func sendMessage(method, address string, message interface{}, addAuthData bool) *http.Response {
	var request *http.Request
	var response *http.Response
	var err error

	encodedMessage := encodeMessageIntoJSON(message)
	body := bytes.NewReader(encodedMessage)

	switch method {
	case "GET":
		if request, err = http.NewRequest("GET", address, body); err != nil {
			break
		}
		response, err = client.Do(request)
	case "POST":
		if request, err = http.NewRequest("POST", address, body); err != nil {
			break
		}
		request.Header.Set("Content-Type", "application/json")
		if addAuthData {
			addAuthDataToRequest(request)
		}
		response, err = client.Do(request)
	case "PUT":
		if request, err = http.NewRequest("PUT", address, body); err != nil {
			break
		}
		request.Header.Set("Content-Type", "application/json")
		if addAuthData {
			addAuthDataToRequest(request)
		}

		response, err = client.Do(request)
	case "DELETE":
		if request, err = http.NewRequest("DELETE", address, body); err != nil {
			break
		}
		if addAuthData {
			addAuthDataToRequest(request)
		}

		response, err = client.Do(request)
	}
	if err != nil {
		panic(fmt.Sprintf("Error sending request. Error: %s", err))
	}

	return response
}

// -----------
// Auth managing: perform registration or log in request, extract/set auth data
// Auth data consists of Bearer JWT token and fingerprint cookie for now

func registerUser(login, email, password string) *http.Response {
	registrationCredentials := models.RegistrationRequest{Username: login, Email: email, Password: password}
	return sendMessage("POST", "http://"+address+"/api/user/register", registrationCredentials, false)
}

func loginUser(login, email, password string) *http.Response {
	loginCredentials := models.LoginRequest{Username: login, Email: email, Password: password}
	return sendMessage("POST", "http://"+address+"/api/user/login", loginCredentials, false)
}

func setNewAuthData(r *http.Response) {
	resp := decodeResponse(r.Body)

	authToken = resp.Body.(string)
	cookies := r.Cookies()
	for _, currentCookie := range cookies {
		if currentCookie.Name == "Secure-Fgp" {
			fgpCookie = currentCookie
		}
	}
}

func addAuthDataToRequest(r *http.Request) {
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	r.AddCookie(fgpCookie)
}

// API for matching response status code and error message

// assertErrorResponse - asserts that response has expected status code and error code
// Use this function for asserting only error responses
func assertErrorResponse(t *testing.T, r *http.Response, expectedStatusCode int, expectedErrorCode models.RequestErrorCode) {
	resp := decodeResponse(r.Body)

	receivedStatusCode := r.StatusCode
	if receivedStatusCode != expectedStatusCode {
		t.Fatalf("Received status code does not match expected one\n. Received: %d\nExpected: %d\n",
			receivedStatusCode, expectedStatusCode)
	}

	receivedErrorCode := resp.Error
	if receivedErrorCode == nil {
		t.Fatalf("Response should contain error code in the body, but was <nil>")
	}
	if receivedErrorCode != expectedErrorCode.Error() {
		t.Fatalf("Received error code does not match expected one\n. Received: %s\nExpected: %s\n",
			receivedErrorCode, expectedErrorCode)
	}
}

// assertNiceResponse - asserts that response has expected status code
// Use this function for asserting only non-error responses
func assertNiceResponse(t *testing.T, r *http.Response, expectedStatusCode int) {
	receivedStatusCode := r.StatusCode
	if receivedStatusCode != expectedStatusCode {
		panic(fmt.Sprintf("Received status code does not match expected one\n. Received: %d\nExpected: %d\n",
			receivedStatusCode, expectedStatusCode))
	}
}

// -----------
// API for encoding and decoding messages into/from JSON

func encodeMessageIntoJSON(message interface{}) []byte {
	encodedMessage, err := json.Marshal(message)
	if err != nil {
		panic(fmt.Sprintf("Error encoding message.\nMessage: %v\n. Error: %s", message, err))
	}

	return encodedMessage
}

// -----------
// blog posts related SDK

func comparePosts(l, r models.Post) bool {
	if l.ID != r.ID || l.Title != r.Title || l.Date != r.Date || l.Content != r.Content {
		return false
	}
	if l.Metadata.Description != r.Metadata.Description {
		return false
	}
	if len(l.Metadata.Keywords) != len(r.Metadata.Keywords) {
		return false
	}
	for keywordIndex := 0; keywordIndex < len(l.Metadata.Keywords); keywordIndex++ {
		if l.Metadata.Keywords[keywordIndex] != r.Metadata.Keywords[keywordIndex] {
			return false
		}
	}

	return true
}

// factory methods

func createPostRequestFactory() models.CreatePostRequest {
	return models.CreatePostRequest{
		Title:   generateRandomAlphanumericString(restapi.MinPostTitleLen),
		Author:  generateRandomAlphanumericString(restapi.MinUsernameLen),
		Snippet: generateRandomAlphanumericString(restapi.MinSnippetLen),
		Content: generateRandomAlphanumericString(restapi.MinSnippetLen),
		Metadata: models.MetaData{
			Description: generateRandomAlphanumericString(restapi.MinMetaDescriptionLen),
			Keywords:    []string{generateRandomAlphanumericString(restapi.MinMetaKeywordLen)},
		},
	}
}

func updatePostRequestFactory() models.UpdatePostRequest {
	return models.UpdatePostRequest{
		Title:   generateRandomAlphanumericString(restapi.MinPostTitleLen),
		Author:  generateRandomAlphanumericString(restapi.MinUsernameLen),
		Snippet: generateRandomAlphanumericString(restapi.MinSnippetLen),
		Content: generateRandomAlphanumericString(restapi.MinSnippetLen),
		Metadata: models.MetaData{
			Description: generateRandomAlphanumericString(restapi.MinMetaDescriptionLen),
			Keywords:    []string{generateRandomAlphanumericString(restapi.MinMetaKeywordLen)},
		},
	}
}

// blog posts related rest api access

func getCertainPost(postID string) *http.Response {
	return sendMessage("GET", "http://"+address+"/api/posts/"+postID, "", false)
}

func getPostsInRange(page, postsPerPage string) *http.Response {
	if postsPerPage == "" {
		return sendMessage("GET", "http://"+address+"/api/posts?page="+page, "", false)
	}
	return sendMessage("GET", "http://"+address+"/api/posts?page="+page+"&posts-per-page="+postsPerPage, "", false)
}

// pass models.CreatePostRequest if you don't want to test bad body
func createPost(message interface{}) *http.Response {
	return sendMessage("POST", "http://"+address+"/api/posts", message, true)
}

// pass models.UpdatePostRequest if you don't want to test bad body
func updatePost(postID string, message interface{}) *http.Response {
	return sendMessage("PUT", "http://"+address+"/api/posts/"+postID, message, true)
}

func deletePost(postID string) *http.Response {
	return sendMessage("DELETE", "http://"+address+"/api/posts/"+postID, "", true)
}

// -----------
// comments related SDK

// factory methods

func createCommentRequestFactory(postID string) models.CreateCommentRequest {
	return models.CreateCommentRequest{
		PostID:          postID,
		ParentCommentID: nil,
		Author:          generateRandomAlphanumericString(restapi.MinUsernameLen),
		Content:         generateRandomAlphanumericString(restapi.MinCommentContentLen),
	}
}

func createCommentWithParentRequestFactory(postID string, parentCommentId string) models.CreateCommentRequest {
	request := createCommentRequestFactory(postID)
	request.ParentCommentID = parentCommentId
	return request
}

func updateCommentRequestFactory() models.UpdateCommentRequest {
	return models.UpdateCommentRequest{
		Content: generateRandomAlphanumericString(restapi.MinCommentContentLen),
	}
}

// comments related rest api access

func createComment(message interface{}) *http.Response {
	return sendMessage("POST", "http://"+address+"/api/comments", message, true)
}

func updateComment(id string, message interface{}) *http.Response {
	return sendMessage("PUT", "http://"+address+"/api/comments/"+id, message, true)
}

func deleteComment(id string) *http.Response {
	return sendMessage("DELETE", "http://"+address+"/api/comments/"+id, "", true)
}

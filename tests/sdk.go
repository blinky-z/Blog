package tests

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/blinky-z/Blog/handler/restApi"
	"github.com/blinky-z/Blog/models"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
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

// internal helper methods

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func generateRandomAlphanumericString(l int) string {
	b := make([]rune, l)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type ResponseWithError struct {
	Error restApi.RequestErrorCode
	Body  interface{}
}

type ResponseLogin struct {
	Error restApi.RequestErrorCode
	Body  string
}

// helpful API for testing

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

	return sendAuthUserMessage("http://"+address+"/api/user/register", registrationCredentials)
}

func loginUser(login, email, password string) *http.Response {
	loginCredentials := models.LoginRequest{Username: login, Email: email, Password: password}

	return sendAuthUserMessage("http://"+address+"/api/user/login", loginCredentials)
}

func setNewAuthData(r *http.Response) {
	var response ResponseLogin
	decodeAuthResponse(r.Body, &response)

	authToken = response.Body
	cookies := r.Cookies()
	for _, currentCookie := range cookies {
		if currentCookie.Name == "Secure-Fgp" {
			fgpCookie = currentCookie
		}
	}
}

// Common helpful functions
// API for matching status code and error message of responses

// checkErrorResponse - check response that should return error message in response body
func checkErrorResponse(r *http.Response, expectedStatusCode int, expectedErrorMessage restApi.RequestErrorCode) {
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

// API for encoding and decoding messages into/from JSON
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
// blog posts related API

type ResponsePost struct {
	Error restApi.RequestErrorCode
	Body  models.Post
}

type ResponsePostWithComments struct {
	Error restApi.RequestErrorCode
	Body  models.CertainPostResponse
}

type ResponseRangePosts struct {
	Error restApi.RequestErrorCode
	Body  []models.Post
}

// -----------
// Work with posts
func comparePosts(l models.Post, r models.Post) bool {
	if l.ID != r.ID || l.Title != r.Title || l.Date != r.Date || l.Content != r.Content {
		return false
	}

	if l.Metadata.Description != r.Metadata.Description {
		return false
	}

	if len(l.Metadata.Keywords) != len(r.Metadata.Keywords) {
		return false
	}
	for currentKeywordNum := 0; currentKeywordNum < len(l.Metadata.Keywords); currentKeywordNum += 1 {
		if l.Metadata.Keywords[currentKeywordNum] != r.Metadata.Keywords[currentKeywordNum] {
			return false
		}
	}

	return true
}

func comparePostLists(lList, rList []models.Post) bool {
	if len(lList) != len(rList) {
		return false
	}

	for currentPost := 0; currentPost < len(lList); currentPost++ {
		if !comparePosts(lList[currentPost], rList[currentPost]) {
			return false
		}
	}

	return true
}

func testPostFactory() models.Post {
	var testPost models.Post
	title := make([]byte, restApi.MinPostTitleLen)
	author := make([]byte, restApi.MinUsernameLen)
	snippet := make([]byte, restApi.MinSnippetLen)
	content := make([]byte, restApi.MinSnippetLen)
	metaDesc := make([]byte, restApi.MinMetaDescriptionLen)
	rand.Read(author)
	rand.Read(snippet)
	rand.Read(content)
	rand.Read(metaDesc)
	metaKeywords := make([][]byte, restApi.MinMetaKeywordsAmount)
	for keywordIndex := range metaKeywords {
		metaKeywords[keywordIndex] = make([]byte, restApi.MinMetaKeywordLen)
		rand.Read(metaKeywords[keywordIndex])
	}
	metaKeywordsAsStringsSlice := make([]string, len(metaKeywords))
	for keywordIndex := range metaKeywordsAsStringsSlice {
		metaKeywordsAsStringsSlice[keywordIndex] = base64.URLEncoding.EncodeToString(metaKeywords[keywordIndex])
	}

	testPost.Title = base64.URLEncoding.EncodeToString(title)
	testPost.Author = base64.URLEncoding.EncodeToString(author)
	testPost.Snippet = base64.URLEncoding.EncodeToString(snippet)
	testPost.Content = base64.URLEncoding.EncodeToString(content)
	testPost.Metadata.Keywords = metaKeywordsAsStringsSlice
	testPost.Metadata.Description = base64.URLEncoding.EncodeToString(metaDesc)

	return testPost
}

// API for encoding and decoding messages

func decodePostResponse(responseBody io.ReadCloser, r *ResponsePost) {
	err := json.NewDecoder(responseBody).Decode(r)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
}

func decodePostWithCommentsResponse(responseBody io.ReadCloser, r *ResponsePostWithComments) {
	err := json.NewDecoder(responseBody).Decode(r)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
}

func decodeRangePostsResponse(responseBody io.ReadCloser, r *ResponseRangePosts) {
	err := json.NewDecoder(responseBody).Decode(r)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
}

// -----------
// API for sending posts handling http requests

func getPost(postID string) *http.Response {
	return sendPostHandleMessage("GET", "http://"+address+"/api/posts/"+postID, "")
}

func getPosts(page string, postsPerPage string) *http.Response {
	if len(postsPerPage) != 0 {
		return sendPostHandleMessage(
			"GET", "http://"+address+"/api/posts?page="+page+"&posts-per-page="+postsPerPage, "")
	} else {
		return sendPostHandleMessage("GET", "http://"+address+"/api/posts?page="+page, "")
	}
}

func createPost(message interface{}) *http.Response {
	return sendPostHandleMessage("POST", "http://"+address+"/api/posts", message)
}

func updatePost(postID string, message interface{}) *http.Response {
	return sendPostHandleMessage("PUT", "http://"+address+"/api/posts/"+postID, message)
}

func deletePost(postID string) *http.Response {
	return sendPostHandleMessage("DELETE", "http://"+address+"/api/posts/"+postID, "")
}

func sendPostHandleMessage(method, address string, message interface{}) *http.Response {
	var response *http.Response

	switch method {
	case "GET":
		request, err := http.NewRequest("GET", address, strings.NewReader(""))
		if err != nil {
			panic(fmt.Sprintf("Can not create request. Error: %s", err))
		}

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send request. Error: %s", err))
		}
	case "POST":
		encodedMessage := encodeMessage(message)

		request, err := http.NewRequest("POST", address, bytes.NewReader(encodedMessage))
		if err != nil {
			panic(fmt.Sprintf("Can not create request. Error: %s", err))
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		request.AddCookie(fgpCookie)

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send request. Error: %s", err))
		}
	case "PUT":
		encodedMessage := encodeMessage(message)

		request, err := http.NewRequest("PUT", address, bytes.NewReader(encodedMessage))
		if err != nil {
			panic(fmt.Sprintf("Can not create request. Error: %s", err))
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		request.AddCookie(fgpCookie)

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send request. Error: %s", err))
		}
	case "DELETE":
		request, err := http.NewRequest("DELETE", address, strings.NewReader(""))
		if err != nil {
			panic(fmt.Sprintf("Can not create request. Error: %s", err))
		}
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		request.AddCookie(fgpCookie)

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send request. Error: %s", err))
		}
	}

	return response
}

// -----------
// comments related API

type ResponseComment struct {
	Error restApi.RequestErrorCode
	Body  models.Comment
}

func testCreateCommentFactory() models.CreateCommentRequest {
	post := testPostFactory()
	post.Title = "post for testing comments"
	post.Content = "post for testing comments"

	r := createPost(post)
	checkNiceResponse(r, http.StatusCreated)

	var responseCreatePost ResponsePost
	decodePostResponse(r.Body, &responseCreatePost)
	createdPost := responseCreatePost.Body

	var testComment models.CreateCommentRequest

	testComment.PostID = createdPost.ID

	return testComment
}

func testUpdateCommentFactory() models.UpdateCommentRequest {
	var testComment models.UpdateCommentRequest

	return testComment
}

func getCommentFromResponseBody(r *http.Response) models.Comment {
	var response ResponseComment
	decodeCommentResponse(r.Body, &response)
	comment := response.Body

	return comment
}

func setCommentRequestParentID(comment models.CreateCommentRequest, parentID string) models.CreateCommentRequest {
	comment.ParentCommentID.Valid = true
	comment.ParentCommentID.String = parentID

	return comment
}

// API for encoding and decoding messages

func decodeCommentResponse(responseBody io.ReadCloser, response *ResponseComment) {
	err := json.NewDecoder(responseBody).Decode(&response)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
}

// -----------
// API for sending comments handling http requests

func createComment(message interface{}) *http.Response {
	return sendCommentHandleMessage("POST", "http://"+address+"/api/comments", message)
}

func updateComment(id string, message interface{}) *http.Response {
	return sendCommentHandleMessage("PUT", "http://"+address+"/api/comments/"+id, message)
}

func deleteComment(id string) *http.Response {
	return sendCommentHandleMessage("DELETE", "http://"+address+"/api/comments/"+id, "")
}

func sendCommentHandleMessage(method, address string, message interface{}) *http.Response {
	var response *http.Response

	switch method {
	case "POST":
		encodedMessage := encodeMessage(message)

		request, err := http.NewRequest("POST", address, bytes.NewReader(encodedMessage))
		if err != nil {
			panic(fmt.Sprintf("Can not create POST comment request. Error: %s", err))
		}
		request.Header.Set("Content-Type", "application/json")

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send POST comment request. Error: %s", err))
		}
	case "PUT":
		encodedMessage := encodeMessage(message)

		request, err := http.NewRequest("PUT", address, bytes.NewReader(encodedMessage))
		if err != nil {
			panic(fmt.Sprintf("Can not create UPDATE comment request. Error: %s", err))
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		request.AddCookie(fgpCookie)

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send UPDATE comment request. Error: %s", err))
		}
	case "DELETE":
		request, err := http.NewRequest("DELETE", address, strings.NewReader(""))
		if err != nil {
			panic(fmt.Sprintf("Can not create DELETE comment request. Error: %s", err))
		}
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		request.AddCookie(fgpCookie)

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send DELETE comment request. Error: %s", err))
		}
	}

	return response
}

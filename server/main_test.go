package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/blinky-z/Blog/server/handler"
	"github.com/blinky-z/Blog/server/models"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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

type Response struct {
	Error handler.PostErrorCode
	Body  models.Post
}

type responseAllPosts struct {
	Error handler.PostErrorCode
	Body  []models.Post
}

type authResponse struct {
	Error handler.PostErrorCode
	Body  string
}

type userCredentials struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// helpful API for testing

// API for matching status code and error message of responses

// checkErrorResponse - check response that should return error
func checkErrorResponse(r *http.Response, expectedStatusCode int, expectedErrorMessage handler.PostErrorCode) {
	var response authResponse
	decodeAuthResponse(r.Body, &response)
	if r.StatusCode != expectedStatusCode || response.Error != expectedErrorMessage {
		panic(fmt.Sprintf("Test Error. Received Error code: %d. Error message: %s\n"+
			"Expected Error code: %d. Error message: %s",
			r.StatusCode, response.Error, expectedStatusCode, expectedErrorMessage))
	}
}

// checkErrorResponse - check response that should not return error
func checkNiceResponse(r *http.Response, expectedStatusCode int) {
	if r.StatusCode != expectedStatusCode {
		var response authResponse
		decodeAuthResponse(r.Body, &response)

		panic(fmt.Sprintf("Test Error. Received Error code: %d\nExpected Error code: %d",
			r.StatusCode, expectedStatusCode))
	}
}

// -----------

// API for encoding and decoding messages
func encodeMessage(message interface{}) []byte {
	encodedMessage, err := json.Marshal(message)
	if err != nil {
		panic(fmt.Sprintf("Error encoding message. Error: %s", err))
	}

	return encodedMessage
}

func decodeResponse(responseBody io.ReadCloser, resp *Response) {
	err := json.NewDecoder(responseBody).Decode(resp)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
}

func decodeResponseAllPosts(responseBody io.ReadCloser, resp *responseAllPosts) {
	err := json.NewDecoder(responseBody).Decode(resp)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
}

func decodeAuthResponse(responseBody io.ReadCloser, response *authResponse) {
	err := json.NewDecoder(responseBody).Decode(&response)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
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

// Authorization system tests

// Helpful API for testing authorization tests
func sendAuthUserMessage(address string, credentials userCredentials) *http.Response {
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
	registrationCredentials := userCredentials{Login: login, Email: email, Password: password}

	return sendAuthUserMessage("http://"+Address+"/user/register", registrationCredentials)
}

func loginUser(login, email, password string) *http.Response {
	loginCredentials := userCredentials{Login: login, Email: email, Password: password}

	return sendAuthUserMessage("http://"+Address+"/user/login", loginCredentials)
}

// tests

// Registration tests

func TestRegisterUser(t *testing.T) {
	loginUsername = uuid.New().String()
	loginEmail = loginUsername + "@gmail.com"
	loginPassword = uuid.New().String() + "Z"

	r := registerUser(loginUsername, loginEmail, loginPassword)

	if r.StatusCode != http.StatusOK {
		var response authResponse
		decodeAuthResponse(r.Body, &response)

		panic(fmt.Sprintf("Registration Test Error. Error code: %d. Error message: %s", r.StatusCode, response.Error))
	}
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

func TestLoginUserWithEmail(t *testing.T) {
	r := loginUser("", loginEmail, loginPassword)

	bodyBytes, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	checkNiceResponse(r, http.StatusAccepted)

	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	var response authResponse
	decodeAuthResponse(r.Body, &response)

	authToken = response.Body
}

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

// -----------
// Posts handling tests

// Helpful API for testing posts handling tests

func getPost(postID string) *http.Response {
	return sendPostHandleMessage("GET", "http://"+Address+"/posts/"+postID, "")
}

func getPosts(page string, postsPerPage string) *http.Response {
	if len(postsPerPage) != 0 {
		return sendPostHandleMessage(
			"GET", "http://"+Address+"/posts?page="+page+"&posts-per-page="+postsPerPage, "")
	} else {
		return sendPostHandleMessage("GET", "http://"+Address+"/posts?page="+page, "")
	}
}

func createPost(message interface{}) *http.Response {
	return sendPostHandleMessage("POST", "http://"+Address+"/posts", message)
}

func updatePost(postID string, message interface{}) *http.Response {
	return sendPostHandleMessage("PUT", "http://"+Address+"/posts/"+postID, message)
}

func deletePost(postID string) *http.Response {
	return sendPostHandleMessage("DELETE", "http://"+Address+"/posts/"+postID, "")
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

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send request. Error: %s", err))
		}
	}

	return response
}

// tests

func TestHandlePostIntegrationTest(t *testing.T) {
	var workingPost models.Post

	// Step 1: Create Post
	{
		var response Response

		var sourcePost models.Post
		sourcePost.Title = "Title1"
		sourcePost.Content = "Content1 Content2 Content3"

		r := createPost(sourcePost)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodeResponse(r.Body, &response)
		if r.StatusCode != http.StatusCreated {
			t.Errorf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}

		workingPost = response.Body

		if workingPost.Title != sourcePost.Title {
			t.Errorf("Created post title does not match source post one\nCreated post: %v\n Source post: %v",
				workingPost.Title, sourcePost.Title)
		}

		if workingPost.Content != sourcePost.Content {
			t.Errorf("Created post content does not match source post one\nCreated post: %v\n Source post: %v",
				workingPost.Content, sourcePost.Content)
		}
	}

	// Step 2: Get created post and compare it with returned in prev step one
	{
		var response Response

		r := getPost(workingPost.ID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodeResponse(r.Body, &response)
		if r.StatusCode != http.StatusOK {
			t.Errorf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}

		receivedPost := response.Body

		if receivedPost != workingPost {
			t.Errorf("Received post does not match proper post\nReceived post: %v\n Proper post: %v",
				receivedPost, workingPost)
		}
	}

	// Step 3: Update created post
	{
		var response Response

		newPost := workingPost
		newPost.Title = "newTitle"
		newPost.Content = "NewContent"

		r := updatePost(workingPost.ID, newPost)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodeResponse(r.Body, &response)
		if r.StatusCode != http.StatusOK {
			t.Errorf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}

		updatedPost := response.Body
		if updatedPost != newPost {
			t.Errorf("Received post does not match proper post\nReceived post: %v\n Proper post: %v",
				updatedPost, newPost)
		}

		workingPost = updatedPost
	}

	// Step 4: Get Updated post
	{
		var response Response

		r := getPost(workingPost.ID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodeResponse(r.Body, &response)
		if r.StatusCode != http.StatusOK {
			t.Errorf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}

		receivedPost := response.Body

		if receivedPost != workingPost {
			t.Errorf("Received post does not match updated post\nReceived post: %v\n Updated post: %v",
				receivedPost, workingPost)
		}
	}

	// Step 5: Delete updated post
	{
		r := deletePost(workingPost.ID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		if r.StatusCode != http.StatusOK {
			var response Response

			decodeResponse(r.Body, &response)

			t.Errorf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}
	}

	// Step 6: Get deleted post
	{
		var response Response

		r := getPost(workingPost.ID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodeResponse(r.Body, &response)
		if r.StatusCode != http.StatusNotFound {
			t.Errorf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}
	}
}

func TestCreatePostWithInvalidRequestBody(t *testing.T) {
	var response Response

	message := `{"bad request body"}`

	resp := createPost(message)
	decodeResponse(resp.Body, &response)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.BadRequestBody {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestCreatePostWithNullTitle(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   "",
		"content": "Content1 Content2 Content3",
	}

	resp := createPost(message)
	decodeResponse(resp.Body, &response)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidTitle {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestCreatePostWithTooLongTitle(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   strings.Repeat("a", handler.MaxPostTitleLen*2),
		"content": "Content1 Content2 Content3",
	}

	resp := createPost(message)
	decodeResponse(resp.Body, &response)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidTitle {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestCreatePostWithNullContent(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   "Title1",
		"content": "",
	}

	resp := createPost(message)
	decodeResponse(resp.Body, &response)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidContent {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestGetPostWithInvalidID(t *testing.T) {
	var response Response

	resp := getPost("post1")
	decodeResponse(resp.Body, &response)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidID {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestGetNonexistentPost(t *testing.T) {
	var response Response

	resp := getPost("-1")
	decodeResponse(resp.Body, &response)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestUpdatePostWithInvalidRequestBody(t *testing.T) {
	var response Response

	message := `{"bad request body":"asd"}`

	resp := updatePost("1", message)
	decodeResponse(resp.Body, &response)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.BadRequestBody {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestUpdatePostWithNullTitle(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   "",
		"content": "Content1 Content2 Content3",
	}

	resp := updatePost("1", message)
	decodeResponse(resp.Body, &response)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidTitle {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestUpdatePostWithTooLongTitle(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   strings.Repeat("a", handler.MaxPostTitleLen*2),
		"content": "Content1 Content2 Content3",
	}

	resp := updatePost("1", message)
	decodeResponse(resp.Body, &response)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidTitle {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestUpdatePostWithNullContent(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   "TITLE2",
		"content": "",
	}

	resp := updatePost("1", message)
	decodeResponse(resp.Body, &response)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidContent {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestUpdatePostNonexistentPost(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   "TITLE2",
		"content": "Content1 Content2 Content3",
	}

	resp := updatePost("-1", message)
	decodeResponse(resp.Body, &response)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestUpdatePostWithInvalidID(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   "TITLE2",
		"content": "Content1 Content2 Content3",
	}

	resp := updatePost("post1", message)
	decodeResponse(resp.Body, &response)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidID {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestDeletePostNonexistentPost(t *testing.T) {
	resp := deletePost("-1")
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		var response Response

		decodeResponse(resp.Body, &response)

		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestDeletePostWithInvalidID(t *testing.T) {
	var response Response

	resp := deletePost("post1")
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	decodeResponse(resp.Body, &response)
	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidID {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func comparePosts(receivedPosts, properPosts []models.Post) {
	for i, receivedPost := range receivedPosts {
		properPost := properPosts[len(receivedPosts)-i-1]
		if receivedPost != properPost {
			log.Fatalf("Received post does not match proper post\nReceived post: %v\n Proper post: %v",
				receivedPost, properPost)
		}
	}
}

func TestGetRangeOfPostsWithCustomPostsPerPage(t *testing.T) {
	var workingPosts []models.Post

	testPostsNumber := 20

	for i := 0; i < testPostsNumber; i++ {
		message := map[string]interface{}{
			"title":   "Title" + strconv.Itoa(i),
			"content": "Content" + strconv.Itoa(i),
		}

		var response Response
		resp := sendPostHandleMessage("POST", "http://"+Address+"/posts", message)
		decodeResponse(resp.Body, &response)

		workingPosts = append(workingPosts, response.Body)
	}

	resp := getPosts("0", "20")

	var response responseAllPosts
	decodeResponseAllPosts(resp.Body, &response)

	receivedPosts := response.Body

	comparePosts(receivedPosts, workingPosts)
}

func TestGetRangeOfPostsWithDefaultPostsPerPage(t *testing.T) {
	var workingPosts []models.Post

	testPostsNumber := 10

	for i := 0; i < testPostsNumber; i++ {
		message := map[string]interface{}{
			"title":   "Title" + strconv.Itoa(i),
			"content": "Content" + strconv.Itoa(i),
		}

		var response Response
		resp := sendPostHandleMessage("POST", "http://"+Address+"/posts", message)
		decodeResponse(resp.Body, &response)

		workingPosts = append(workingPosts, response.Body)
	}

	resp := getPosts("0", "")

	var response responseAllPosts
	decodeResponseAllPosts(resp.Body, &response)

	receivedPosts := response.Body

	comparePosts(receivedPosts, workingPosts)
}

func TestGetRangeOfPostsWithNegativePage(t *testing.T) {
	resp := getPosts("-1", "")

	var response responseAllPosts
	decodeResponseAllPosts(resp.Body, &response)

	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidRange {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestGetRangeOfPostsWithNonNumberPage(t *testing.T) {
	resp := getPosts("adasdf", "")

	var response responseAllPosts
	decodeResponseAllPosts(resp.Body, &response)

	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidRange {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestGetRangeOfPostsWithTooLongPostsPerPage(t *testing.T) {
	resp := getPosts("0", strconv.Itoa(handler.MaxPostsPerPage*2))

	var response responseAllPosts
	decodeResponseAllPosts(resp.Body, &response)

	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidRange {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestGetRangeOfPostsWithNonNumberPostsPerPage(t *testing.T) {
	resp := getPosts("0", "asddfa")

	var response responseAllPosts
	decodeResponseAllPosts(resp.Body, &response)

	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidRange {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

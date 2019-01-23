package main

// Posts handling tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/blinky-z/Blog/server/handler"
	"github.com/blinky-z/Blog/server/models"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

type ResponseSinglePost struct {
	Error handler.PostErrorCode
	Body  models.Post
}

type ResponseRangePosts struct {
	Error handler.PostErrorCode
	Body  []models.Post
}

// -----------

// API for encoding and decoding messages

func decodeSinglePostResponse(responseBody io.ReadCloser, resp *ResponseSinglePost) {
	err := json.NewDecoder(responseBody).Decode(resp)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
}

func decodeRangePostsResponse(responseBody io.ReadCloser, resp *ResponseRangePosts) {
	err := json.NewDecoder(responseBody).Decode(resp)
	if err != nil {
		panic(fmt.Sprintf("Error decoding received body. Error: %s", err))
	}
}

// -----------
// Helpful API for sending posts handling http requests

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

// -----------
// tests

func TestHandlePostIntegrationTest(t *testing.T) {
	var workingPost models.Post

	// Step 1: Create Post
	{
		var response ResponseSinglePost

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
		decodeSinglePostResponse(r.Body, &response)
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
		var response ResponseSinglePost

		r := getPost(workingPost.ID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodeSinglePostResponse(r.Body, &response)
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
		var response ResponseSinglePost

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
		decodeSinglePostResponse(r.Body, &response)
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
		var response ResponseSinglePost

		r := getPost(workingPost.ID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodeSinglePostResponse(r.Body, &response)
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
			var response ResponseSinglePost

			decodeSinglePostResponse(r.Body, &response)

			t.Errorf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}
	}

	// Step 6: Get deleted post
	{
		var response ResponseSinglePost

		r := getPost(workingPost.ID)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodeSinglePostResponse(r.Body, &response)
		if r.StatusCode != http.StatusNotFound {
			t.Errorf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}
	}
}

func TestCreatePostWithInvalidRequestBody(t *testing.T) {
	message := `{"bad request body"}`

	resp := createPost(message)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.BadRequestBody)
}

func TestCreatePostWithNullTitle(t *testing.T) {
	message := map[string]interface{}{
		"title":   "",
		"content": "Content1 Content2 Content3",
	}

	resp := createPost(message)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.InvalidTitle)
}

func TestCreatePostWithTooLongTitle(t *testing.T) {
	message := map[string]interface{}{
		"title":   strings.Repeat("a", handler.MaxPostTitleLen*2),
		"content": "Content1 Content2 Content3",
	}

	resp := createPost(message)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.InvalidTitle)
}

func TestCreatePostWithNullContent(t *testing.T) {
	message := map[string]interface{}{
		"title":   "Title1",
		"content": "",
	}

	resp := createPost(message)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.InvalidContent)
}

func TestGetPostWithInvalidID(t *testing.T) {
	resp := getPost("post1")
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.InvalidID)
}

func TestGetNonexistentPost(t *testing.T) {
	resp := getPost("-1")
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusNotFound, handler.NoSuchPost)
}

func TestUpdatePostWithInvalidRequestBody(t *testing.T) {
	message := `{"bad request body":"asd"}`

	resp := updatePost("1", message)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.BadRequestBody)
}

func TestUpdatePostWithNullTitle(t *testing.T) {
	message := map[string]interface{}{
		"title":   "",
		"content": "Content1 Content2 Content3",
	}

	resp := updatePost("1", message)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.InvalidTitle)
}

func TestUpdatePostWithTooLongTitle(t *testing.T) {
	message := map[string]interface{}{
		"title":   strings.Repeat("a", handler.MaxPostTitleLen*2),
		"content": "Content1 Content2 Content3",
	}

	resp := updatePost("1", message)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.InvalidTitle)
}

func TestUpdatePostWithNullContent(t *testing.T) {
	message := map[string]interface{}{
		"title":   "TITLE2",
		"content": "",
	}

	resp := updatePost("1", message)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.InvalidContent)
}

func TestUpdatePostNonexistentPost(t *testing.T) {
	message := map[string]interface{}{
		"title":   "TITLE2",
		"content": "Content1 Content2 Content3",
	}

	resp := updatePost("-1", message)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusNotFound, handler.NoSuchPost)
}

func TestUpdatePostWithInvalidID(t *testing.T) {
	message := map[string]interface{}{
		"title":   "TITLE2",
		"content": "Content1 Content2 Content3",
	}

	resp := updatePost("post1", message)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.InvalidID)
}

func TestDeletePostNonexistentPost(t *testing.T) {
	resp := deletePost("-1")
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkNiceResponse(resp, http.StatusOK)
}

func TestDeletePostWithInvalidID(t *testing.T) {
	resp := deletePost("post1")
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.InvalidID)
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

		var response ResponseSinglePost
		resp := sendPostHandleMessage("POST", "http://"+Address+"/posts", message)
		decodeSinglePostResponse(resp.Body, &response)

		workingPosts = append(workingPosts, response.Body)
	}

	resp := getPosts("0", "20")

	var response ResponseRangePosts
	decodeRangePostsResponse(resp.Body, &response)

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

		var response ResponseSinglePost
		resp := sendPostHandleMessage("POST", "http://"+Address+"/posts", message)
		decodeSinglePostResponse(resp.Body, &response)

		workingPosts = append(workingPosts, response.Body)
	}

	resp := getPosts("0", "")

	var response ResponseRangePosts
	decodeRangePostsResponse(resp.Body, &response)

	receivedPosts := response.Body

	comparePosts(receivedPosts, workingPosts)
}

func TestGetRangeOfPostsWithNegativePage(t *testing.T) {
	resp := getPosts("-1", "")
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.InvalidRange)
}

func TestGetRangeOfPostsWithNonNumberPage(t *testing.T) {
	resp := getPosts("adasdf", "")
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.InvalidRange)
}

func TestGetRangeOfPostsWithTooLongPostsPerPage(t *testing.T) {
	resp := getPosts("0", strconv.Itoa(handler.MaxPostsPerPage*2))
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.InvalidRange)
}

func TestGetRangeOfPostsWithNonNumberPostsPerPage(t *testing.T) {
	resp := getPosts("0", "asddfa")
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(resp, http.StatusBadRequest, handler.InvalidRange)
}

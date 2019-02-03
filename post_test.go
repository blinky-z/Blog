package main

// Posts handling tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/blinky-z/Blog/handler/api"
	"github.com/blinky-z/Blog/models"
	"gotest.tools/assert"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

type ResponseSinglePost struct {
	Error api.PostErrorCode
	Body  models.Post
}

type ResponseRangePosts struct {
	Error api.PostErrorCode
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
	testPost.Metadata.Keywords = []string{"testMetadata1", "testMetadata2"}
	testPost.Metadata.Description = "test meta description"

	return testPost
}

// API for encoding and decoding messages

func decodeSinglePostResponse(responseBody io.ReadCloser, r *ResponseSinglePost) {
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
// Helpful API for sending posts handling http requests

func getPost(postID string) *http.Response {
	return sendPostHandleMessage("GET", "http://"+Address+"/api/posts/"+postID, "")
}

func getPosts(page string, postsPerPage string) *http.Response {
	if len(postsPerPage) != 0 {
		return sendPostHandleMessage(
			"GET", "http://"+Address+"/api/posts?page="+page+"&posts-per-page="+postsPerPage, "")
	} else {
		return sendPostHandleMessage("GET", "http://"+Address+"/api/posts?page="+page, "")
	}
}

func createPost(message interface{}) *http.Response {
	return sendPostHandleMessage("POST", "http://"+Address+"/api/posts", message)
}

func updatePost(postID string, message interface{}) *http.Response {
	return sendPostHandleMessage("PUT", "http://"+Address+"/api/posts/"+postID, message)
}

func deletePost(postID string) *http.Response {
	return sendPostHandleMessage("DELETE", "http://"+Address+"/api/posts/"+postID, "")
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
		request.AddCookie(ctxCookie)

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
		request.AddCookie(ctxCookie)

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
		request.AddCookie(ctxCookie)

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
	workingPost := testPostFactory()

	// Step 1: Create Post
	{
		var response ResponseSinglePost

		sourcePost := testPostFactory()
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
			t.Fatalf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}

		workingPost = response.Body

		if workingPost.Title != sourcePost.Title {
			t.Fatalf("Created post title does not match source post one\nCreated post: %v\n Source post: %v",
				workingPost.Title, sourcePost.Title)
		}

		if workingPost.Content != sourcePost.Content {
			t.Fatalf("Created post content does not match source post one\nCreated post: %v\n Source post: %v",
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
			t.Fatalf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}

		receivedPost := response.Body

		if !comparePosts(receivedPost, workingPost) {
			t.Fatalf("Received post does not match proper post\nReceived post: %v\n Proper post: %v",
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
			t.Fatalf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}

		updatedPost := response.Body
		if !comparePosts(updatedPost, newPost) {
			t.Fatalf("Received post does not match proper post\nReceived post: %v\n Proper post: %v",
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
			t.Fatalf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}

		receivedPost := response.Body

		if !comparePosts(receivedPost, workingPost) {
			t.Fatalf("Received post does not match updated post\nReceived post: %v\n Updated post: %v",
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

			t.Fatalf("Error %d. Error message: %s", r.StatusCode, response.Error)
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
			t.Fatalf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}
	}
}

func TestCreatePostWithInvalidRequestBody(t *testing.T) {
	message := `{"bad request body"}`

	r := createPost(message)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.BadRequestBody)
}

func TestCreatePostWithNullTitle(t *testing.T) {
	message := map[string]interface{}{
		"title":   "",
		"content": "Content1 Content2 Content3",
	}

	r := createPost(message)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidTitle)
}

func TestCreatePostWithTooLongTitle(t *testing.T) {
	message := map[string]interface{}{
		"title":   strings.Repeat("a", api.MaxPostTitleLen*2),
		"content": "Content1 Content2 Content3",
	}

	r := createPost(message)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidTitle)
}

func TestCreatePostWithNullContent(t *testing.T) {
	post := testPostFactory()
	post.Title = "Title1"
	post.Content = ""

	r := createPost(post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidContent)
}

func TestGetPostWithInvalidID(t *testing.T) {
	r := getPost("post1")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidID)
}

func TestGetNonexistentPost(t *testing.T) {
	deletePost("0")

	r := getPost("0")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusNotFound, api.NoSuchPost)
}

func TestUpdatePostWithInvalidRequestBody(t *testing.T) {
	message := `{"bad request body":"asd"}`

	r := updatePost("1", message)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.BadRequestBody)
}

func TestUpdatePostWithNullTitle(t *testing.T) {
	post := testPostFactory()
	post.Title = ""
	post.Content = "Content1 Content2 Content3"

	r := updatePost("1", post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidTitle)
}

func TestUpdatePostWithTooLongTitle(t *testing.T) {
	post := testPostFactory()
	post.Title = strings.Repeat("a", api.MaxPostTitleLen*2)
	post.Content = "Content1 Content2 Content3"

	r := updatePost("1", post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidTitle)
}

func TestUpdatePostWithNullContent(t *testing.T) {
	post := testPostFactory()
	post.Title = "TITLE2"
	post.Content = ""

	r := updatePost("1", post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidContent)
}

func TestUpdatePostNonexistentPost(t *testing.T) {
	deletePost("0")

	post := testPostFactory()
	post.Title = "TITLE2"
	post.Content = "Content1 Content2 Content3"

	r := updatePost("0", post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusNotFound, api.NoSuchPost)
}

func TestUpdatePostWithInvalidID(t *testing.T) {
	post := testPostFactory()
	post.Title = "TITLE2"
	post.Content = "Content1 Content2 Content3"

	r := updatePost("post1", post)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidID)
}

func TestDeletePostNonexistentPost(t *testing.T) {
	postToDeleteId := "1"

	deletePost(postToDeleteId)

	r := deletePost(postToDeleteId)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkNiceResponse(r, http.StatusOK)
}

func TestDeletePostWithInvalidID(t *testing.T) {
	r := deletePost("post1")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidID)
}

func TestGetRangeOfPostsWithCustomPostsPerPage(t *testing.T) {
	var workingPosts []models.Post

	testPostsNumber := 20

	for i := 0; i < testPostsNumber; i++ {
		currentPost := testPostFactory()
		currentPost.Title = "Title" + strconv.Itoa(i)
		currentPost.Content = "Content" + strconv.Itoa(i)

		var response ResponseSinglePost
		r := sendPostHandleMessage("POST", "http://"+Address+"/api/posts", currentPost)
		decodeSinglePostResponse(r.Body, &response)

		workingPosts = append([]models.Post{response.Body}, workingPosts...)
	}

	r := getPosts("0", "20")

	var response ResponseRangePosts
	decodeRangePostsResponse(r.Body, &response)

	receivedPosts := response.Body

	if !comparePostLists(receivedPosts, workingPosts) {
		log.Fatalf("Received post list does not match proper post list\nReceived post list: %v\n Proper post list: %v",
			receivedPosts, workingPosts)
	}
}

func TestGetRangeOfPostsWithDefaultPostsPerPage(t *testing.T) {
	var workingPosts []models.Post

	testPostsNumber := 10

	for i := 0; i < testPostsNumber; i++ {
		currentPost := testPostFactory()
		currentPost.Title = "Title" + strconv.Itoa(i)
		currentPost.Content = "Content" + strconv.Itoa(i)

		var response ResponseSinglePost
		r := sendPostHandleMessage("POST", "http://"+Address+"/api/posts", currentPost)
		decodeSinglePostResponse(r.Body, &response)

		workingPosts = append([]models.Post{response.Body}, workingPosts...)
	}

	r := getPosts("0", "")

	var response ResponseRangePosts
	decodeRangePostsResponse(r.Body, &response)

	receivedPosts := response.Body

	if !comparePostLists(receivedPosts, workingPosts) {
		log.Fatalf("Received post list does not match proper post list\nReceived post list: %v\n Proper post list: %v",
			receivedPosts, workingPosts)
	}
}

func TestGetRangeOfPostsWithNegativePage(t *testing.T) {
	r := getPosts("-1", "")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidRange)
}

func TestGetRangeOfPostsWithNonNumberPage(t *testing.T) {
	r := getPosts("adasdf", "")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidRange)
}

func TestGetRangeOfPostsWithTooLongPostsPerPage(t *testing.T) {
	r := getPosts("0", strconv.Itoa(api.MaxPostsPerPage*2))
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidRange)
}

func TestGetRangeOfPostsWithNonNumberPostsPerPage(t *testing.T) {
	r := getPosts("0", "asddfa")
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	checkErrorResponse(r, http.StatusBadRequest, api.InvalidRange)
}

func TestGetRangeOfPostsGetEmptyPage(t *testing.T) {
	r := getPosts("10000000", "")

	var response ResponseRangePosts
	decodeRangePostsResponse(r.Body, &response)

	receivedPosts := response.Body

	assert.Assert(t, len(receivedPosts) == 0)
}

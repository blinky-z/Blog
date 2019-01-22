package main

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
	"time"
)

var (
	client = &http.Client{}
)

type Response struct {
	Error handler.PostErrorCode
	Body  models.Post
}

type responseAllPosts struct {
	Error handler.PostErrorCode
	Body  []models.Post
}

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

func getPost(postID string) *http.Response {
	return sendMessage("GET", "http://"+Address+"/posts/"+postID, "")
}

func getPosts(page string, postsPerPage string) *http.Response {
	if len(postsPerPage) != 0 {
		return sendMessage("GET", "http://"+Address+"/posts?page="+page+"&posts-per-page="+postsPerPage, "")
	} else {
		return sendMessage("GET", "http://"+Address+"/posts?page="+page, "")
	}
}

func createPost(message interface{}) *http.Response {
	return sendMessage("POST", "http://"+Address+"/posts", message)
}

func updatePost(postID string, message interface{}) *http.Response {
	return sendMessage("PUT", "http://"+Address+"/posts/"+postID, message)
}

func deletePost(postID string) *http.Response {
	return sendMessage("DELETE", "http://"+Address+"/posts/"+postID, "")
}

func sendMessage(method, address string, message interface{}) *http.Response {
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
		request.Header.Set("Content-Type", "application/json")
		if err != nil {
			panic(fmt.Sprintf("Can not create request. Error: %s", err))
		}

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send request. Error: %s", err))
		}
	case "PUT":
		encodedMessage := encodeMessage(message)

		request, err := http.NewRequest("PUT", address, bytes.NewReader(encodedMessage))
		request.Header.Set("Content-Type", "application/json")
		if err != nil {
			panic(fmt.Sprintf("Can not create request. Error: %s", err))
		}

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send request. Error: %s", err))
		}
	case "DELETE":
		request, err := http.NewRequest("DELETE", address, strings.NewReader(""))
		if err != nil {
			panic(fmt.Sprintf("Can not create request. Error: %s", err))
		}

		response, err = client.Do(request)
		if err != nil {
			panic(fmt.Sprintf("Can not send request. Error: %s", err))
		}
	}

	return response
}

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

		log.Printf("Created post id: %s", workingPost.ID)
	}

	// Step 2: Get created post and compare it with created one
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
				updatedPost, workingPost)
		}

		workingPost = updatedPost
	}

	// Step 4: Delete updated post
	{
		var response Response

		r := deletePost(workingPost.ID)
		decodeResponse(r.Body, &response)
		if r.StatusCode != http.StatusOK {
			t.Errorf("Error %d. Error message: %s", r.StatusCode, response.Error)
		}

		r = getPost(workingPost.ID)
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
		"title":   strings.Repeat("a", 130),
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
		"title":   strings.Repeat("a", 130),
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
	var response Response

	resp := deletePost("-1")
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	decodeResponse(resp.Body, &response)
	if resp.StatusCode != http.StatusOK {
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
		resp := sendMessage("POST", "http://"+Address+"/posts", message)
		decodeResponse(resp.Body, &response)

		workingPosts = append(workingPosts, response.Body)
	}

	resp := sendMessage("GET", "http://"+Address+"/posts?page=0&posts-per-page=20", "")

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
		resp := sendMessage("POST", "http://"+Address+"/posts", message)
		decodeResponse(resp.Body, &response)

		workingPosts = append(workingPosts, response.Body)
	}

	resp := sendMessage("GET", "http://"+Address+"/posts?page=0", "")

	var response responseAllPosts
	decodeResponseAllPosts(resp.Body, &response)

	receivedPosts := response.Body

	comparePosts(receivedPosts, workingPosts)
}

func TestGetRangeOfPostsWithNegativePage(t *testing.T) {
	resp := sendMessage("GET", "http://"+Address+"/posts?page=-1", "")

	var response responseAllPosts
	decodeResponseAllPosts(resp.Body, &response)

	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidRange {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestGetRangeOfPostsWithNonNumberPage(t *testing.T) {
	resp := sendMessage("GET", "http://"+Address+"/posts?page=asdsa", "")

	var response responseAllPosts
	decodeResponseAllPosts(resp.Body, &response)

	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidRange {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestGetRangeOfPostsWithTooLongPostsPerPage(t *testing.T) {
	resp := sendMessage("GET", "http://"+Address+"/posts?page=0&posts-per-page=100", "")

	var response responseAllPosts
	decodeResponseAllPosts(resp.Body, &response)

	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidRange {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestGetRangeOfPostsWithNonNumberPostsPerPage(t *testing.T) {
	resp := sendMessage("GET", "http://"+Address+"/posts?page=0&posts-per-page=advsc", "")

	var response responseAllPosts
	decodeResponseAllPosts(resp.Body, &response)

	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidRange {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

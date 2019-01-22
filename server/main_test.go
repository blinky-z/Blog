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

func encodeMessage(message interface{}) []byte {
	encodedMessage, err := json.Marshal(message)
	if err != nil {
		panic(fmt.Sprintf("Error encoding message. Error: %s", err))
	}

	return encodedMessage
}

func decodeMessage(responseBody io.ReadCloser, resp *Response) {
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

		message := map[string]interface{}{
			"title":   "Title1",
			"content": "Content1 Content2 Content3",
		}

		resp := createPost(message)
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodeMessage(resp.Body, &response)
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
		}

		workingPost = response.Body

		log.Printf("Created post id: %s", workingPost.ID)
	}

	// Step 2: Get created post
	{
		var response Response

		resp := getPost(workingPost.ID)
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodeMessage(resp.Body, &response)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
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

		resp := updatePost(workingPost.ID, newPost)
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodeMessage(resp.Body, &response)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
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

		resp := deletePost(workingPost.ID)
		decodeMessage(resp.Body, &response)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
		}

		resp = getPost(workingPost.ID)
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		decodeMessage(resp.Body, &response)
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
		}
	}
}

func TestCreatePostWithInvalidRequestBody(t *testing.T) {
	var response Response

	message := `{"bad request body"}`

	resp := createPost(message)
	decodeMessage(resp.Body, &response)
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
	decodeMessage(resp.Body, &response)
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

func TestCreatePostWithLongTitle(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   strings.Repeat("a", 130),
		"content": "Content1 Content2 Content3",
	}

	resp := createPost(message)
	decodeMessage(resp.Body, &response)
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
	decodeMessage(resp.Body, &response)
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
	decodeMessage(resp.Body, &response)
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
	decodeMessage(resp.Body, &response)
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
	decodeMessage(resp.Body, &response)
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
	decodeMessage(resp.Body, &response)
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

func TestUpdatePostWithLongTitle(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   strings.Repeat("a", 130),
		"content": "Content1 Content2 Content3",
	}

	resp := updatePost("1", message)
	decodeMessage(resp.Body, &response)
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
	decodeMessage(resp.Body, &response)
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
	decodeMessage(resp.Body, &response)
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
	decodeMessage(resp.Body, &response)
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
	decodeMessage(resp.Body, &response)
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
	decodeMessage(resp.Body, &response)
	if resp.StatusCode != http.StatusBadRequest || response.Error != handler.InvalidID {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

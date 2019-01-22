package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/blinky-z/Blog/server/models"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	workingPost Post
	client      = &http.Client{}
)

type Response struct {
	Error string
	Body  Post
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

func TestCreatePost(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   "Title1",
		"content": "Content1 Content2 Content3",
	}

	resp := sendMessage("POST", "http://"+Address+"/posts", message)
	decodeMessage(resp.Body, &response)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}

	workingPost = response.Body

	log.Printf("Created post id: %s", workingPost.ID)
}

func TestCreatePostWithInvalidTitle(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   strings.Repeat("a", 130),
		"content": "Content1 Content2 Content3",
	}

	resp := sendMessage("POST", "http://"+Address+"/posts", message)
	decodeMessage(resp.Body, &response)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestGetCertainPost(t *testing.T) {
	var response Response

	resp := sendMessage("GET", "http://"+Address+"/posts/"+workingPost.ID, "")
	decodeMessage(resp.Body, &response)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}

	receivedPost := response.Body

	if receivedPost != workingPost {
		t.Errorf("Received post does not match proper post\nReceived post: %v\n Proper post: %v",
			receivedPost, workingPost)
	}
}

func TestGetNonexistentPost(t *testing.T) {
	var response Response

	resp := sendMessage("GET", "http://"+Address+"/posts/-1", "")
	decodeMessage(resp.Body, &response)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestUpdatePost(t *testing.T) {
	var response Response

	newPost := workingPost
	newPost.Title = "newTitle"
	newPost.Content = "NewContent"

	resp := sendMessage("PUT", "http://"+Address+"/posts/"+workingPost.ID, newPost)
	decodeMessage(resp.Body, &response)
	defer resp.Body.Close()
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

func TestUpdatePostWithInvalidTitle(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   strings.Repeat("a", 130),
		"content": "Content1 Content2 Content3",
	}

	resp := sendMessage("PUT", "http://"+Address+"/posts/"+workingPost.ID, message)
	decodeMessage(resp.Body, &response)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestUpdateNonexistentPost(t *testing.T) {
	var response Response

	message := map[string]interface{}{
		"title":   "TITLE2",
		"content": "Content1 Content2 Content3",
	}

	resp := sendMessage("PUT", "http://"+Address+"/posts/-1", message)
	decodeMessage(resp.Body, &response)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestDeletePost(t *testing.T) {
	var response Response

	resp := sendMessage("DELETE", "http://"+Address+"/posts/"+workingPost.ID, "")
	decodeMessage(resp.Body, &response)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
	resp.Body.Close()

	resp = sendMessage("GET", "http://"+Address+"/posts/"+workingPost.ID, "")
	defer resp.Body.Close()
	decodeMessage(resp.Body, &response)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

func TestDeleteNonexistentPost(t *testing.T) {
	var response Response

	resp := sendMessage("DELETE", "http://"+Address+"/posts/-1", "")
	defer resp.Body.Close()
	decodeMessage(resp.Body, &response)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Error %d. Error message: %s", resp.StatusCode, response.Error)
	}
}

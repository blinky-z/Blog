package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

type Post struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Date    time.Time `json:"date"`
	Content string    `json:"content"`
}

func TestCreatePost(t *testing.T) {
	message := map[string]interface{}{
		"title":   "Title1",
		"content": "Content1 Content2 Content3",
	}

	encodedMessage, err := json.Marshal(message)
	if err != nil {
		log.Fatalf("Can not encode message. Error: %s", err)
	}

	resp, err := http.Post("http://localhost:8080/posts", "application/json", bytes.NewBuffer(encodedMessage))
	defer resp.Body.Close()
	if err != nil {
		log.Fatalf("Can not send request. Error: %s", err)
	}
}

func TestGetCertainPost(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/posts/1")
	if err != nil {
		log.Fatalf("Can not send request. Error: %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading received body. Error: %s", err)
	}

	var receivedPost Post
	err = json.Unmarshal(body, &receivedPost)
	if err != nil {
		log.Fatalf("Error decoding received body. Error: %s", err)
	}

	var properPost Post
	err = json.Unmarshal(body, &encodedMessage)
	if err != nil {
		log.Fatalf("Error decoding proper body. Error: %s", err)
	}

	if receivedPost != properPost {
		log.Fatalf("Received post does not match proper post\nReceived post: %v\n Proper post: %v",
			receivedPost, properPost)
	}
}

func main() {

}

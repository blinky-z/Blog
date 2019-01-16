package main

import (
	. "./models"
	"bytes"
	"encoding/json"
	"github.com/spf13/cast"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestRunServer(t *testing.T) {
	go RunServer("testConfig", ".")
	time.Sleep(3 * time.Second)
	Db.Exec("DELETE from testdb.public.posts")
	Db.Exec("ALTER SEQUENCE posts_id_seq RESTART WITH 1")
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
	if err != nil {
		log.Fatalf("Can not send request. Error: %s", err)
	}
	if resp.StatusCode != http.StatusCreated {
		log.Fatal(resp.Body)
	}
	defer resp.Body.Close()
}

func TestGetCertainPost(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/posts/1")
	if err != nil {
		log.Fatalf("Can not send request. Error: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Body)
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
	properPost.ID = "1"
	properPost.Title = "Title1"
	properPost.Content = "Content1 Content2 Content3"

	if receivedPost.ID != properPost.ID || receivedPost.Title != properPost.Title ||
		receivedPost.Content != properPost.Content {
		log.Fatalf("Received post does not match proper post\nReceived post: %v\n Proper post: %v",
			receivedPost, properPost)
	}
}

func TestGetPosts(t *testing.T) {
	for i := 0; i < 10; i++ {
		message := map[string]interface{}{
			"title":   "Title" + cast.ToString(i),
			"content": "Content" + cast.ToString(i),
		}

		encodedMessage, _ := json.Marshal(message)
		http.Post("http://localhost:8080/posts", "application/json", bytes.NewBuffer(encodedMessage))
	}

	resp, err := http.Get("http://localhost:8080/posts")
	if err != nil {
		log.Fatalf("Can not send request. Error: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Body)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading received body. Error: %s", err)
	}

	var receivedPosts []Post
	err = json.Unmarshal(body, &receivedPosts)
	if err != nil {
		log.Fatalf("Error decoding received body. Error: %s", err)
	}

	for i, receivedPost := range receivedPosts[1:] {
		curPostIndex := cast.ToString(i)
		databaseOffset := cast.ToString(i + 2)
		properPost := Post{ID: databaseOffset, Title: "Title" + curPostIndex, Content: "Content" + curPostIndex}

		if receivedPost.ID != properPost.ID || receivedPost.Title != properPost.Title ||
			receivedPost.Content != properPost.Content {
			log.Fatalf("Received post does not match proper post\nReceived post: %v\n Proper post: %v",
				receivedPost, properPost)
		}
	}
}

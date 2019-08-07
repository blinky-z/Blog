package models

import "time"

// Post - represents blog post
type Post struct {
	ID       string
	Title    string
	Date     time.Time
	Snippet  string
	Content  string
	Metadata MetaData
	Tags     []string
}

//CreatePostRequest - represents post creation request
type CreatePostRequest struct {
	Title    string   `json:"title"`
	Snippet  string   `json:"snippet"`
	Content  string   `json:"content"`
	Metadata MetaData `json:"metadata"`
	Tags     []string `json:"tags"`
}

//UpdatePostRequest - represents post update request
type UpdatePostRequest struct {
	Title    string   `json:"title"`
	Snippet  string   `json:"snippet"`
	Content  string   `json:"content"`
	Metadata MetaData `json:"metadata"`
	Tags     []string `json:"tags"`
}

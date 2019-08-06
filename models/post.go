package models

import "time"

// Post - represents blog post
type Post struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Author   string    `json:"author"`
	Date     time.Time `json:"date"`
	Snippet  string    `json:"snippet"`
	Content  string    `json:"content"`
	Metadata MetaData  `json:"metadata"`
	Tags     []string  `json:"tags"`
}

//CreatePostRequest - represents post creation request
type CreatePostRequest struct {
	Title    string   `json:"title"`
	Author   string   `json:"author"`
	Snippet  string   `json:"snippet"`
	Content  string   `json:"content"`
	Metadata MetaData `json:"metadata"`
	Tags     []string `json:"tags"`
}

//UpdatePostRequest - represents post update request
type UpdatePostRequest struct {
	Title    string   `json:"title"`
	Author   string   `json:"author"`
	Snippet  string   `json:"snippet"`
	Content  string   `json:"content"`
	Metadata MetaData `json:"metadata"`
	Tags     []string `json:"tags"`
}

// CertainPost - this struct is used to return not only post data, but also comments of this post
type CertainPost struct {
	Post
	Comments []*CommentWithChilds `json:"comments"`
}

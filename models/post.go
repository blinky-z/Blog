package models

import "time"

// Post - represents blog post
// 'ID' - post id. Id is filed automatically by database
// 'Title' - title of post
// 'Date' - post creation time
// 'Content' - post content
type Post struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Date     time.Time `json:"date"`
	Content  string    `json:"content"`
	Metadata MetaData  `json:"metadata"`
}

// CertainPostResponse - use this struct in GetCertainPost http handler (REST API)
// This struct extends Post, adding Comments field allowing to store post's comments
type CertainPostResponse struct {
	Post
	Comments []Comment
}

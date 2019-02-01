package models

import "time"

// Post - represents row in table contains posts
// 'ID' - post id. Id is filed automatically by database
// 'Title' - title of post
// 'Date' - post creation time
// 'Content' - post content
type Post struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Date    time.Time `json:"date"`
	Content string    `json:"content"`
}

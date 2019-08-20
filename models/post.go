package models

import "time"

// Post - represents blog post
// @ID - ID created by database
// @Title - title
// @Date - creation time
// @Snippet - short description of this post
// @Content - content
// @Metadata - site metadata for this post. It replaces description and keywords in <head> tag
// @Tags - tags
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
	Title     string   `json:"title"`
	Snippet   string   `json:"snippet"`
	Content   string   `json:"content"`
	ContentMD string   `json:"contentMD"`
	Metadata  MetaData `json:"metadata"`
	Tags      []string `json:"tags"`
}

//UpdatePostRequest - represents post update request
type UpdatePostRequest struct {
	Title     string   `json:"title"`
	Snippet   string   `json:"snippet"`
	Content   string   `json:"content"`
	ContentMD string   `json:"contentMD"`
	Metadata  MetaData `json:"metadata"`
	Tags      []string `json:"tags"`
}

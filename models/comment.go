package models

import (
	"time"
)

// Comment - represents user's comment
type Comment struct {
	ID       string     `json:"id"`
	PostID   string     `json:"postID"`
	ParentID NullString `json:"parentID"`
	Author   string     `json:"author"`
	Date     time.Time  `json:"date"`
	Content  string     `json:"content"`
	Deleted  bool       `json:"deleted"`
}

// CommentCreateRequest - represents comment creation request
type CommentCreateRequest struct {
	PostID   string     `json:"postID"`
	ParentID NullString `json:"parentID"`
	Author   string     `json:"author"`
	Content  string     `json:"content"`
}

// CommentUpdateRequest - represents comment update request
type CommentUpdateRequest struct {
	Content string `json:"content"`
}

package models

import (
	"time"
)

// Comment - represents user comment
type Comment struct {
	ID       string     `json:"id"`
	PostID   string     `json:"postID"`
	ParentID NullString `json:"parentID"`
	Author   string     `json:"author"`
	Date     time.Time  `json:"date"`
	Content  string     `json:"content"`
	Deleted  bool       `json:"deleted"`
}

// CreateCommentRequest - represents comment creation request
type CreateCommentRequest struct {
	PostID          string      `json:"postID"`
	ParentCommentID interface{} `json:"parentCommentID"` // may be nil
	Author          string      `json:"author"`
	Content         string      `json:"content"`
}

// UpdateCommentRequest - represents comment update request
type UpdateCommentRequest struct {
	Content string `json:"content"`
}

// represents a comment with childs (reply comments)
type CommentWithChilds struct {
	Comment
	Childs []*CommentWithChilds `json:"childs"`
}

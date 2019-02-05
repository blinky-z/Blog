package models

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"time"
)

type NullString sql.NullString

func (ns *NullString) Value() string {
	return ns.String
}

func (ns *NullString) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*ns = NullString{String: s.String, Valid: false}
	} else {
		*ns = NullString{String: s.String, Valid: true}
	}

	return nil
}

// MarshalJSON for NullString
func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON for NullString
func (ns *NullString) UnmarshalJSON(b []byte) error {
	var x interface{}
	err := json.Unmarshal(b, &x)
	if err != nil {
		return err
	}
	switch s := x.(type) {
	case nil:
		ns.Valid = false
	case string:
		ns.String = s
		ns.Valid = true
	}

	return nil
}

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

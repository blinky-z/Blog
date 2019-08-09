package models

// Tag - represents a tag
type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CreateTagRequest- represents tag creation or update HTTP request
type CreateTagRequest struct {
	Name string `json:"name"`
}

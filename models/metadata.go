package models

// MetaData - represents page meta data such as "description" and "keywords"
type MetaData struct {
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

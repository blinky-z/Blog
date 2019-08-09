package models

// MetaData - represents site metadata in <head> tag
type MetaData struct {
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

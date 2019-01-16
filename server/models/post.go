package models

import "time"

type Post struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Date    time.Time `json:"date"`
	Content string    `json:"content"`
}

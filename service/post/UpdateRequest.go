package post

import (
	"github.com/blinky-z/Blog/models"
)

type UpdateRequest struct {
	ID       string
	Title    string
	Author   string
	Snippet  string
	Content  string
	Metadata models.MetaData
}

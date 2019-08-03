package post

import (
	"github.com/blinky-z/Blog/models"
)

type SaveRequest struct {
	Title    string
	Author   string
	Snippet  string
	Content  string
	Metadata models.MetaData
}

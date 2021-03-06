package postService

import (
	"github.com/blinky-z/Blog/models"
)

type UpdateRequest struct {
	ID        string
	Title     string
	Snippet   string
	Content   string
	ContentMD string
	Metadata  models.MetaData
	Tags      []string
}

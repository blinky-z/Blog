package postService

import (
	"github.com/blinky-z/Blog/models"
)

type SaveRequest struct {
	Title     string
	Snippet   string
	Content   string
	ContentMD string
	Metadata  models.MetaData
	Tags      []string
}

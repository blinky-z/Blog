package comment

type SaveRequest struct {
	PostID          string
	ParentCommentID interface{}
	Author          string
	Content         string
}

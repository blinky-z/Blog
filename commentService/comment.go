package commentService

import (
	"github.com/blinky-z/Blog/models"
)

const (
	// DbCommentInputFields - fields that should be filled while inserting new comment
	DbCommentInputFields = "post_id, parent_id, author, content"
	// DbCommentFields - all comment fields
	DbCommentFields = "id, post_id, parent_id, author, date, content, deleted"
)

// GetComments - return all comments of certain post
func GetComments(env *models.Env, postID string) ([]models.Comment, error) {
	env.LogInfo.Print("Got new Comments GET job")

	var comments []models.Comment

	env.LogInfo.Printf("Getting comments of post with id %s from database", postID)

	rows, err := env.Db.Query("select "+DbCommentFields+" from comments where post_id = $1 order by id ASC", postID)
	if err != nil {
		env.LogError.Print(err)
		return comments, err
	}

	for rows.Next() {
		var currentComment models.Comment
		if err = rows.Scan(
			&currentComment.ID, &currentComment.PostID, &currentComment.ParentID, &currentComment.Author, &currentComment.Date,
			&currentComment.Content, &currentComment.Deleted); err != nil {
			env.LogError.Print(err)
			return comments, err
		}

		comments = append(comments, currentComment)
	}

	env.LogInfo.Printf("Comments of post with id %s successfully arrived from database", postID)

	return comments, nil
}

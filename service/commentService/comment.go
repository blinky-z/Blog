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
func GetComments(env *models.Env, postID string) ([]*models.CommentWithChilds, error) {
	env.LogInfo.Print("Got new Comments GET job")

	env.LogInfo.Printf("Getting comments of post with id %s from database", postID)

	rows, err := env.Db.Query("select "+DbCommentFields+" from comments where post_id = $1 order by id ASC", postID)
	if err != nil {
		env.LogError.Print(err)
		return []*models.CommentWithChilds{}, err
	}

	var rawComments []models.Comment
	for rows.Next() {
		var currentComment models.Comment
		if err = rows.Scan(
			&currentComment.ID, &currentComment.PostID, &currentComment.ParentID, &currentComment.Author, &currentComment.Date,
			&currentComment.Content, &currentComment.Deleted); err != nil {
			env.LogError.Print(err)
			return []*models.CommentWithChilds{}, err
		}

		rawComments = append(rawComments, currentComment)
	}

	env.LogInfo.Printf("Comments of post with id %s successfully arrived from database", postID)

	commentWithChildsAsMap := make(map[string]*models.CommentWithChilds)
	var parentComments []string

	for _, comment := range rawComments {
		commentWithChilds := &models.CommentWithChilds{}
		commentWithChilds.Comment = comment

		commentWithChildsAsMap[comment.ID] = commentWithChilds

		if !comment.ParentID.Valid {
			parentComments = append(parentComments, comment.ID)
		}
	}

	for _, comment := range rawComments {
		if comment.ParentID.Valid {
			parent := commentWithChildsAsMap[comment.ParentID.Value().(string)]
			parent.Childs = append(parent.Childs, commentWithChildsAsMap[comment.ID])
			commentWithChildsAsMap[comment.ParentID.Value().(string)] = parent
		}
	}

	var parentCommentWithChilds []*models.CommentWithChilds
	for _, parentCommendID := range parentComments {
		parentCommentWithChilds = append(parentCommentWithChilds, commentWithChildsAsMap[parentCommendID])
	}

	return parentCommentWithChilds, nil
}

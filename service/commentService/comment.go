package commentService

import (
	"database/sql"
	"github.com/blinky-z/Blog/models"
	"html"
)

const (
	// commentsInsertFields - fields that should be filled while inserting a new entity
	commentsInsertFields = "post_id, parent_id, author, content"
	// commentsAllFields - all entity fields
	commentsAllFields = "id, post_id, parent_id, author, date, content, deleted"

	// deletedCommentContent - message to show instead of deleted comment
	DeletedCommentContent = "Содержимое этого комментария было удалено"
)

// Save - saves a new comment in database
// returns a created comment and error
func Save(db *sql.DB, request *SaveRequest) (*models.Comment, error) {
	createdComment := &models.Comment{}
	if err := db.QueryRow("insert into comments ("+commentsInsertFields+") values($1, $2, $3, $4) "+
		"RETURNING "+commentsAllFields,
		request.PostID, request.ParentCommentID, html.EscapeString(request.Author), html.EscapeString(request.Content)).
		Scan(&createdComment.ID, &createdComment.PostID, &createdComment.ParentID, &createdComment.Author,
			&createdComment.Date, &createdComment.Content, &createdComment.Deleted); err != nil {
		return createdComment, err
	}
	return createdComment, nil
}

// Update - updates a new comment in database
// returns an updated comment and error
func Update(db *sql.DB, request *UpdateRequest) (*models.Comment, error) {
	updatedComment := &models.Comment{}
	if err := db.QueryRow("UPDATE comments SET content = $1 WHERE id = $2 RETURNING "+commentsAllFields,
		html.EscapeString(request.NewContent), request.CommentId).
		Scan(&updatedComment.ID, &updatedComment.PostID, &updatedComment.ParentID, &updatedComment.Author,
			&updatedComment.Date, &updatedComment.Content, &updatedComment.Deleted); err != nil {
		return updatedComment, err
	}
	return updatedComment, nil
}

// ExistsById - checks if comment with the given ID exists in database
// returns boolean indicating whether comment exists or not and error
func ExistsById(db *sql.DB, commentId string) (bool, error) {
	var commentExists bool
	err := db.QueryRow("select exists(select 1 from comments where id = $1)", commentId).Scan(&commentExists)
	return commentExists, err
}

// Delete - deletes a comment from database
// if any child exist, comment will not be deleted but content will be replaced with a special deletion message
func Delete(db *sql.DB, commentId string) error {
	var childsExists bool
	if err := db.QueryRow("select exists(select 1 from comments where parent_id = $1)", commentId).
		Scan(&childsExists); err != nil {
		return err
	}

	// TODO: написать тест на то что удаление несуществующего коммента не дает исключения
	if childsExists {
		if _, err := db.Exec("UPDATE comments SET deleted = TRUE, content = $1 WHERE id = $2",
			DeletedCommentContent, commentId); err != nil {
			if err != sql.ErrNoRows {
				return err
			}
		}
	} else {
		if _, err := db.Exec("DELETE from comments where id = $1", commentId); err != nil {
			return err
		}
	}
	return nil
}

// Returns a comment with the given id
// If comment is not exists, function returns sql.ErrNoRows as error
func GetById(db *sql.DB, commentId string) (*models.Comment, error) {
	comment := &models.Comment{}
	err := db.QueryRow("select "+commentsAllFields+" from comments where id = $1", commentId).
		Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.Author, &comment.Date,
			&comment.Content, &comment.Deleted)
	return comment, err
}

// GetAllByPostId - return all comments of the post with the given ID
func GetAllByPostId(db *sql.DB, postID string) ([]*models.CommentWithChilds, error) {
	rows, err := db.Query("select "+commentsAllFields+" from comments where post_id = $1 order by id ASC", postID)
	if err != nil {
		return nil, err
	}

	var rawComments []models.Comment
	for rows.Next() {
		var currentComment models.Comment
		if err = rows.Scan(
			&currentComment.ID, &currentComment.PostID, &currentComment.ParentID, &currentComment.Author, &currentComment.Date,
			&currentComment.Content, &currentComment.Deleted); err != nil {
			return nil, err
		}

		rawComments = append(rawComments, currentComment)
	}

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

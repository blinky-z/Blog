package api

import (
	"database/sql"
	"encoding/json"
	"github.com/blinky-z/Blog/commentService"
	"github.com/blinky-z/Blog/models"
	"github.com/gorilla/mux"
	"html"
	"net/http"
)

// CommentAPI - environment container struct to declare all comment handlers as methods
type CommentAPI struct {
	Env *models.Env
}

var (
	// CommentEnv - instance of CommentAPI struct. Initialized by main
	CommentEnv CommentAPI
)

const (
	//NoSuchComment - comment with requested id doesn't exist
	NoSuchComment PostErrorCode = "NO_SUCH_COMMENT"

	// MaxCommentContentLen - max length of comment content
	MaxCommentContentLen int = 2048

	// DeletedCommentContent - message that replaces content of deleted comment
	DeletedCommentContent = "Содержимое этого комментария было удалено"
)

func validateCommentContent(content string) PostErrorCode {
	contentLen := len(content)
	if contentLen > MaxCommentContentLen || contentLen == 0 {
		return InvalidContent
	}

	return NoError
}

func validateCreateComment(r *http.Request) (comment models.CommentCreateRequest, validateError PostErrorCode) {
	validateError = NoError
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		validateError = BadRequestBody
		return
	}

	validatePostIDError := ValidateID(comment.PostID)
	if validatePostIDError != NoError {
		validateError = validatePostIDError
		return
	}

	if comment.ParentID.Valid {
		validateParentIDError := ValidateID(comment.ParentID.Value().(string))
		if validateParentIDError != NoError {
			validateError = validateParentIDError
			return
		}
	}

	validateCommentContentError := validateCommentContent(comment.Content)
	if validateCommentContentError != NoError {
		validateError = validateCommentContentError
		return
	}

	authorLen := len(comment.Author)
	if authorLen > MaxLoginLen || authorLen < MinLoginLen || authorLen == 0 {
		validateError = InvalidLogin
		return
	}

	return
}

func validateUpdateComment(r *http.Request) (comment models.CommentUpdateRequest, validateError PostErrorCode) {
	validateError = NoError
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		validateError = BadRequestBody
		return
	}

	validateCommentContentError := validateCommentContent(comment.Content)
	if validateCommentContentError != NoError {
		validateError = validateCommentContentError
		return
	}

	return
}

// CreateComment - create comment http handler
func (api *CommentAPI) CreateComment() http.Handler {
	env := api.Env
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.LogInfo.Print("Got new Comment CREATE job")

		comment, validateCommentError := validateCreateComment(r)
		if validateCommentError != NoError {
			env.LogInfo.Print("Can't create comment: comment is invalid")
			RespondWithError(w, http.StatusBadRequest, validateCommentError, env.LogError)
			return
		}

		if err := env.Db.QueryRow("select from posts where id = $1", comment.PostID).Scan(); err != nil {
			if err == sql.ErrNoRows {
				env.LogInfo.Printf("Can not CREATE comment to post with id %s : post does not exist", comment.PostID)
				RespondWithError(w, http.StatusNotFound, NoSuchPost, env.LogError)
				return
			}

			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		if comment.ParentID.Valid {
			var parentCommentPostID string

			if err := env.Db.QueryRow("select post_id from comments where id = $1", comment.ParentID.Value()).
				Scan(&parentCommentPostID); err != nil {
				if err == sql.ErrNoRows {
					env.LogInfo.Printf("Can not CREATE (reply) comment to parent comment with id %s to post with id %s : "+
						"parent comment does not exist", comment.ParentID.Value(), comment.PostID)
					RespondWithError(w, http.StatusNotFound, NoSuchComment, env.LogError)
					return
				}

				env.LogError.Print(err)
				RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
				return
			}

			if parentCommentPostID != comment.PostID {
				env.LogInfo.Printf("Can not CREATE (reply) comment to parent comment with id %s to post with id %s : "+
					"parent comment belongs to not this post, but to post with id %s",
					comment.ParentID.Value(), comment.PostID, parentCommentPostID)
				RespondWithError(w, http.StatusNotFound, NoSuchComment, env.LogError)
				return
			}
		}

		var createdComment models.Comment

		env.LogInfo.Printf("Inserting comment with id %s into database", comment.PostID)

		if _, err := env.Db.Exec("BEGIN TRANSACTION"); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}
		if err := env.Db.QueryRow("insert into comments("+commentService.DbCommentInputFields+") values($1, $2, $3, $4) "+
			"RETURNING "+commentService.DbCommentFields,
			comment.PostID, comment.ParentID.Value(), html.EscapeString(comment.Author), html.EscapeString(comment.Content)).
			Scan(&createdComment.ID, &createdComment.PostID, &createdComment.ParentID, &createdComment.Author,
				&createdComment.Date, &createdComment.Content, &createdComment.Deleted); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}
		if _, err := env.Db.Exec("END TRANSACTION"); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("Comment with id %s to post with id %s successfully created",
			createdComment.ID, createdComment.PostID)

		RespondWithBody(w, http.StatusCreated, &createdComment, env.LogError)
	})
}

// UpdateComment - update comment http handler
func (api *CommentAPI) UpdateComment() http.Handler {
	env := api.Env
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.LogInfo.Print("Got new Comment UPDATE job")

		userRole := r.Context().Value(CtxKey).(models.UserRole)
		if userRole != roleAdmin {
			env.LogInfo.Printf("User with role %s doesn't have permissions to UPDATE comment", userRole)
			RespondWithError(w, http.StatusForbidden, NoPermissions, env.LogError)
			return
		}

		id := mux.Vars(r)["id"]
		validateIDError := ValidateID(id)
		if validateIDError != NoError {
			env.LogInfo.Print("Can not UPDATE comment: ID of Comment to update is invalid")
			RespondWithError(w, http.StatusBadRequest, validateIDError, env.LogError)
			return
		}

		comment, validateCommentError := validateUpdateComment(r)
		if validateCommentError != NoError {
			env.LogInfo.Print("Can not UPDATE comment: new comment is invalid")
			RespondWithError(w, http.StatusBadRequest, validateCommentError, env.LogError)
			return
		}

		if err := env.Db.QueryRow("select from comments where id = $1", id).Scan(); err != nil {
			if err == sql.ErrNoRows {
				env.LogInfo.Printf("Can not UPDATE comment with id %s : comment does not exist", id)
				RespondWithError(w, http.StatusNotFound, NoSuchComment, env.LogError)
				return
			}

			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("Updating comment with ID %s in database", id)

		var updatedComment models.Comment

		if _, err := env.Db.Exec("BEGIN TRANSACTION"); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}
		if err := env.Db.QueryRow("UPDATE comments SET content = $1 WHERE id = $2 RETURNING "+commentService.DbCommentFields,
			html.EscapeString(comment.Content), id).
			Scan(&updatedComment.ID, &updatedComment.PostID, &updatedComment.ParentID, &updatedComment.Author,
				&updatedComment.Date, &updatedComment.Content, &updatedComment.Deleted); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}
		if _, err := env.Db.Exec("END TRANSACTION"); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("Comment with ID %s successfully updated", id)

		RespondWithBody(w, http.StatusOK, &updatedComment, env.LogError)
	})
}

// DeleteComment - delete comment http handler
func (api *CommentAPI) DeleteComment() http.Handler {
	env := api.Env
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.LogInfo.Print("Got new Comment DELETE job")

		userRole := r.Context().Value(CtxKey).(models.UserRole)
		if userRole != roleAdmin {
			env.LogInfo.Printf("User with role %s doesn't have permissions to DELETE comment", userRole)
			RespondWithError(w, http.StatusForbidden, NoPermissions, env.LogError)
			return
		}

		commentID := mux.Vars(r)["id"]
		validateIDError := ValidateID(commentID)
		if validateIDError != NoError {
			env.LogInfo.Print("Can not DELETE comment: ID of Comment to delete is invalid")
			RespondWithError(w, http.StatusBadRequest, validateIDError, env.LogError)
			return
		}

		if err := env.Db.QueryRow("select from comments where id = $1", commentID).Scan(); err != nil {
			if err == sql.ErrNoRows {
				env.LogInfo.Printf("Can not DELETE comment with id %s : comment does not exist", commentID)
				RespondWithError(w, http.StatusNotFound, NoSuchComment, env.LogError)
				return
			}

			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("Marking comment with ID %s as deleted in database", commentID)

		var childsExists bool
		if err := env.Db.QueryRow("select exists(select 1 from comments where parent_id = $1)", commentID).
			Scan(&childsExists); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		if _, err := env.Db.Exec("BEGIN TRANSACTION"); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}
		if childsExists {
			if _, err := env.Db.Exec("UPDATE comments SET deleted = TRUE, content = $1 WHERE id = $2",
				DeletedCommentContent, commentID); err != nil {
				env.LogError.Print(err)
				RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
				return
			}
		} else {
			if _, err := env.Db.Exec("DELETE from comments where id = $1", commentID); err != nil {
				env.LogError.Print(err)
				RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
				return
			}
		}
		if _, err := env.Db.Exec("END TRANSACTION"); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("Comment with ID %s successfully deleted", commentID)

		Respond(w, http.StatusOK)
	})
}

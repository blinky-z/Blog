package restapi

import (
	"database/sql"
	"encoding/json"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/service/comment"
	"github.com/blinky-z/Blog/service/post"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// CommentAPIHandler - used for dependency injection
type CommentAPIHandler struct {
	db       *sql.DB
	logInfo  *log.Logger
	logError *log.Logger
}

func NewCommentAPIHandler(db *sql.DB, logInfo, logError *log.Logger) *CommentAPIHandler {
	return &CommentAPIHandler{
		db:       db,
		logInfo:  logInfo,
		logError: logError,
	}
}

// error codes for this API
const (
	//NoSuchComment - comment does not exist
	NoSuchComment RequestErrorCode = "NO_SUCH_COMMENT"
	// InvalidCommentContent - invalid comment content
	InvalidCommentContent RequestErrorCode = "Invalid comment content"
)

// constants for use in validator methods
const (
	// MinCommentContentLen - comment's content max length
	MinCommentContentLen int = 10
	// MaxCommentContentLen - comment's content max length
	MaxCommentContentLen int = 4096
)

func validateCommentContent(content string) RequestErrorCode {
	content = strings.TrimSpace(content)
	contentLen := len(content)
	if contentLen > MaxCommentContentLen || contentLen < MinCommentContentLen {
		return InvalidCommentContent
	}

	return NoError
}

func validateCreateCommentRequest(request models.CreateCommentRequest) RequestErrorCode {
	if !isCommentIDValid(request.PostID) {
		return InvalidRequest
	}
	if request.ParentCommentID != nil {
		if !isCommentIDValid(request.ParentCommentID.(string)) {
			return InvalidRequest
		}
	}
	if validateAuthorError := validateUsername(request.Author); validateAuthorError != NoError {
		return validateAuthorError
	}
	if validateContentError := validateCommentContent(request.Content); validateContentError != NoError {
		return validateContentError
	}

	return NoError
}

func validateUpdateCommentRequest(request models.UpdateCommentRequest) RequestErrorCode {
	if validateCommentContentError := validateCommentContent(request.Content); validateCommentContentError != NoError {
		return validateCommentContentError
	}

	return NoError
}

func isCommentIDValid(id string) bool {
	if id == "" {
		return false
	}
	num, err := strconv.Atoi(id)
	if err != nil || num < 0 {
		return false
	}

	return true
}

// CreateCommentHandler - this handler serves comment creation requests
func (api *CommentAPIHandler) CreateCommentHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := models.CreateCommentRequest{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, BadRequestBody)
			return
		}

		logInfo.Printf("Got new comment creation request. Request: %+v", request)

		validateRequestError := validateCreateCommentRequest(request)
		if validateRequestError != NoError {
			logError.Printf("Can't create comment: invalid request. Error: %s", validateRequestError)
			RespondWithError(w, http.StatusBadRequest, validateRequestError)
			return
		}

		postID := request.PostID

		if isPostExists, err := post.ExistsByID(api.db, postID); err != nil {
			logError.Printf("Can't create comment: error checking post for presence. Post ID: %s. Error: %s",
				postID, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		} else if !isPostExists {
			logError.Printf("Can't create comment: post does not exist. Post ID: %s", postID)
			RespondWithError(w, http.StatusBadRequest, InvalidRequest)
			return
		}

		if request.ParentCommentID != nil {
			parentCommentID := request.ParentCommentID.(string)
			parentComment, err := comment.GetByID(api.db, parentCommentID)
			if err != nil {
				if err == sql.ErrNoRows {
					logError.Printf("Can't create comment: parent comment does not exist. Parent comment ID: %s",
						parentCommentID)
					RespondWithError(w, http.StatusBadRequest, InvalidRequest)
					return
				}

				logError.Printf("Can't create comment: error checking parent comment for presence. "+
					"Parent comment ID: %s. Error: %s", parentCommentID, err)
				RespondWithError(w, http.StatusInternalServerError, TechnicalError)
				return
			}
			if parentComment.PostID != postID {
				logError.Printf("Can't create comment: parent comment belongs to the other post. Parent comment ID: %s",
					parentCommentID)
				RespondWithError(w, http.StatusBadRequest, InvalidRequest)
				return
			}
		}

		saveRequest := &comment.SaveRequest{
			PostID:          postID,
			ParentCommentID: request.ParentCommentID,
			Author:          request.Author,
			Content:         request.Content,
		}
		createdComment, err := comment.Save(api.db, saveRequest)
		if err != nil {
			logError.Printf("Error saving comment in database: %s", err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		logInfo.Printf("Comment saved. Comment: %+v", createdComment)
		RespondWithBody(w, http.StatusCreated, createdComment)
	})
}

// UpdateCommentHandler - this handler serves comment update requests
func (api *CommentAPIHandler) UpdateCommentHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := models.UpdateCommentRequest{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, BadRequestBody)
			return
		}

		commentID := mux.Vars(r)["id"]
		logInfo.Printf("Got new comment update request. Comment ID: %s, Request: %+v", commentID, request)

		if !isCommentIDValid(commentID) {
			logError.Printf("Can't update comment: invalid comment ID. Comment ID: %s", commentID)
			RespondWithError(w, http.StatusBadRequest, InvalidRequest)
			return
		}

		validateRequestError := validateUpdateCommentRequest(request)
		if validateRequestError != NoError {
			logError.Printf("Can't update comment: invalid request. Comment ID: %s. Error: %s",
				commentID, validateRequestError)
			RespondWithError(w, http.StatusBadRequest, validateRequestError)
			return
		}

		userRole := r.Context().Value(CtxRoleKey).(models.UserRole)
		if userRole != roleAdmin {
			RespondWithError(w, http.StatusForbidden, NoPermissions)
			return
		}

		isCommentExists, err := comment.ExistsByID(api.db, commentID)
		if err != nil {
			logError.Printf("Can't update comment: error checking comment for presence. Comment ID: %s. Error: %s",
				commentID, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}
		if !isCommentExists {
			logInfo.Printf("Can't update comment: comment does not exist. Comment ID: %s", commentID)
			RespondWithError(w, http.StatusBadRequest, NoSuchComment)
			return
		}

		updateRequest := &comment.UpdateRequest{
			CommentID:  commentID,
			NewContent: request.Content,
		}
		updatedComment, err := comment.Update(api.db, updateRequest)
		if err != nil {
			logError.Printf("Error updating comment in database. Comment ID: %s. Error: %s", commentID, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		logInfo.Printf("Comment updated. Comment ID: %s", commentID)
		RespondWithBody(w, http.StatusCreated, updatedComment)
	})
}

// DeleteCommentHandler - this handler server comment deletion requests
func (api *CommentAPIHandler) DeleteCommentHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value(CtxRoleKey).(models.UserRole)
		if userRole != roleAdmin {
			RespondWithError(w, http.StatusForbidden, NoPermissions)
			return
		}

		commentID := mux.Vars(r)["id"]
		logInfo.Printf("Got new comment deletion request. Comment ID: %s", commentID)

		if !isCommentIDValid(commentID) {
			logError.Printf("Can't delete comment: invalid comment ID. Comment ID: %s", commentID)
			RespondWithError(w, http.StatusBadRequest, InvalidRequest)
			return
		}

		if err := comment.Delete(api.db, commentID); err != nil {
			logError.Printf("Error deleting comment from database. Comment ID: %s. Error: %s", commentID, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		logInfo.Printf("Comment deleted. Comment ID: %s", commentID)
		Respond(w, http.StatusOK)
	})
}

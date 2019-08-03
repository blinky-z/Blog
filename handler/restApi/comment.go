package restApi

import (
	"database/sql"
	"encoding/json"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/service/commentService"
	"github.com/blinky-z/Blog/service/postService"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// CommentApiHandler - used for dependency injection
type CommentApiHandler struct {
	db       *sql.DB
	logInfo  *log.Logger
	logError *log.Logger
}

func NewCommentApiHandler(db *sql.DB, logInfo, logError *log.Logger) *CommentApiHandler {
	return &CommentApiHandler{
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
	if !isCommentIdValid(request.PostID) {
		return InvalidRequest
	}
	if request.ParentCommentID != nil {
		if !isCommentIdValid(request.ParentCommentID.(string)) {
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

func isCommentIdValid(id string) bool {
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
func (api *CommentApiHandler) CreateCommentHandler() http.Handler {
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

		postId := request.PostID

		if isPostExists, err := postService.ExistsById(api.db, postId); err != nil {
			logError.Printf("Can't create comment: error checking post for presence. Post ID: %s. Error: %s",
				postId, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		} else if !isPostExists {
			logError.Printf("Can't create comment: post does not exist. Post ID: %s", postId)
			RespondWithError(w, http.StatusBadRequest, InvalidRequest)
			return
		}

		if request.ParentCommentID != nil {
			parentCommentId := request.ParentCommentID.(string)
			parentComment, err := commentService.GetById(api.db, parentCommentId)
			if err != nil {
				if err == sql.ErrNoRows {
					logError.Printf("Can't create comment: parent comment does not exist. Parent comment ID: %s",
						parentCommentId)
					RespondWithError(w, http.StatusBadRequest, InvalidRequest)
					return
				}

				logError.Printf("Can't create comment: error checking parent comment for presence. Parent comment ID: %s. Error: %s",
					parentCommentId, err)
				RespondWithError(w, http.StatusInternalServerError, TechnicalError)
				return
			}
			if parentComment.PostID != postId {
				logError.Printf("Can't create comment: parent comment belongs to the other post. Parent comment ID: %s",
					parentCommentId)
				RespondWithError(w, http.StatusBadRequest, InvalidRequest)
				return
			}
		}

		saveRequest := &commentService.SaveRequest{
			PostID:          postId,
			ParentCommentID: request.ParentCommentID,
			Author:          request.Author,
			Content:         request.Content,
		}
		createdComment, err := commentService.Save(api.db, saveRequest)
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
func (api *CommentApiHandler) UpdateCommentHandler() http.Handler {
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

		if !isCommentIdValid(commentID) {
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

		isCommentExists, err := commentService.ExistsById(api.db, commentID)
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

		updateRequest := &commentService.UpdateRequest{
			CommentId:  commentID,
			NewContent: request.Content,
		}
		updatedComment, err := commentService.Update(api.db, updateRequest)
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
func (api *CommentApiHandler) DeleteCommentHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value(CtxRoleKey).(models.UserRole)
		if userRole != roleAdmin {
			RespondWithError(w, http.StatusForbidden, NoPermissions)
			return
		}

		commentId := mux.Vars(r)["id"]
		logInfo.Printf("Got new comment deletion request. Comment ID: %s", commentId)

		if !isCommentIdValid(commentId) {
			logError.Printf("Can't delete comment: invalid comment ID. Comment ID: %s", commentId)
			RespondWithError(w, http.StatusBadRequest, InvalidRequest)
			return
		}

		if err := commentService.Delete(api.db, commentId); err != nil {
			logError.Printf("Error deleting comment from database. Comment ID: %s. Error: %s", commentId, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		logInfo.Printf("Comment deleted. Comment ID: %s", commentId)
		Respond(w, http.StatusOK)
	})
}

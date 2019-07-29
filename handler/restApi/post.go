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

// PostApiHandler - used for dependency injection
type PostApiHandler struct {
	db       *sql.DB
	logInfo  *log.Logger
	logError *log.Logger
}

func NewPostApiHandler(db *sql.DB, logInfo, logError *log.Logger) *PostApiHandler {
	return &PostApiHandler{
		db:       db,
		logInfo:  logInfo,
		logError: logError,
	}
}

// GetPostsRequestQueryParams - structure for storing query params of GET request for range of posts
type GetPostsRequestQueryParams struct {
	Page         string
	PostsPerPage string
}

// error codes for this API
const (
	// InvalidPostTitle - invalid post title
	InvalidPostTitle RequestErrorCode = "INVALID_TITLE"
	// InvalidID - invalid post id
	InvalidID RequestErrorCode = "INVALID_ID" // TODO: решить что делать с ID. ID есть везде, а не только в постах
	// InvalidPostSnippet - invalid post snippet
	InvalidPostSnippet RequestErrorCode = "INVALID_SNIPPET"
	// InvalidPostContent - invalid post content
	InvalidPostContent RequestErrorCode = "INVALID_CONTENT"
	// InvalidPostMetadata - invalid meta description or keywords
	InvalidPostMetadata RequestErrorCode = "INVALID_METADATA"
	// NoSuchPost - post does not exist
	NoSuchPost RequestErrorCode = "NO_SUCH_POST"
	// InvalidPostsRange - invalid range of posts
	InvalidPostsRange RequestErrorCode = "INVALID_POSTS_RANGE" // TODO: нужно ли это вообще?
)

// constants for use in validator methods
const (
	// MinPostTitleLen - maximum post title length
	MinPostTitleLen int = 10
	// MaxPostTitleLen - maximum post title length
	MaxPostTitleLen int = 120

	// MaxPostsPerPage - maximum posts that can be displayed per page
	MaxPostsPerPage int = 40

	// MinMetaDescriptionLen - min meta description length.
	MinMetaDescriptionLen int = 40
	// MaxMetaDescriptionLen - max meta description length.
	// 160 is a good value for search engines
	MaxMetaDescriptionLen int = 160

	// MinMetaKeywordsAmount - min allowed amount of meta keywords
	// 4 is a good value for search engines
	MinMetaKeywordsAmount int = 1
	// MaxMetaKeywordsAmount - max allowed amount of meta keywords. Don't overuse keywords.
	// 4 is a good value for search engines
	MaxMetaKeywordsAmount int = 4
	// MinMetaKeywordLen - min length of each meta keyword
	MinMetaKeywordLen int = 4
	// MaxMetaKeywordLen - max length of each meta keyword
	MaxMetaKeywordLen int = 20

	// MaxSnippetLen - max length of post snippet
	MaxSnippetLen int = 350
	// MinSnippetLen - min length of post snippet
	MinSnippetLen int = 40
)

// other API constants
const (
	defaultPage         string = "0"
	defaultPostsPerPage string = "10"
)

// ValidateGetPostsRequestQueryParams - validate query params of GET request for range of posts
func ValidateGetPostsRequestQueryParams(rangeParams *GetPostsRequestQueryParams) RequestErrorCode {
	pageAsString := rangeParams.Page
	if pageAsString != "" {
		if pageAsInt, err := strconv.Atoi(pageAsString); err != nil || pageAsInt < 0 {
			return InvalidPostsRange
		}
	}
	postsPerPageAsString := rangeParams.PostsPerPage
	if postsPerPageAsString != "" {
		if postsPerPageAsInt, err := strconv.Atoi(rangeParams.PostsPerPage); err != nil ||
			postsPerPageAsInt > MaxPostsPerPage || postsPerPageAsInt < 0 {
			return InvalidPostsRange
		}
	}

	return NoError
}

func validatePostTitle(title *string) RequestErrorCode {
	titleLen := len(strings.TrimSpace(*title))
	if titleLen > MaxPostTitleLen || titleLen < MinPostTitleLen {
		return InvalidPostTitle
	}
	return NoError
}

func validatePostMetadata(metadata *models.MetaData) RequestErrorCode {
	descriptionLen := len(strings.TrimSpace(metadata.Description))
	if descriptionLen > MaxMetaDescriptionLen || descriptionLen < MinMetaDescriptionLen {
		return InvalidPostMetadata
	}

	keywordsAmount := len(metadata.Keywords)
	if keywordsAmount > MaxMetaKeywordsAmount || keywordsAmount < MinMetaKeywordsAmount {
		return InvalidPostMetadata
	}

	for _, keyword := range metadata.Keywords {
		keywordLen := len(strings.TrimSpace(keyword))
		if keywordLen > MaxMetaKeywordLen || keywordLen < MinMetaKeywordLen {
			return InvalidPostMetadata
		}
	}
	return NoError
}

func validatePostSnippet(snippet *string) RequestErrorCode {
	snippetLen := len(strings.TrimSpace(*snippet))
	if snippetLen > MaxSnippetLen || snippetLen < MinSnippetLen {
		return InvalidPostSnippet
	}
	return NoError
}

func validatePostContent(content *string) RequestErrorCode {
	if len(*content) == 0 {
		return InvalidPostContent
	}
	return NoError
}

func validateCreatePostRequest(request *models.CreatePostRequest) RequestErrorCode {
	if err := validatePostTitle(&request.Title); err != NoError {
		return err
	}
	if err := validatePostMetadata(&request.Metadata); err != NoError {
		return err
	}
	if err := validateUsername(request.Author); err != NoError {
		return err
	}
	if err := validatePostSnippet(&request.Snippet); err != NoError {
		return err
	}
	if err := validatePostContent(&request.Content); err != NoError {
		return err
	}

	return NoError
}

func validateUpdatePostRequest(request *models.UpdatePostRequest) RequestErrorCode {
	if err := validatePostTitle(&request.Title); err != NoError {
		return err
	}
	if err := validatePostMetadata(&request.Metadata); err != NoError {
		return err
	}
	if err := validateUsername(request.Author); err != NoError {
		return err
	}
	if err := validatePostSnippet(&request.Snippet); err != NoError {
		return err
	}
	if err := validatePostContent(&request.Content); err != NoError {
		return err
	}

	return NoError
}

// TODO: тоже не должно здесь быть
// ValidateID - validates id of post or comment
func ValidateID(idAsString string) (validateError RequestErrorCode) {
	validateError = NoError

	if len(idAsString) == 0 {
		validateError = InvalidID
		return
	}

	num, err := strconv.Atoi(idAsString)
	if err != nil {
		validateError = InvalidID
		return
	}

	if num < 0 {
		validateError = InvalidID
		return
	}

	return
}

// CreatePostHandler - this handler server post creation requests
func (api *PostApiHandler) CreatePostHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value(CtxRoleKey).(models.UserRole)
		if userRole != roleAdmin {
			logError.Printf("User role doesn't have permissions to create posts. User role: %s", userRole)
			RespondWithError(w, http.StatusForbidden, NoPermissions)
			return
		}

		request := models.CreatePostRequest{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, BadRequestBody)
			return
		}

		logInfo.Printf("Got new post creation request. Request: %+v", request)

		validatePostError := validateCreatePostRequest(&request)
		if validatePostError != NoError {
			logError.Printf("Can't create post: invalid request. Error: %s", validatePostError)
			RespondWithError(w, http.StatusBadRequest, validatePostError)
			return
		}

		saveRequest := &postService.SaveRequest{
			Title:    request.Title,
			Author:   request.Author,
			Snippet:  request.Snippet,
			Content:  request.Content,
			Metadata: request.Metadata,
		}
		createdPost, err := postService.Save(api.db, saveRequest)
		if err != nil {
			logError.Printf("Error saving post in database: %s", err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		logInfo.Printf("Post saved. Post: %+v", createdPost)
		RespondWithBody(w, http.StatusCreated, createdPost)
	})
}

// UpdatePostHandler - this handler serves post update requests
func (api *PostApiHandler) UpdatePostHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value(CtxRoleKey).(models.UserRole)
		if userRole != roleAdmin {
			RespondWithError(w, http.StatusForbidden, NoPermissions)
			return
		}

		request := models.UpdatePostRequest{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, BadRequestBody)
			return
		}

		postId := mux.Vars(r)["id"]
		logInfo.Printf("Got new post update request. Post ID: %s", postId)

		validateIDError := ValidateID(postId)
		if validateIDError != NoError {
			logError.Printf("Can't update post: invalid post ID. Post ID: %s", postId)
			RespondWithError(w, http.StatusBadRequest, validateIDError)
			return
		}

		validatePostError := validateUpdatePostRequest(&request)
		if validatePostError != NoError {
			logInfo.Printf("Can't update post: invalid request. Post ID: %s. Error: %s", postId, validatePostError)
			RespondWithError(w, http.StatusBadRequest, validatePostError)
			return
		}

		isPostExists, err := postService.ExistsById(api.db, postId)
		if err != nil {
			logError.Printf("Can't update post: error checking post for presence. Post ID: %s. Error: %s",
				postId, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}
		if !isPostExists {
			logInfo.Printf("Can't update post: post does not exist. Post ID: %s", postId)
			RespondWithError(w, http.StatusBadRequest, NoSuchPost)
			return
		}

		updateRequest := &postService.UpdateRequest{
			ID:       postId,
			Title:    request.Title,
			Author:   request.Author,
			Snippet:  request.Snippet,
			Content:  request.Content,
			Metadata: request.Metadata,
		}
		updatedPost, err := postService.Update(api.db, updateRequest)
		if err != nil {
			logError.Printf("Error updating post in database: %s", err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		logInfo.Printf("Post updated. Post: %+v", updatedPost)
		RespondWithBody(w, http.StatusCreated, updatedPost)
	})
}

// DeletePostHandler - this handler serves post deletion requests
func (api *PostApiHandler) DeletePostHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value(CtxRoleKey).(models.UserRole)
		if userRole != roleAdmin {
			RespondWithError(w, http.StatusForbidden, NoPermissions)
			return
		}

		postId := mux.Vars(r)["id"]
		logInfo.Printf("Got new post deletion request. Post ID: %s", postId)

		validateIDError := ValidateID(postId)
		if validateIDError != NoError {
			logError.Printf("Can't delete post: invalid post ID. Post ID: %s", postId)
			RespondWithError(w, http.StatusBadRequest, validateIDError)
			return
		}

		if err := postService.Delete(api.db, postId); err != nil {
			logError.Printf("Error deleting post. Post ID: %s. Error: %s", postId, err)
		}

		logInfo.Printf("Post deleted. Post ID: %s", postId)
		Respond(w, http.StatusOK)
	})
}

// GetCertainPostHandler - this handler serves GET request for single post
func (api *PostApiHandler) GetCertainPostHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postId := mux.Vars(r)["id"]
		logInfo.Printf("Got single post retrieve request. Post ID: %s", postId)

		validateIDError := ValidateID(postId)
		if validateIDError != NoError {
			logError.Printf("Can't retrieve post: invalid post ID. Post ID: %s", postId)
			RespondWithError(w, http.StatusBadRequest, validateIDError)
			return
		}

		post, err := postService.GetById(api.db, postId)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				logError.Printf("Can't retrieve post: no such post. Post ID: %s", postId)
				RespondWithError(w, http.StatusBadRequest, NoSuchPost)
				return
			default:
				logError.Printf("Error retrieving post from database. Post ID: %s. Error: %s", postId, err)
				RespondWithError(w, http.StatusInternalServerError, TechnicalError)
				return
			}
		}

		comments, err := commentService.GetAllByPostId(api.db, postId)
		if err != nil {
			logError.Printf("Error retrieving post from database: error retrieving comments. Post ID: %s. Error: %s",
				postId, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		certainPostResponse := &models.CertainPostResponse{
			Post:     post,
			Comments: comments,
		}
		RespondWithBody(w, http.StatusOK, certainPostResponse)
	})
}

// GetPostsHandler - this handler serves GET request for all posts in the given range
func (api *PostApiHandler) GetPostsHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeParams := &GetPostsRequestQueryParams{
			Page:         r.FormValue("page"),
			PostsPerPage: r.FormValue("posts-per-page"),
		}

		logInfo.Printf("Got range of posts retrieve request. Range params: %+v", rangeParams)

		validateQueryParamsError := ValidateGetPostsRequestQueryParams(rangeParams)
		if validateQueryParamsError != NoError {
			logError.Printf("Can't retrieve range of posts: invalid query params. Error: %s", validateQueryParamsError)
			RespondWithError(w, http.StatusBadRequest, validateQueryParamsError)
			return
		}

		// set default values if params are missed
		pageAsString := rangeParams.Page
		postsPerPageAsString := rangeParams.PostsPerPage
		if pageAsString == "" {
			pageAsString = defaultPage
		}
		if postsPerPageAsString == "" {
			postsPerPageAsString = defaultPostsPerPage
		}

		// we know that params are valid, so ignore the errors
		pageAsInt, _ := strconv.Atoi(pageAsString)
		postsPerPageAsInt, _ := strconv.Atoi(pageAsString)

		posts, err := postService.GetPostsInRange(api.db, pageAsInt, postsPerPageAsInt)
		if err != nil {
			logError.Printf("Error retrieving range of posts from database: %s", err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		RespondWithBody(w, http.StatusOK, posts)
	})
}

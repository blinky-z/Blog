package restapi

import (
	"database/sql"
	"encoding/json"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/service/postService"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// PostAPIHandler - used for dependency injection
type PostAPIHandler struct {
	db       *sql.DB
	logInfo  *log.Logger
	logError *log.Logger
}

func NewPostAPIHandler(db *sql.DB, logInfo, logError *log.Logger) *PostAPIHandler {
	return &PostAPIHandler{
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
var (
	// InvalidPostTitle - invalid post title
	InvalidPostTitle = models.NewRequestErrorCode("INVALID_TITLE")
	// InvalidPostSnippet - invalid post snippet
	InvalidPostSnippet = models.NewRequestErrorCode("INVALID_SNIPPET")
	// InvalidPostContent - invalid post content
	InvalidPostContent = models.NewRequestErrorCode("INVALID_CONTENT")
	// InvalidPostMetadata - invalid meta description or keywords
	InvalidPostMetadata = models.NewRequestErrorCode("INVALID_METADATA")
	// NoSuchPost - post does not exist
	NoSuchPost = models.NewRequestErrorCode("NO_SUCH_POST")
	// InvalidPostsRange - invalid range of posts
	InvalidPostsRange = models.NewRequestErrorCode("INVALID_POSTS_RANGE")
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
	MinMetaKeywordsAmount int = 0
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
	DefaultPage         string = "0"
	DefaultPostsPerPage string = "10"
)

// ValidateGetPostsRequestQueryParams - validate query params of GET request for range of posts
func ValidateGetPostsRequestQueryParams(rangeParams *GetPostsRequestQueryParams) models.RequestErrorCode {
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

	return nil
}

func validatePostTitle(title *string) models.RequestErrorCode {
	titleLen := len(strings.TrimSpace(*title))
	if titleLen > MaxPostTitleLen || titleLen < MinPostTitleLen {
		return InvalidPostTitle
	}
	return nil
}

func validatePostMetadata(metadata *models.MetaData) models.RequestErrorCode {
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
	return nil
}

func validatePostSnippet(snippet *string) models.RequestErrorCode {
	snippetLen := len(strings.TrimSpace(*snippet))
	if snippetLen > MaxSnippetLen || snippetLen < MinSnippetLen {
		return InvalidPostSnippet
	}
	return nil
}

func validatePostContent(content *string) models.RequestErrorCode {
	if len(*content) == 0 {
		return InvalidPostContent
	}
	return nil
}

func validateCreatePostRequest(request *models.CreatePostRequest) models.RequestErrorCode {
	if err := validatePostTitle(&request.Title); err != nil {
		return err
	}
	if err := validatePostMetadata(&request.Metadata); err != nil {
		return err
	}
	if err := validatePostSnippet(&request.Snippet); err != nil {
		return err
	}
	if err := validatePostContent(&request.Content); err != nil {
		return err
	}

	return nil
}

func validateUpdatePostRequest(request *models.UpdatePostRequest) models.RequestErrorCode {
	if err := validatePostTitle(&request.Title); err != nil {
		return err
	}
	if err := validatePostMetadata(&request.Metadata); err != nil {
		return err
	}
	if err := validatePostSnippet(&request.Snippet); err != nil {
		return err
	}
	if err := validatePostContent(&request.Content); err != nil {
		return err
	}

	return nil
}

func IsPostIDValid(id string) bool {
	if id == "" {
		return false
	}
	num, err := strconv.Atoi(id)
	if err != nil || num < 0 {
		return false
	}

	return true
}

// CreatePostHandler - this handler server post creation requests
func (api *PostAPIHandler) CreatePostHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//userRole := r.Context().Value(CtxRoleKey).(models.UserRole)
		//if userRole != roleAdmin {
		//	logError.Printf("User role doesn't have permissions to create posts. User role: %s", userRole)
		//	RespondWithError(w, http.StatusForbidden, NoPermissions)
		//	return
		//}

		request := models.CreatePostRequest{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, BadRequestBody)
			return
		}

		logInfo.Printf("Got new post creation request. Request: %+v", request)

		validatePostError := validateCreatePostRequest(&request)
		if validatePostError != nil {
			logError.Printf("Can't create post: invalid request. Error: %s", validatePostError)
			RespondWithError(w, http.StatusBadRequest, validatePostError)
			return
		}

		saveRequest := &postService.SaveRequest{
			Title:    request.Title,
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
func (api *PostAPIHandler) UpdatePostHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//userRole := r.Context().Value(CtxRoleKey).(models.UserRole)
		//if userRole != roleAdmin {
		//	RespondWithError(w, http.StatusForbidden, NoPermissions)
		//	return
		//}

		request := models.UpdatePostRequest{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, BadRequestBody)
			return
		}

		postID := mux.Vars(r)["id"]
		logInfo.Printf("Got new post update request. Post ID: %s", postID)

		if !IsPostIDValid(postID) {
			logError.Printf("Can't update post: invalid post ID. Post ID: %s", postID)
			RespondWithError(w, http.StatusBadRequest, InvalidRequest)
			return
		}

		validatePostError := validateUpdatePostRequest(&request)
		if validatePostError != nil {
			logInfo.Printf("Can't update post: invalid request. Post ID: %s. Error: %s", postID, validatePostError)
			RespondWithError(w, http.StatusBadRequest, validatePostError)
			return
		}

		isPostExists, err := postService.ExistsByID(api.db, postID)
		if err != nil {
			logError.Printf("Can't update post: error checking post for presence. Post ID: %s. Error: %s",
				postID, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}
		if !isPostExists {
			logInfo.Printf("Can't update post: post does not exist. Post ID: %s", postID)
			RespondWithError(w, http.StatusBadRequest, NoSuchPost)
			return
		}

		updateRequest := &postService.UpdateRequest{
			ID:       postID,
			Title:    request.Title,
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
func (api *PostAPIHandler) DeletePostHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//userRole := r.Context().Value(CtxRoleKey).(models.UserRole)
		//if userRole != roleAdmin {
		//	RespondWithError(w, http.StatusForbidden, NoPermissions)
		//	return
		//}

		postID := mux.Vars(r)["id"]
		logInfo.Printf("Got new post deletion request. Post ID: %s", postID)

		if !IsPostIDValid(postID) {
			logError.Printf("Can't delete post: invalid post ID. Post ID: %s", postID)
			RespondWithError(w, http.StatusBadRequest, InvalidRequest)
			return
		}

		if err := postService.Delete(api.db, postID); err != nil {
			logError.Printf("Error deleting post. Post ID: %s. Error: %s", postID, err)
		}

		logInfo.Printf("Post deleted. Post ID: %s", postID)
		Respond(w, http.StatusOK)
	})
}

// unused
// GetCertainPostHandler - this handler serves GET request for single post
func (api *PostAPIHandler) GetCertainPostHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := mux.Vars(r)["id"]
		logInfo.Printf("Got single post retrieve request. Post ID: %s", postID)

		if !IsPostIDValid(postID) {
			logError.Printf("Can't retrieve post: invalid post ID. Post ID: %s", postID)
			RespondWithError(w, http.StatusBadRequest, InvalidRequest)
			return
		}

		post, err := postService.GetByID(api.db, postID)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				logError.Printf("Can't retrieve post: no such post. Post ID: %s", postID)
				RespondWithError(w, http.StatusNotFound, NoSuchPost)
				return
			default:
				logError.Printf("Error retrieving post from database. Post ID: %s. Error: %s", postID, err)
				RespondWithError(w, http.StatusInternalServerError, TechnicalError)
				return
			}
		}

		RespondWithBody(w, http.StatusOK, post)
	})
}

// unused
// GetPostsHandler - this handler serves GET request for all posts in the given range
func (api *PostAPIHandler) GetPostsHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeParams := &GetPostsRequestQueryParams{
			Page:         r.FormValue("page"),
			PostsPerPage: r.FormValue("posts-per-page"),
		}

		logInfo.Printf("Got range of posts retrieve request. Range params: %+v", rangeParams)

		validateQueryParamsError := ValidateGetPostsRequestQueryParams(rangeParams)
		if validateQueryParamsError != nil {
			logError.Printf("Can't retrieve range of posts: invalid query params. Error: %s", validateQueryParamsError)
			RespondWithError(w, http.StatusBadRequest, validateQueryParamsError)
			return
		}

		// set default values if params are missed
		pageAsString := rangeParams.Page
		postsPerPageAsString := rangeParams.PostsPerPage

		if pageAsString == "" {
			pageAsString = DefaultPage
		}
		if postsPerPageAsString == "" {
			postsPerPageAsString = DefaultPostsPerPage
		}

		// we know that params are valid so ignore the errors
		pageAsInt, _ := strconv.Atoi(pageAsString)
		postsPerPageAsInt, _ := strconv.Atoi(postsPerPageAsString)

		posts, err := postService.GetPostsInRange(api.db, pageAsInt, postsPerPageAsInt)
		if err != nil {
			logError.Printf("Error retrieving range of posts from database: %s", err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		RespondWithBody(w, http.StatusOK, posts)
	})
}

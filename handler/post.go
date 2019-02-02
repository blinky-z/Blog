package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/postService"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

// GetPostsRangeParams - structure for storing query params of get posts request
type GetPostsRangeParams struct {
	Page         int
	PostsPerPage int
}

const (
	// InvalidTitle - incorrect user input - invalid title of post
	InvalidTitle PostErrorCode = "INVALID_TITLE"
	// InvalidID - incorrect user input - invalid id of post
	InvalidID PostErrorCode = "INVALID_ID"
	// InvalidContent - incorrect user input - invalid content of post
	InvalidContent PostErrorCode = "INVALID_CONTENT"
	// BadRequestBody - incorrect user post - invalid json post
	BadRequestBody PostErrorCode = "BAD_BODY"
	// NoSuchPost - incorrect user input - requested post does not exist in database
	NoSuchPost PostErrorCode = "NO_SUCH_POST"
	// InvalidRange - user inputs invalid range of posts to get from database
	InvalidRange PostErrorCode = "INVALID_POSTS_RANGE"
	// NoPermissions - user doesn't permissions to create/update/delete post
	NoPermissions PostErrorCode = "NO_PERMISSIONS"

	// MaxPostTitleLen - maximum length of post title
	MaxPostTitleLen int = 120

	// MaxPostsPerPage - maximum posts can be displayed on one page
	MaxPostsPerPage int = 40

	defaultPostsPerPage string = "10"

	roleAdmin = "admin"
	roleUser  = "user"
)

// ValidateGetPostsParams - validate query params of get posts request
func ValidateGetPostsParams(r *http.Request) (params GetPostsRangeParams, validateError PostErrorCode) {
	validateError = NoError

	var page int
	var postsPerPage int
	var err error

	pageAsString := r.FormValue("page")
	if len(pageAsString) == 0 {
		pageAsString = "0"
	}

	if page, err = strconv.Atoi(pageAsString); err != nil || page < 0 {
		validateError = InvalidRange
		return
	}

	postsPerPageAsString := r.FormValue("posts-per-page")
	if len(postsPerPageAsString) == 0 {
		postsPerPageAsString = defaultPostsPerPage
	}

	postsPerPage, err = strconv.Atoi(postsPerPageAsString)
	if err != nil {
		validateError = InvalidRange
		return
	}

	if postsPerPage < 0 || postsPerPage > MaxPostsPerPage {
		validateError = InvalidRange
		return
	}

	params.Page = page
	params.PostsPerPage = postsPerPage

	return
}

func validatePost(r *http.Request) (post models.Post, validateError PostErrorCode) {
	validateError = NoError
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		validateError = BadRequestBody
		return
	}

	if len(post.Title) > MaxPostTitleLen || len(post.Title) == 0 {
		validateError = InvalidTitle
		return
	}

	if len(post.Content) == 0 {
		validateError = InvalidContent
		return
	}

	return
}

// ValidatePostID - validates post id on posts/{id} endpoint
func ValidatePostID(r *http.Request) (id string, validateError PostErrorCode) {
	validateError = NoError
	vars := mux.Vars(r)

	if _, err := strconv.Atoi(vars["id"]); err != nil {
		validateError = InvalidID
		return
	}

	id = vars["id"]
	return
}

// CreatePost - create post http handler
func CreatePost(env *models.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.LogInfo.Print("Got new Post CREATE job")

		userRole := r.Context().Value(CtxKey).(string)
		if userRole != roleAdmin {
			env.LogInfo.Printf("User with role %s doesn't have permissions to CREATE post", userRole)
			RespondWithError(w, http.StatusForbidden, NoPermissions, env.LogError)
			return
		}

		post, validatePostError := validatePost(r)
		if validatePostError != NoError {
			env.LogInfo.Print("Can not create post: post is invalid")
			RespondWithError(w, http.StatusBadRequest, validatePostError, env.LogError)
			return
		}

		var createdPost models.Post

		env.LogInfo.Printf("Inserting post with Title %s into database", post.Title)

		if _, err := env.Db.Exec("BEGIN TRANSACTION"); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}
		if err := env.Db.QueryRow("insert into posts(title, content) values($1, $2) RETURNING id, title, date, content",
			post.Title, post.Content).
			Scan(&createdPost.ID, &createdPost.Title, &createdPost.Date, &createdPost.Content); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}
		if _, err := env.Db.Exec("END TRANSACTION"); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("Post with Title %s successfully created", createdPost.Title)

		RespondWithBody(w, http.StatusCreated, createdPost, env.LogError)
	})
}

// UpdatePost - update post http handler
func UpdatePost(env *models.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.LogInfo.Print("Got new Post UPDATE job")

		userRole := r.Context().Value(CtxKey).(string)
		if userRole != roleAdmin {
			env.LogInfo.Printf("User with role %s doesn't have permissions to UPDATE post", userRole)
			RespondWithError(w, http.StatusForbidden, NoPermissions, env.LogError)
			return
		}

		id, validateIDError := ValidatePostID(r)
		if validateIDError != NoError {
			env.LogInfo.Print("Can not UPDATE post: ID of Post to update is invalid")
			RespondWithError(w, http.StatusBadRequest, validateIDError, env.LogError)
			return
		}

		post, validatePostError := validatePost(r)
		if validatePostError != NoError {
			env.LogInfo.Printf("Can not UPDATE post with ID %s : New Post is invalid", id)
			RespondWithError(w, http.StatusBadRequest, validatePostError, env.LogError)
			return
		}

		if err := env.Db.QueryRow("select from posts where id = $1", id).Scan(); err != nil {
			if err == sql.ErrNoRows {
				env.LogInfo.Printf("Can not UPDATE post with ID %s : post does not exist", id)
				RespondWithError(w, http.StatusNotFound, NoSuchPost, env.LogError)
				return
			}

			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		var updatedPost models.Post

		env.LogInfo.Printf("Updating post with ID %s in database", id)

		if _, err := env.Db.Exec("BEGIN TRANSACTION"); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}
		if err := env.Db.QueryRow("UPDATE posts SET title = $1, content = $2 WHERE id = $3 RETURNING id, title, date, content",
			post.Title, post.Content, id).
			Scan(&updatedPost.ID, &updatedPost.Title, &updatedPost.Date, &updatedPost.Content); err != nil {
			if err != nil {
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

		env.LogInfo.Printf("Post with ID %s successfully updated", id)

		RespondWithBody(w, http.StatusOK, updatedPost, env.LogError)
	})
}

// DeletePost - delete post http handler
func DeletePost(env *models.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.LogInfo.Print("Got new Post DELETE job")

		userRole := r.Context().Value(CtxKey).(string)
		if userRole != roleAdmin {
			env.LogInfo.Printf("User with role %s doesn't have permissions to DELETE post", userRole)
			RespondWithError(w, http.StatusForbidden, NoPermissions, env.LogError)
			return
		}

		id, validateIDError := ValidatePostID(r)
		if validateIDError != NoError {
			env.LogInfo.Print("Can not DELETE post: post ID is invalid")
			RespondWithError(w, http.StatusBadRequest, validateIDError, env.LogError)
			return
		}

		env.LogInfo.Printf("Deleting post with ID %s from database", id)

		if _, err := env.Db.Exec("BEGIN TRANSACTION"); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}
		if _, err := env.Db.Exec("DELETE FROM posts WHERE id = $1", id); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}
		if _, err := env.Db.Exec("END TRANSACTION"); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("Post with ID %s successfully deleted", id)

		respond(w, http.StatusOK)
	})
}

// GetCertainPost - get single post from database http handler
func GetCertainPost(env *models.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, validateIDError := ValidatePostID(r)
		if validateIDError != NoError {
			env.LogInfo.Print("Can not GET post: post ID is invalid")
			RespondWithError(w, http.StatusBadRequest, validateIDError, env.LogError)
			return
		}

		post, err := postService.GetCertainPost(env, id)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				RespondWithError(w, http.StatusNotFound, NoSuchPost, env.LogError)
			default:
				RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			}
		}

		RespondWithBody(w, http.StatusOK, post, env.LogError)
	})
}

// GetPosts - get one page of posts from database http handler
func GetPosts(env *models.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params, validateError := ValidateGetPostsParams(r)
		if validateError != NoError {
			env.LogInfo.Print("Can not GET range of posts : get posts Query params are invalid")
			RespondWithError(w, http.StatusBadRequest, validateError, env.LogError)
			return
		}

		page := params.Page
		postsPerPage := params.PostsPerPage

		posts, err := postService.GetPosts(env, page, postsPerPage)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, TechnicalError, env.LogError)
		}

		RespondWithBody(w, http.StatusOK, posts, env.LogError)
	})
}

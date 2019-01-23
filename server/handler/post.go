package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/blinky-z/Blog/server/models"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type getPostsRangeParams struct {
	page         int
	postsPerPage int
}

const (
	// TechnicalError - server error
	TechnicalError PostErrorCode = "TECHNICAL_ERROR"
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
	// NoError - no error occurred while handling request
	NoError PostErrorCode = ""

	// MaxPostTitleLen - maximum length of post title
	MaxPostTitleLen int = 120

	// MaxPostsPerPage - maximum posts can be displayed on one page
	MaxPostsPerPage int = 40

	defaultMaxPostsPerPage string = "10"
)

func validateGetPostsParams(r *http.Request) (params getPostsRangeParams, validateError PostErrorCode) {
	validateError = NoError

	var page int
	var postsPerPage int
	var err error

	if len(r.FormValue("page")) == 0 {
		validateError = InvalidRange
		return
	}

	if page, err = strconv.Atoi(r.FormValue("page")); err != nil || page < 0 {
		validateError = InvalidRange
		return
	}

	postsPerPageAsString := r.FormValue("posts-per-page")
	if len(postsPerPageAsString) == 0 {
		postsPerPageAsString = defaultMaxPostsPerPage
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

	params.page = page
	params.postsPerPage = postsPerPage

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

func validatePostID(r *http.Request) (id string, validateError PostErrorCode) {
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
func CreatePost(w http.ResponseWriter, r *http.Request) {
	post, validatePostError := validatePost(r)
	if validatePostError != NoError {
		respondWithError(w, http.StatusBadRequest, validatePostError)
		return
	}

	LogInfo.Printf("Got new post creation job. New post: %v", post)

	var createdPost models.Post

	if _, err := Db.Exec("BEGIN TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if err := Db.QueryRow("insert into posts(title, content) values($1, $2) RETURNING id, title, date, content",
		post.Title, post.Content).
		Scan(&createdPost.ID, &createdPost.Title, &createdPost.Date, &createdPost.Content); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if _, err := Db.Exec("END TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Created new post: %v", createdPost)

	respondWithBody(w, http.StatusCreated, createdPost)
}

// UpdatePost - update post http handler
func UpdatePost(w http.ResponseWriter, r *http.Request) {
	id, validateIDError := validatePostID(r)
	if validateIDError != NoError {
		respondWithError(w, http.StatusBadRequest, validateIDError)
		return
	}

	post, validatePostError := validatePost(r)
	if validatePostError != NoError {
		respondWithError(w, http.StatusBadRequest, validatePostError)
		return
	}

	LogInfo.Printf("Got new post update job. New post: %v", post)

	if err := Db.QueryRow("select from posts where id = $1", id).Scan(); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, NoSuchPost)
			return
		}

		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	var updatedPost models.Post

	if _, err := Db.Exec("BEGIN TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if err := Db.QueryRow("UPDATE posts SET title = $1, content = $2 WHERE id = $3 RETURNING id, title, date, content",
		post.Title, post.Content, id).
		Scan(&updatedPost.ID, &updatedPost.Title, &updatedPost.Date, &updatedPost.Content); err != nil {
		if err != nil {
			LogError.Print(err)
			respondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}
	}
	if _, err := Db.Exec("END TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Updated post: %v", updatedPost)

	respondWithBody(w, http.StatusOK, updatedPost)
}

// DeletePost - delete post http handler
func DeletePost(w http.ResponseWriter, r *http.Request) {
	id, validateIDError := validatePostID(r)
	if validateIDError != NoError {
		respondWithError(w, http.StatusBadRequest, validateIDError)
		return
	}

	LogInfo.Printf("Got new post deletion job. Post id: %s", id)

	if _, err := Db.Exec("BEGIN TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if _, err := Db.Exec("DELETE FROM posts WHERE id = $1", id); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if _, err := Db.Exec("END TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Post with id %s deleted", id)

	respond(w, http.StatusOK)
}

// GetCertainPost - get single post from database http handler
func GetCertainPost(w http.ResponseWriter, r *http.Request) {
	var post models.Post

	id, validateIDError := validatePostID(r)
	if validateIDError != NoError {
		respondWithError(w, http.StatusBadRequest, validateIDError)
		return
	}

	LogInfo.Printf("Got new get certain post job. Post id: %s", id)

	if err := Db.QueryRow("select id, title, date, content from posts where id = $1", id).
		Scan(&post.ID, &post.Title, &post.Date, &post.Content); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, NoSuchPost)
			return
		}

		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Post with id %s arrived from database", id)

	respondWithBody(w, 200, post)
}

// GetPosts - get one page of posts from database http handler
func GetPosts(w http.ResponseWriter, r *http.Request) {
	params, validateError := validateGetPostsParams(r)
	if validateError != NoError {
		respondWithError(w, http.StatusBadRequest, validateError)
		return
	}

	page := params.page
	postsPerPage := params.postsPerPage

	var posts []models.Post

	LogInfo.Printf("Got new get range of posts job. Page: %d. Posts per page: %d", page, postsPerPage)

	rows, err := Db.Query("select id, title, date, content from posts order by id DESC offset $1 limit $2",
		page*postsPerPage, postsPerPage)
	if err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	for rows.Next() {
		var currentPost models.Post
		if err = rows.Scan(&currentPost.ID, &currentPost.Title, &currentPost.Date, &currentPost.Content); err != nil {
			LogError.Print(err)
			respondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}
		posts = append(posts, currentPost)
	}

	LogInfo.Print("Posts arrived from database")

	respondWithBody(w, 200, posts)
}

package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/blinky-z/Blog/server/models"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

var (
	// LogInfo - log for writing+ messages
	LogInfo *log.Logger
	// LogError - log for writing server errors
	LogError *log.Logger

	// Db - database connection. This variable is set by main function
	Db *sql.DB
)

// Response - behaves like Either Monad
// 'Error' field is set while error occurred.
// Otherwise 'Body' field is used to return post from database
type Response struct {
	Error PostErrorCode `json:"error"`
	Body  interface{}   `json:"body"`
}

// PostErrorCode - represents error occurred while handling request
type PostErrorCode string

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

	maxPostTitleLen int = 120
)

func respondWithBody(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_, err := w.Write(response)
	if err != nil {
		LogError.Print(err)
	}
}

func validateUserGetPostsParams(r *http.Request) (params map[string]int, validateError PostErrorCode) {
	params = make(map[string]int)
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

	if len(r.FormValue("posts-per-page")) == 0 {
		postsPerPage = 10
	} else {
		if postsPerPage, err = strconv.Atoi(r.FormValue("posts-per-page")); err != nil || postsPerPage < 0 ||
			postsPerPage > 40 {
			validateError = InvalidRange
			return
		}
	}

	params["page"] = page
	params["posts-per-page"] = postsPerPage

	return
}

func validateUserPost(r *http.Request) (post models.Post, validateError PostErrorCode) {
	validateError = NoError
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		validateError = BadRequestBody
		return
	}

	if len(post.Title) > maxPostTitleLen || len(post.Title) == 0 {
		validateError = InvalidTitle
		return
	}

	if len(post.Content) == 0 {
		validateError = InvalidContent
		return
	}

	return
}

func validateUserID(r *http.Request) (id string, validateError PostErrorCode) {
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
	var response Response

	post, validatePostError := validateUserPost(r)
	if validatePostError != NoError {
		response.Error = validatePostError
		respondWithBody(w, http.StatusBadRequest, response)
		return
	}

	LogInfo.Printf("Got new post creation job. New post: %v", post)

	var createdPost models.Post

	_, _ = Db.Exec("BEGIN TRANSACTION")
	err := Db.QueryRow("insert into posts(title, content) values($1, $2) RETURNING id, title, date, content",
		post.Title, post.Content).Scan(&createdPost.ID, &createdPost.Title, &createdPost.Date, &createdPost.Content)
	_, _ = Db.Exec("END TRANSACTION")
	if err != nil {
		response.Error = TechnicalError
		LogError.Print(err)
		respondWithBody(w, http.StatusInternalServerError, response)
		return
	}

	LogInfo.Printf("Created new post: %v", createdPost)

	response.Body = createdPost
	respondWithBody(w, http.StatusCreated, response)
}

// UpdatePost - update post http handler
func UpdatePost(w http.ResponseWriter, r *http.Request) {
	var response Response

	id, validateIDError := validateUserID(r)
	if validateIDError != NoError {
		response.Error = validateIDError
		respondWithBody(w, http.StatusBadRequest, response)
		return
	}

	post, validatePostError := validateUserPost(r)
	if validatePostError != NoError {
		response.Error = validatePostError
		respondWithBody(w, http.StatusBadRequest, response)
		return
	}

	LogInfo.Printf("Got new post update job. New post: %v", post)

	err := Db.QueryRow("select from posts where id = $1", id).Scan()
	if err != nil {
		if err == sql.ErrNoRows {
			response.Error = NoSuchPost
			respondWithBody(w, http.StatusNotFound, response)
			return
		}

		response.Error = TechnicalError
		LogError.Print(err)
		respondWithBody(w, http.StatusInternalServerError, response)
		return
	}

	var updatedPost models.Post

	_, _ = Db.Exec("BEGIN TRANSACTION")
	err = Db.QueryRow("UPDATE posts SET title = $1, content = $2 WHERE id = $3 RETURNING id, title, date, content",
		post.Title, post.Content, id).Scan(&updatedPost.ID, &updatedPost.Title, &updatedPost.Date, &updatedPost.Content)
	_, _ = Db.Exec("END TRANSACTION")
	if err != nil {
		response.Error = TechnicalError
		LogError.Print(err)
		respondWithBody(w, http.StatusInternalServerError, response)
		return
	}

	LogInfo.Printf("Updated post: %v", updatedPost)

	response.Body = updatedPost
	respondWithBody(w, http.StatusOK, response)
}

// DeletePost - delete post http handler
func DeletePost(w http.ResponseWriter, r *http.Request) {
	var response Response

	id, validateIDError := validateUserID(r)
	if validateIDError != NoError {
		response.Error = validateIDError
		respondWithBody(w, http.StatusBadRequest, response)
		return
	}

	LogInfo.Printf("Got new post deletion job. Post id: %s", id)

	_, _ = Db.Exec("BEGIN TRANSACTION")
	_, err := Db.Exec("DELETE FROM posts WHERE id = $1", id)
	_, _ = Db.Exec("END TRANSACTION")
	if err != nil {
		response.Error = TechnicalError
		LogError.Print(err)
		respondWithBody(w, http.StatusInternalServerError, response)
		return
	}

	LogInfo.Printf("Post with id %s deleted", id)

	respondWithBody(w, http.StatusOK, response)
}

// GetCertainPost - get single post from database http handler
func GetCertainPost(w http.ResponseWriter, r *http.Request) {
	var response Response

	var post models.Post

	id, validateIDError := validateUserID(r)
	if validateIDError != NoError {
		response.Error = validateIDError
		respondWithBody(w, http.StatusBadRequest, response)
		return
	}

	LogInfo.Printf("Got new get certain post job. Post id: %s", id)

	err := Db.QueryRow("select id, title, date, content from posts where id = $1", id).Scan(
		&post.ID, &post.Title, &post.Date, &post.Content)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Error = NoSuchPost
			respondWithBody(w, http.StatusNotFound, response)
			return
		}

		response.Error = TechnicalError
		LogError.Print(err)
		respondWithBody(w, http.StatusInternalServerError, response)
		return
	}

	LogInfo.Printf("Post with id %s arrived from database", id)

	response.Body = post
	respondWithBody(w, 200, response)
}

// GetPosts - get one page of posts from database http handler
func GetPosts(w http.ResponseWriter, r *http.Request) {
	var response Response

	params, validateError := validateUserGetPostsParams(r)
	if validateError != NoError {
		response.Error = validateError
		respondWithBody(w, http.StatusBadRequest, response)
		return
	}

	page := params["page"]
	postsPerPage := params["posts-per-page"]

	var posts []models.Post

	LogInfo.Printf("Got new get range of posts job. Page: %d. Posts per page: %d", page, postsPerPage)

	rows, err := Db.Query("select id, title, date, content from posts order by id DESC offset $1 limit $2",
		page*postsPerPage, postsPerPage)
	if err != nil {
		response.Error = TechnicalError
		LogError.Print(err)
		respondWithBody(w, http.StatusInternalServerError, response)
		return
	}

	for rows.Next() {
		var currentPost models.Post
		err = rows.Scan(&currentPost.ID, &currentPost.Title, &currentPost.Date, &currentPost.Content)
		if err != nil {
			response.Error = TechnicalError
			LogError.Print(err)
			respondWithBody(w, http.StatusInternalServerError, response)
			return
		}
		posts = append(posts, currentPost)
	}

	LogInfo.Print("Posts arrived from database")

	response.Body = posts
	respondWithBody(w, 200, response)
}

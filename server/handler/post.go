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

// represents error occurred while handling request
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

	err := Db.QueryRow("insert into posts(title, content) values($1, $2) RETURNING id, title, date, content;",
		post.Title, post.Content).Scan(&createdPost.ID, &createdPost.Title, &createdPost.Date, &createdPost.Content)
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
	err = Db.QueryRow("UPDATE posts SET title = $1, content = $2 WHERE id = $3 RETURNING id, title, date, content;",
		post.Title, post.Content, id).Scan(&updatedPost.ID, &updatedPost.Title, &updatedPost.Date, &updatedPost.Content)
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

	_, err := Db.Exec("DELETE FROM posts WHERE id = $1;", id)
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

	LogInfo.Printf("Got new post get job. Post id: %s", id)

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

// GetPosts - get all posts from database http handler
func GetPosts(w http.ResponseWriter, r *http.Request) {
	var response Response

	var posts []models.Post

	rows, err := Db.Query("select id, title, date, content from posts")
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

	response.Body = posts
	respondWithBody(w, 200, response)
}

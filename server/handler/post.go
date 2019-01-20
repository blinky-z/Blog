package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/Blog/server/models"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var (
	LogInfo  *log.Logger
	LogError *log.Logger

	Db *sql.DB
)

type Response struct {
	Error postErrorCode `json:"error"`
	Body  interface{}   `json:"body"`
}

type postErrorCode string

const (
	TechnicalError postErrorCode = "TECHNICAL_ERROR"
	InvalidTitle   postErrorCode = "INVALID_TITLE"
	BadPostBody    postErrorCode = "INVALID_BODY"
	NoSuchPost     postErrorCode = "NO_SUCH_POST"

	maxPostTitleLen int = 120
)

func checkError(err error) {
	if err != nil {
		LogError.Print(err)
	}
}

func respondWithBody(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	checkError(err)
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	var response Response

	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		response.Error = BadPostBody
		respondWithBody(w, http.StatusBadRequest, response)
		return
	}

	if len(post.Title) > maxPostTitleLen {
		response.Error = InvalidTitle
		respondWithBody(w, http.StatusBadRequest, response)
		return
	}

	LogInfo.Printf("Got new post creation job. New post: %v", post)

	var createdPost models.Post

	err = Db.QueryRow("insert into posts(title, content) values($1, $2) RETURNING *;",
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

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	var response Response

	vars := mux.Vars(r)
	id := vars["id"]

	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		response.Error = BadPostBody
		respondWithBody(w, http.StatusBadRequest, response)
		return
	}

	if len(post.Title) > maxPostTitleLen {
		response.Error = InvalidTitle
		respondWithBody(w, http.StatusBadRequest, response)
		return
	}

	LogInfo.Printf("Got new post update job. New post: %v", post)

	err = Db.QueryRow("select from posts where id = $1", id).Scan()
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
	err = Db.QueryRow("UPDATE posts SET title = $1, content = $2 WHERE id = $3 RETURNING *;",
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

func DeletePost(w http.ResponseWriter, r *http.Request) {
	var response Response

	vars := mux.Vars(r)
	id := vars["id"]

	LogInfo.Printf("Got new post deletion job. Post id: %s", id)

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

	_, err = Db.Exec("DELETE FROM posts WHERE id = $1;", id)
	if err != nil {
		response.Error = TechnicalError
		LogError.Print(err)
		respondWithBody(w, http.StatusInternalServerError, response)
		return
	}

	LogInfo.Printf("Post with id %s deleted", id)

	respondWithBody(w, http.StatusOK, response)
}

func GetCertainPost(w http.ResponseWriter, r *http.Request) {
	var response Response

	var post models.Post

	vars := mux.Vars(r)
	id := vars["id"]

	err := Db.QueryRow("select * from posts where id = $1", id).Scan(
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

	response.Body = post
	respondWithBody(w, 200, response)
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	var response Response

	var posts []models.Post

	rows, err := Db.Query("select * from posts")
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

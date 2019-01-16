package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/Blog/server/models"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

var (
	LogInfo  *log.Logger
	LogError *log.Logger

	Db *sql.DB
)

func checkError(err error) {
	if err != nil {
		LogError.Print(err)
	}
}

func respond(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	_, err := w.Write([]byte{})
	checkError(err)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	checkError(err)
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Bad post body")
		return
	}

	LogInfo.Print("Got new post creation job\nPost:")
	LogInfo.Print(post)

	_, err = Db.Query("insert into posts(title, content) values($1, $2);",
		post.Title, post.Content)

	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, err)
		return
	}
	respond(w, http.StatusCreated)
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	var posts []models.Post

	rows, err := Db.Query("select * from posts")
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, err)
		return
	}

	for rows.Next() {
		var currentPost models.Post
		err = rows.Scan(&currentPost.ID, &currentPost.Title, &currentPost.Date, &currentPost.Content)
		if err != nil {
			respondWithJSON(w, http.StatusInternalServerError, err)
			return
		}
		posts = append(posts, currentPost)
	}

	respondWithJSON(w, 200, posts)
}

func GetCertainPost(w http.ResponseWriter, r *http.Request) {
	var post models.Post

	vars := mux.Vars(r)
	id := vars["id"]

	row := Db.QueryRow("select * from posts where id = $1", id)

	err := row.Scan(&post.ID, &post.Title, &post.Date, &post.Content)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, 200, post)
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	var post models.Post

	vars := mux.Vars(r)
	id := vars["id"]

	row := Db.QueryRow("select * from posts where id = $1", id)

	err := row.Scan(&post.ID, &post.Title, &post.Date, &post.Content)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, 200, post)
}

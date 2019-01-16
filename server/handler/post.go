package handler

import (
	. "../models"
	"database/sql"
	"encoding/json"
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
	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, "Bad post body")
		return
	}

	LogInfo.Print("Got new post creation job\nPost:")
	LogInfo.Print(post)

	_, err = Db.Query("insert into posts(id, title, date, content) values(DEFAULT, $1, DEFAULT, $2);",
		post.Title, post.Content)

	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, err)
	} else {
		respond(w, http.StatusCreated)
	}
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	var posts []Post

	rows, err := Db.Query("select * from posts")
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, err)
		return
	}

	for rows.Next() {
		var currentPost Post
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
	var post Post

	vars := mux.Vars(r)
	id := vars["id"]

	row := Db.QueryRow("select * from posts where id = $1", id)

	err := row.Scan(&post.ID, &post.Title, &post.Date, &post.Content)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, err)
	} else {
		respondWithJSON(w, 200, post)
	}
}

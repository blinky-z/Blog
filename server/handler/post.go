package handler

import (
	. "../models"
	"database/sql"
	"encoding/json"
	"fmt"
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

func respondwithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	fmt.Println(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	checkError(err)
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	var posts []Post

	rows, err := Db.Query("select * from posts")
	defer rows.Close()
	checkError(err)
	rows.Scan(posts)

	respondwithJSON(w, 200, posts)
}

func GetCertainPost(w http.ResponseWriter, r *http.Request) {
	var post Post

	vars := mux.Vars(r)
	id := vars["id"]

	row, _ := Db.Query("select * from posts WHERE id = ?", id)
	defer row.Close()
	err := row.Scan(&post)
	checkError(err)

	respondwithJSON(w, 200, post)
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	checkError(err)
	LogInfo.Print("Got new post creation job\nPost:")
	LogInfo.Print(post)

	_, err = Db.Query("insert into posts(id, title, date, content) values(DEFAULT, $1, DEFAULT, $2);",
		post.Title, post.Content)
	checkError(err)

	respond(w, http.StatusCreated)
}

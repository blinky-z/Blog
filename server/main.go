package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"time"
)

var (
	db *sql.DB
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "postgres"
	DB_NAME     = "testdb"
)

type Post struct {
	ID      string    `json:"id"`
	Title   string    `json:"name"`
	Date    time.Time `json:"time"`
	Content string    `json:"content"`
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
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

func getPosts(w http.ResponseWriter, r *http.Request) {
	var posts []Post

	rows, err := db.Query("select * from posts")
	defer rows.Close()
	checkError(err)
	rows.Scan(posts)

	respondwithJSON(w, 200, posts)
}

func getCertainPost(w http.ResponseWriter, r *http.Request) {
	var post Post
	var id string
	json.NewDecoder(r.Body).Decode(&id)

	row, err := db.Query("select * from posts WHERE id = ?", id)
	defer row.Close()
	checkError(err)
	row.Scan(&post)

	respondwithJSON(w, 200, post)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	var post Post
	json.NewDecoder(r.Body).Decode(&post)

	_, err := db.Exec("insert into posts(id, title, date, content) values(DEFAULT, $1, $2, $3);",
		post.Title, post.Date, post.Content)
	checkError(err)

	respond(w, http.StatusCreated)
}

func main() {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkError(err)
	defer db.Close()

	router := mux.NewRouter()

	server := &http.Server{ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      router,
		Addr:         "localhost:8080"}

	router.HandleFunc("/posts", getPosts).Methods("GET")
	router.HandleFunc("/posts", createPost).Methods("POST")
	router.HandleFunc("/posts/{id}", getCertainPost).Methods("GET")

	fmt.Printf("listening on address %s", server.Addr)
	log.Fatal(http.ListenAndServe(":8000", router))
}

package main

import (
	"database/sql"
	"fmt"
	"github.com/blinky-z/Blog/server/handler"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
)

var (
	logInfoOutfile, _  = os.OpenFile("./logs/Info.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	logErrorOutfile, _ = os.OpenFile("./logs/Error.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

	logInfo  = log.New(logInfoOutfile, "INFO: ", log.Ltime)
	logError = log.New(logErrorOutfile, "ERROR: ", log.Ltime)

	Port    = "8080"
	Address = "localhost:" + Port

	user     string
	password string
	dbName   string

	Db *sql.DB
)

func RunServer(configName, configPath string) {
	viper.SetConfigName(configName)
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		logError.Fatalf("Fatal error config file: %s \n", err)
	}

	user = viper.GetString("user")
	password = viper.GetString("password")
	dbName = viper.GetString("db_name")

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s",
		user, password, dbName)

	logInfo.Printf("Logging into postgres database with following credentials: %s", dbinfo)

	Db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logError.Fatal(err)
	}
	defer Db.Close()
	handler.Db = Db
	handler.LogInfo = logInfo
	handler.LogError = logError

	router := mux.NewRouter()

	router.HandleFunc("/posts", handler.GetPosts).Methods("GET")
	router.HandleFunc("/posts", handler.CreatePost).Methods("POST")
	router.HandleFunc("/posts/{id}", handler.GetCertainPost).Methods("GET")
	router.HandleFunc("/posts/{id}", handler.UpdatePost).Methods("PUT")
	router.HandleFunc("/posts/{id}", handler.DeletePost).Methods("DELETE")
	router.HandleFunc("/hc", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}).Methods("GET")

	Db.Prepare("SELECT id FROM posts WHERE id = $1")

	logInfo.Printf("listening on address %s", Address)
	logError.Fatal(http.ListenAndServe(Address, router))
}

func main() {
	RunServer(os.Args[1], os.Args[2])
}

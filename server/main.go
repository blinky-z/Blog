package main

import (
	"./handler"
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
)

var (
	logInfoOutfile, _  = os.OpenFile("./logs/Info.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	logErrorOutfile, _ = os.OpenFile("./logs/Error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	logInfo  = log.New(logInfoOutfile, "INFO: ", log.Ltime)
	logError = log.New(logErrorOutfile, "ERROR: ", log.Ltime)

	port    = "8080"
	address = "localhost:" + port

	DbUser     string
	DbPassword string
	DbName     string
)

func main() {
	viper.SetConfigName(os.Args[1])
	viper.AddConfigPath(os.Args[2])
	err := viper.ReadInConfig()
	if err != nil {
		logError.Fatalf("Fatal error config file: %s \n", err)
	}

	DbUser = viper.GetString("DB_USER")
	DbPassword = viper.GetString("DB_PASSWORD")
	DbName = viper.GetString("DB_NAME")

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DbUser, DbPassword, DbName)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logError.Fatal(err)
	}
	defer db.Close()
	handler.Db = db
	handler.LogInfo = logInfo
	handler.LogError = logError

	router := mux.NewRouter()

	router.HandleFunc("/posts", handler.GetPosts).Methods("GET")
	router.HandleFunc("/posts", handler.CreatePost).Methods("POST")
	router.HandleFunc("/posts/{id}", handler.GetCertainPost).Methods("GET")

	logInfo.Printf("listening on address %s", address)
	logInfo.Fatal(http.ListenAndServe(address, router))
}

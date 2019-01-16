package main

import (
	"database/sql"
	"fmt"
	"github.com/Blog/server/handler"
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

	Db *sql.DB
)

func RunServer(configName, configPath string) {
	viper.SetConfigName(configName)
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		logError.Fatalf("Fatal error config file: %s \n", err)
	}

	DbUser = viper.GetString("db_user")
	DbPassword = viper.GetString("db_password")
	DbName = viper.GetString("db_name")

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s",
		DbUser, DbPassword, DbName)

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

	logInfo.Printf("listening on address %s", address)
	logInfo.Fatal(http.ListenAndServe(address, router))
}

func main() {
	RunServer(os.Args[1], os.Args[2])
}

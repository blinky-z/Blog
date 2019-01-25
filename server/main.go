package main

import (
	"database/sql"
	"fmt"
	"github.com/blinky-z/Blog/server/handler"
	"github.com/blinky-z/Blog/server/models"
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

	// Port - server Port
	Port = "8080"
	// Address - server address with port
	Address = "localhost:" + Port

	user     string
	password string
	dbName   string

	// Db - database connection
	Db *sql.DB

	signingKey []byte
)

// RunServer - run server function. Config file name and path should be passed
func RunServer(configName, configPath string) {
	viper.SetConfigName(configName)
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		logError.Fatalf("Fatal error reading config file: %s \n", err)
	}

	user = viper.GetString("user")
	password = viper.GetString("password")
	dbName = viper.GetString("db_name")
	signingKey = []byte(viper.GetString("secretKey"))

	testAdminsFile := viper.GetString("adminsConfigFile")
	viper.SetConfigName(testAdminsFile)
	err = viper.ReadInConfig()
	if err != nil {
		logError.Fatalf("Fatal error reading config file: %s \n", err)
	}

	var admins []models.User
	err = viper.UnmarshalKey("admins", &admins)
	if err != nil {
		logError.Fatalf("Fatal error unmarshaling admins: %s \n", err)
	}

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s",
		user, password, dbName)

	logInfo.Printf("Logging into postgres database with following credentials: %s", dbinfo)

	Db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logError.Fatal(err)
	}
	defer func() {
		err := Db.Close()
		if err != nil {
			panic(err)
		}
	}()
	handler.Db = Db
	handler.LogInfo = logInfo
	handler.LogError = logError
	handler.SigningKey = signingKey
	handler.Admins = admins

	router := mux.NewRouter()

	router.HandleFunc("/posts", handler.GetPosts).Queries("page", "{page}",
		"posts-per-page", "{posts-per-page}").Methods("GET")
	router.HandleFunc("/posts", handler.GetPosts).Queries("page", "{page}").Methods("GET")
	router.HandleFunc("/posts/{id}", handler.GetCertainPost).Methods("GET")
	router.Handle("/posts", handler.JwtAuthentication(http.HandlerFunc(handler.CreatePost))).Methods("POST")
	router.Handle("/posts/{id}", handler.JwtAuthentication(http.HandlerFunc(handler.UpdatePost))).Methods("PUT")
	router.Handle("/posts/{id}", handler.JwtAuthentication(http.HandlerFunc(handler.DeletePost))).Methods("DELETE")
	router.HandleFunc("/hc", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}).Methods("GET")
	router.HandleFunc("/user/register", handler.RegisterUserHandler).Methods("GET")
	router.HandleFunc("/user/login", handler.LoginUserHandler).Methods("GET")

	logInfo.Printf("listening on address %s", Address)
	logError.Fatal(http.ListenAndServe(Address, router))
}

func main() {
	RunServer(os.Args[1], os.Args[2])
}

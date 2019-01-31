package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/auth0/go-jwt-middleware"
	"github.com/blinky-z/Blog/server/handler"
	"github.com/blinky-z/Blog/server/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	logInfoOutfile, _  = os.OpenFile("./logs/Info.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	logErrorOutfile, _ = os.OpenFile("./logs/Error.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

	logInfo  = log.New(logInfoOutfile, "INFO: ", log.Ltime)
	logError = log.New(logErrorOutfile, "ERROR: ", log.Ltime)

	// Port - server Port
	Port = "8080"
	// Address - server address with port
	Address = "localhost:" + Port

	frontFolder = "front/"

	dbUser     string
	dbPassword string
	dbName     string

	// Db - database connection
	Db *sql.DB

	signingKey []byte

	jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return signingKey, nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			logInfo.Printf("Error checking JWT Token: Malformed or invalid JWT Token")
			handler.RespondWithError(w, http.StatusUnauthorized, handler.InvalidToken)
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
)

var handleHTMLFile = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	currentURLPath := r.URL.Path
	currentURLPath = strings.TrimSuffix(currentURLPath, ".html")

	var fileName string
	if currentURLPath == "" {
		fileName = "index.html"
	} else {
		fileName = currentURLPath + ".html"
	}

	filePath := frontFolder + fileName
	http.ServeFile(w, r, filePath)
})

// RunServer - run server function. Config file name and path should be passed
func RunServer(serverConfigPath, adminsConfigPath string) {
	viper.SetConfigFile(serverConfigPath)
	err := viper.ReadInConfig()
	if err != nil {
		logError.Fatalf("Fatal error reading server config file: %s \n", err)
	}

	dbUser = viper.GetString("db_user")
	dbPassword = viper.GetString("db_password")
	dbName = viper.GetString("db_name")
	signingKey = []byte(viper.GetString("jwtSecretKey"))

	viper.SetConfigFile(adminsConfigPath)
	err = viper.ReadInConfig()
	if err != nil {
		logError.Fatalf("Fatal error reading admins list config file: %s", err)
	}

	var admins []models.User
	err = viper.UnmarshalKey("admins", &admins)
	if err != nil {
		logError.Fatalf("Fatal error unmarshaling admins list: %s", err)
	}

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s",
		dbUser, dbPassword, dbName)

	logInfo.Printf("Logging into postgres database with following credentials: %s", dbinfo)

	Db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logError.Fatalf("Error opening connection with database: %s", err)
	}
	defer func() {
		err := Db.Close()
		if err != nil {
			logError.Fatalf("Error closing connection with database: %s", err)
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
	router.Handle("/posts",
		jwtMiddleware.Handler(handler.JwtAuthentication(http.HandlerFunc(handler.CreatePost)))).Methods("POST")
	router.Handle("/posts/{id}",
		jwtMiddleware.Handler(handler.JwtAuthentication(http.HandlerFunc(handler.UpdatePost)))).Methods("PUT")
	router.Handle("/posts/{id}",
		jwtMiddleware.Handler(handler.JwtAuthentication(http.HandlerFunc(handler.DeletePost)))).Methods("DELETE")
	router.HandleFunc("/hc", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}).Methods("GET")
	router.HandleFunc("/user/register", handler.RegisterUserHandler).Methods("POST")
	router.HandleFunc("/user/login", handler.LoginUserHandler).Methods("POST")

	router.PathPrefix("/css").Handler(http.StripPrefix("/css", http.FileServer(http.Dir(frontFolder+"css"))))
	router.PathPrefix("/scripts").Handler(http.StripPrefix("/scripts", http.FileServer(http.Dir(frontFolder+"scripts"))))
	router.PathPrefix("/images").Handler(http.StripPrefix("/images", http.FileServer(http.Dir(frontFolder+"images"))))
	router.PathPrefix("/").Handler(http.StripPrefix("/", handleHTMLFile))

	logInfo.Printf("listening on address %s", Address)
	logError.Fatal(http.ListenAndServe(Address, router))
}

func main() {
	userConfigPath := flag.String("config", "configs/config.json", "config file path")
	adminsListConfigPath := flag.String("admins", "configs/admins.json", "admins list")

	flag.Parse()
	RunServer(*userConfigPath, *adminsListConfigPath)
}

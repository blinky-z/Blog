package main

import (
	"database/sql"
	"fmt"
	"github.com/auth0/go-jwt-middleware"
	"github.com/blinky-z/Blog/handler/api"
	"github.com/blinky-z/Blog/handler/web"
	"github.com/blinky-z/Blog/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	env *models.Env = &models.Env{}

	jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return env.SigningKey, nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			logInfo.Printf("Error checking JWT Token: Malformed or invalid JWT Token")
			api.RespondWithError(w, http.StatusUnauthorized, api.InvalidToken, env.LogError)
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
)

// RunServer - run server function. Config file name and path should be passed
func RunServer(serverConfigPath, adminsConfigPath string) {
	viper.SetConfigFile(serverConfigPath)
	err := viper.ReadInConfig()
	if err != nil {
		logError.Fatalf("Fatal error reading server config file: %s \n", err)
	}

	dbUser := viper.GetString("db_user")
	dbPassword := viper.GetString("db_password")
	dbName := viper.GetString("db_name")
	signingKey := []byte(viper.GetString("jwtSecretKey"))

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

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		dbUser, dbPassword, dbName)

	logInfo.Printf("Logging into postgres database with following credentials: %s", dbinfo)

	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logError.Fatalf("Error opening connection with database: %s", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			logError.Fatalf("Error closing connection with database: %s", err)
		}
	}()

	env.LogInfo = logInfo
	env.LogError = logError
	env.Admins = admins
	env.SigningKey = signingKey
	env.Db = db

	router := mux.NewRouter()

	router.Handle("/api/posts", api.GetPosts(env)).Queries("page", "{page}",
		"posts-per-page", "{posts-per-page}").Methods("GET")
	router.Handle("/api/posts", api.GetPosts(env)).Queries("page", "{page}").Methods("GET")
	router.Handle("/api/posts/{id}", api.GetCertainPost(env)).Methods("GET")
	router.Handle("/api/posts",
		jwtMiddleware.Handler(api.JwtAuthentication(env, api.CreatePost(env)))).Methods("POST")
	router.Handle("/api/posts/{id}",
		jwtMiddleware.Handler(api.JwtAuthentication(env, api.UpdatePost(env)))).Methods("PUT")
	router.Handle("/api/posts/{id}",
		jwtMiddleware.Handler(api.JwtAuthentication(env, api.DeletePost(env)))).Methods("DELETE")
	router.HandleFunc("/api/hc", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}).Methods("GET")
	router.Handle("/api/user/register", api.RegisterUserHandler(env)).Methods("POST")
	router.Handle("/api/user/login", api.LoginUserHandler(env)).Methods("POST")

	router.PathPrefix("/css").Handler(http.StripPrefix("/css", http.FileServer(http.Dir(frontFolder+"/css"))))
	router.PathPrefix("/scripts").Handler(http.StripPrefix("/scripts", http.FileServer(http.Dir(frontFolder+"/scripts"))))
	router.PathPrefix("/images").Handler(http.StripPrefix("/images", http.FileServer(http.Dir(frontFolder+"/images"))))
	router.PathPrefix("/posts/{id}").Handler(web.GeneratePostPage(env))
	router.PathPrefix("/").Handler(http.StripPrefix("/", web.HandleHTMLFile(env, frontFolder)))

	logInfo.Printf("listening on address %s", Address)
	logError.Fatal(http.ListenAndServe(Address, router))
}

func main() {
	userConfigPath := pflag.StringP("config", "c",
		filepath.FromSlash("configs/config.json"), "config file path")
	adminsListConfigPath := pflag.StringP("admins", "a",
		filepath.FromSlash("configs/admins.json"), "admins list")

	pflag.Parse()
	RunServer(*userConfigPath, *adminsListConfigPath)
}

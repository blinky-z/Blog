package server

import (
	"database/sql"
	"fmt"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/blinky-z/Blog/handler/renderApi"
	"github.com/blinky-z/Blog/handler/restApi"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	logInfo  = log.New(os.Stdout, "INFO: ", log.Ltime)
	logError = log.New(os.Stderr, "ERROR: ", log.Ltime)

	Db                 *sql.DB
	frontendFolderPath = filepath.FromSlash("front/")
	jwtMiddleware      *jwtmiddleware.JWTMiddleware
)

// keys to access env variables
const (
	dbUserEnvKey     string = "db_user"
	dbPasswordEnvKey string = "db_password"
	dbNameEnvKey     string = "db_name"
	dbHostEnvKey     string = "db_host"
	dbPortEnvKey     string = "db_port"
	jwtSecretEnvKey  string = "jwt_secret_key"
	adminsEnvKey     string = "admins"
	serverPortEnvKey string = "server_port"
)

// we need to export this function to use in tests
func RunServer() {
	// bind env variables. Access them by the same key
	_ = viper.BindEnv(dbUserEnvKey, "DB_USER")
	_ = viper.BindEnv(dbPasswordEnvKey, "DB_PASSWORD")
	_ = viper.BindEnv(dbNameEnvKey, "DB_NAME")
	_ = viper.BindEnv(dbHostEnvKey, "DB_HOST")
	_ = viper.BindEnv(dbPortEnvKey, "DB_PORT")
	_ = viper.BindEnv(jwtSecretEnvKey, "JWT_SECRET_KEY")
	_ = viper.BindEnv(adminsEnvKey, "ADMINS")
	_ = viper.BindEnv(serverPortEnvKey, "SERVER_PORT")

	dbUser := viper.GetString(dbUserEnvKey)
	dbPassword := viper.GetString(dbPasswordEnvKey)
	dbName := viper.GetString(dbNameEnvKey)
	dbHost := viper.GetString(dbHostEnvKey)
	dbPort := viper.GetString(dbPortEnvKey)
	jwtSecret := []byte(viper.GetString(jwtSecretEnvKey))

	var admins []string
	err := viper.UnmarshalKey(adminsEnvKey, &admins)
	if err != nil {
		logError.Fatalf("Error unmarshaling admins list: %s", err)
	}

	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	logInfo.Printf("Opening database on host=%s, port=%s, user=%s, db name=%s...", dbHost, dbPort, dbUser, dbName)
	Db, err = sql.Open("postgres", connString)
	if err != nil {
		logError.Fatalf("Error opening database: %s", err)
	}
	defer func() {
		err := Db.Close()
		if err != nil {
			logError.Printf("Error closing database: %s", err)
		}
	}()
	// validate data source
	if err = Db.Ping(); err != nil {
		logError.Fatalf("Invalid data source: %s", err)
	}
	logInfo.Print("Database successfully opened")

	// create JWT Middleware
	// it intercepts requests on secured paths and checks jwt token
	jwtUserPropety := "user"
	jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
		UserProperty: jwtUserPropety,
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			restApi.RespondWithError(w, http.StatusUnauthorized, restApi.InvalidToken)
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	commentApiHandler := restApi.NewCommentApiHandler(Db,
		log.New(os.Stdout, "[api.comment] INFO: ", log.Ltime),
		log.New(os.Stderr, "[api.comment] ERROR: ", log.Ltime))
	postApiHandler := restApi.NewPostApiHandler(Db,
		log.New(os.Stdout, "[api.post] INFO: ", log.Ltime),
		log.New(os.Stderr, "[api.post] ERROR: ", log.Ltime))
	userApiHandler := restApi.NewUserApiHandler(Db,
		jwtSecret,
		&admins,
		jwtUserPropety,
		log.New(os.Stdout, "[api.user] INFO: ", log.Ltime),
		log.New(os.Stderr, "[api.user] ERROR: ", log.Ltime))
	renderApiHandler := renderApi.NewRenderApiHandler(Db,
		&admins,
		log.New(os.Stdout, "[renderApi.render] INFO: ", log.Ltime),
		log.New(os.Stderr, "[renderApi.render] ERROR: ", log.Ltime))

	router := mux.NewRouter()

	// set api handlers
	router.Handle("/api/posts", postApiHandler.GetPostsHandler()).Queries("page", "{page}",
		"posts-per-page", "{posts-per-page}").Methods("GET")
	router.Handle("/api/posts", postApiHandler.GetPostsHandler()).Queries("page", "{page}").Methods("GET")
	router.Handle("/api/posts/{id}", postApiHandler.GetCertainPostHandler()).Methods("GET")
	router.Handle("/api/posts",
		jwtMiddleware.Handler(userApiHandler.FgpAuthentication(postApiHandler.CreatePostHandler()))).Methods("POST")
	router.Handle("/api/posts/{id}",
		jwtMiddleware.Handler(userApiHandler.FgpAuthentication(postApiHandler.UpdatePostHandler()))).Methods("PUT")
	router.Handle("/api/posts/{id}",
		jwtMiddleware.Handler(userApiHandler.FgpAuthentication(postApiHandler.DeletePostHandler()))).Methods("DELETE")
	router.HandleFunc("/api/hc", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}).Methods("GET")
	router.Handle("/api/user/register", userApiHandler.RegisterUserHandler()).Methods("POST")
	router.Handle("/api/user/login", userApiHandler.LoginUserHandler()).Methods("POST")
	router.Handle("/api/comments", commentApiHandler.CreateCommentHandler()).Methods("POST")
	router.Handle("/api/comments/{id}",
		jwtMiddleware.Handler(userApiHandler.FgpAuthentication(commentApiHandler.UpdateCommentHandler()))).Methods("PUT")
	router.Handle("/api/comments/{id}",
		jwtMiddleware.Handler(userApiHandler.FgpAuthentication(commentApiHandler.DeleteCommentHandler()))).Methods("DELETE")

	// set frontend files paths
	router.PathPrefix("/css").Handler(
		http.StripPrefix("/css", http.FileServer(http.Dir(frontendFolderPath+"/css"))))
	router.PathPrefix("/scripts").Handler(
		http.StripPrefix("/scripts", http.FileServer(http.Dir(frontendFolderPath+"/scripts"))))
	router.PathPrefix("/images").Handler(
		http.StripPrefix("/images", http.FileServer(http.Dir(frontendFolderPath+"/images"))))

	// set pages rendering handlers
	router.PathPrefix("/posts/{id}").Handler(renderApiHandler.RenderPostPageHandler())
	router.PathPrefix("/").Handler(http.StripPrefix("/", renderApi.HandleHTMLFile(renderApiHandler, frontendFolderPath)))

	serverPort := viper.GetString(serverPortEnvKey)
	logInfo.Printf("Starting server on port %s", serverPort)
	// omitting host will run server on all interfaces
	logError.Fatal(http.ListenAndServe(":"+serverPort, router))
}

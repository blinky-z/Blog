package server

import (
	"database/sql"
	"fmt"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/blinky-z/Blog/handler/renderapi"
	"github.com/blinky-z/Blog/handler/restapi"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // import postgres driver
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

var (
	logInfo  = log.New(os.Stdout, "INFO: ", log.Ltime)
	logError = log.New(os.Stderr, "ERROR: ", log.Ltime)

	Db                  *sql.DB
	frontendLayoutsPath = filepath.FromSlash("front/layouts/")
	frontendStaticPath  = filepath.FromSlash("front/static/")
	jwtMiddleware       *jwtmiddleware.JWTMiddleware
)

// keys to access env variables
const (
	dbUserEnvKey     string = "db_user"
	dbPasswordEnvKey string = "db_password"
	dbNameEnvKey     string = "db_name"
	dbHostEnvKey     string = "db_host"
	dbPortEnvKey     string = "db_port"
	domainEnvKey     string = "domain"
	//jwtSecretEnvKey  string = "jwt_secret_key"
	//adminsEnvKey     string = "admins"
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
	_ = viper.BindEnv(domainEnvKey, "DOMAIN")
	//_ = viper.BindEnv(jwtSecretEnvKey, "JWT_SECRET_KEY")
	//_ = viper.BindEnv(adminsEnvKey, "ADMINS")
	_ = viper.BindEnv(serverPortEnvKey, "SERVER_PORT")

	dbUser := viper.GetString(dbUserEnvKey)
	dbPassword := viper.GetString(dbPasswordEnvKey)
	dbName := viper.GetString(dbNameEnvKey)
	dbHost := viper.GetString(dbHostEnvKey)
	dbPort := viper.GetString(dbPortEnvKey)
	domain, err := url.Parse(viper.GetString(domainEnvKey))
	if err != nil {
		logError.Fatalf("Error parsing domain: %s", err)
	}
	//jwtSecret := []byte(viper.GetString(jwtSecretEnvKey))

	//var admins []string
	//err := viper.UnmarshalKey(adminsEnvKey, &admins)
	//if err != nil {
	//	logError.Fatalf("Error unmarshaling admins list: %s", err)
	//}

	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	logInfo.Printf("Opening database on host=%s, port=%s, user=%s, db name=%s...", dbHost, dbPort, dbUser, dbName)
	Db, err := sql.Open("postgres", connString)
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
	//jwtUserProperty := "user"
	//jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	//	UserProperty: jwtUserProperty,
	//	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
	//		return jwtSecret, nil
	//	},
	//	ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
	//		restapi.RespondWithError(w, http.StatusUnauthorized, restapi.InvalidToken)
	//	},
	//	SigningMethod: jwt.SigningMethodHS256,
	//})

	postAPIHandler := restapi.NewPostAPIHandler(Db,
		log.New(os.Stdout, "[restApi.post] INFO: ", log.Ltime),
		log.New(os.Stderr, "[restApi.post] ERROR: ", log.Ltime))
	tagAPIHandler := restapi.NewTagAPIHandler(Db,
		log.New(os.Stdout, "[restApi.tag] INFO: ", log.Ltime),
		log.New(os.Stderr, "[restApi.tag] ERROR: ", log.Ltime))
	//userAPIHandler := restapi.NewUserAPIHandler(Db,
	//	jwtSecret,
	//	&admins,
	//	jwtUserProperty,
	//	log.New(os.Stdout, "[restApi.user] INFO: ", log.Ltime),
	//	log.New(os.Stderr, "[restApi.user] ERROR: ", log.Ltime))
	renderAPIHandler := renderapi.NewRenderAPIHandler(Db,
		frontendLayoutsPath,
		domain,
		log.New(os.Stdout, "[renderApi.render] INFO: ", log.Ltime),
		log.New(os.Stderr, "[renderApi.render] ERROR: ", log.Ltime))

	router := mux.NewRouter()
	mainRouter := router.Host(domain.Host).Subrouter()

	// set rest api handlers
	//router.Handle("/api/posts",
	//	jwtMiddleware.Handler(userAPIHandler.FgpAuthentication(postAPIHandler.CreatePostHandler()))).Methods("POST")
	//router.Handle("/api/posts/{id}",
	//	jwtMiddleware.Handler(userAPIHandler.FgpAuthentication(postAPIHandler.UpdatePostHandler()))).Methods("PUT")
	//router.Handle("/api/posts/{id}",
	//	jwtMiddleware.Handler(userAPIHandler.FgpAuthentication(postAPIHandler.DeletePostHandler()))).Methods("DELETE")

	router.HandleFunc("/api/hc", func(w http.ResponseWriter, r *http.Request) {
		if err = Db.Ping(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(200)
	}).Methods("GET")

	// set auth handlers
	//router.Handle("/api/user/register", userAPIHandler.RegisterUserHandler()).Methods("POST")
	//router.Handle("/api/user/login", userAPIHandler.LoginUserHandler()).Methods("POST")

	// set frontend static files paths
	router.PathPrefix("/css").Handler(
		http.StripPrefix("/css", http.FileServer(http.Dir(frontendStaticPath+"/css"))))
	router.PathPrefix("/js").Handler(
		http.StripPrefix("/js", http.FileServer(http.Dir(frontendStaticPath+"/js"))))
	router.PathPrefix("/images").Handler(
		http.StripPrefix("/images", http.FileServer(http.Dir(frontendStaticPath+"/images"))))

	// set pages rendering handlers
	mainRouter.Path("/posts").Handler(renderAPIHandler.RenderAllPostsPageHandler()).Methods("GET")
	mainRouter.Path("/posts/{id}").Handler(renderAPIHandler.RenderPostPageHandler()).Methods("GET")
	mainRouter.Path("/tags").Handler(renderAPIHandler.RenderAllTagsPageHandler()).Methods("GET")
	mainRouter.Path("/tags/{tag}").Handler(renderAPIHandler.RenderAllPostsPageHandler()).Methods("GET")
	mainRouter.Path("/about").Handler(renderAPIHandler.RenderAboutPageHandler()).Methods("GET")
	mainRouter.Path("/index").Handler(renderAPIHandler.RenderIndexPageHandler()).Methods("GET")
	mainRouter.Path("/").Handler(renderAPIHandler.RenderIndexPageHandler()).Methods("GET")
	mainRouter.Path("/robots.txt").Handler(http.FileServer(http.Dir(""))).Methods("GET")
	mainRouter.Path("/sitemap").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "sitemap.xml")
	}).Methods("GET")

	adminRouter := router.Host("admin." + domain.Host).Subrouter()
	adminRouter.Path("/").Handler(renderAPIHandler.RenderAdminPageHandler()).Methods("GET")
	adminRouter.Path("/editor").Handler(renderAPIHandler.RenderAdminEditorPageHandler()).Methods("GET")
	adminRouter.Path("/manage-posts").Handler(renderAPIHandler.RenderAdminManagePostsPageHandler()).Methods("GET")
	adminRouter.Path("/manage-tags").Handler(renderAPIHandler.RenderAdminManageTagsPageHandler()).Methods("GET")
	adminRouter.Path("/robots.txt").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "robots_admin.txt")
	}).Methods("GET")

	// set blog posts related rest api
	adminRouter.Handle("/api/posts", postAPIHandler.CreatePostHandler()).Methods("POST")
	adminRouter.Handle("/api/posts/{id}", postAPIHandler.UpdatePostHandler()).Methods("PUT")
	adminRouter.Handle("/api/posts/{id}", postAPIHandler.DeletePostHandler()).Methods("DELETE")
	adminRouter.Handle("/api/tags", tagAPIHandler.CreateTagHandler()).Methods("POST")
	adminRouter.Handle("/api/tags/{id}", tagAPIHandler.UpdateTagHandler()).Methods("PUT")
	adminRouter.Handle("/api/tags/{id}", tagAPIHandler.DeleteTagHandler()).Methods("DELETE")

	serverPort := viper.GetString(serverPortEnvKey)
	logInfo.Printf("Starting server on port %s", serverPort)
	// omitting host will run server on all interfaces
	logError.Fatal(http.ListenAndServe(":"+serverPort, router))
}

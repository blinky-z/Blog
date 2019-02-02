package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/auth0/go-jwt-middleware"
	"github.com/blinky-z/Blog/handler"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/postService"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"html/template"
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

	frontFolder     = "front/"
	templatesFolder = frontFolder + "templates/"

	env *models.Env = &models.Env{}

	jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return env.SigningKey, nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			logInfo.Printf("Error checking JWT Token: Malformed or invalid JWT Token")
			handler.RespondWithError(w, http.StatusUnauthorized, handler.InvalidToken, env.LogError)
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
)

type BlogPost struct {
	PostLink         string
	PostTitle        string
	PostAuthor       string
	PostCreationTime string
	PostSnippet      string
	PostContent      string
}

type PostsList struct {
	Posts []BlogPost
}

type PageSelector struct {
	HasNewerPosts  bool
	NewerPostsLink string
	OlderPostsLink string
	HasOlderPosts  bool
}

type IndexPage struct {
	PostsList
	PageSelector
}

type PostPage struct {
	BlogPost
}

func GeneratePostPage(w http.ResponseWriter, r *http.Request) {
	logInfo.Printf("Rendering post page")

	id, validateIDError := handler.ValidatePostID(r)
	if validateIDError != handler.NoError {
		env.LogInfo.Print("Can not GET post: post ID is invalid")
		handler.RespondWithError(w, http.StatusBadRequest, validateIDError, env.LogError)
		return
	}

	logInfo.Printf("Getting post with id %s from database", id)
	post, err := postService.GetCertainPost(env, id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			handler.RespondWithError(w, http.StatusNotFound, handler.NoSuchPost, env.LogError)
		default:
			handler.RespondWithError(w, http.StatusInternalServerError, handler.TechnicalError, env.LogError)
		}
	}

	logInfo.Printf("Getting post page template")
	postTemplate, err :=
		template.ParseFiles(templatesFolder+"header.html", templatesFolder+"postPage.html", templatesFolder+"footer.html")
	if err != nil {
		logError.Print(err)
		handler.RespondWithError(w, http.StatusInternalServerError, handler.TechnicalError, env.LogError)
		return
	}

	logInfo.Printf("Setting post page template data")

	var data PostPage

	data.PostTitle = post.Title
	data.PostAuthor = "Dmitry"
	data.PostCreationTime = post.Date.Format("Mon Jan 2 15:04:05")
	data.PostContent = post.Content

	logInfo.Printf("Executing post template")
	if err := postTemplate.ExecuteTemplate(w, "postPage", data); err != nil {
		if err != nil {
			logError.Print(err)
			handler.RespondWithError(w, http.StatusInternalServerError, handler.TechnicalError, env.LogError)
		}
	}
}

func GenerateIndexPage(w http.ResponseWriter, r *http.Request) {
	logInfo.Printf("Rendering index page")

	params, validateError := handler.ValidateGetPostsParams(r)
	if validateError != handler.NoError {
		logInfo.Print("Can not GET range of posts : get posts Query params are invalid")
		handler.RespondWithError(w, http.StatusBadRequest, validateError, env.LogError)
		return
	}
	page := params.Page

	logInfo.Printf("Getting posts from database")
	posts, err := postService.GetPosts(env, page, 11)
	if err != nil {
		logError.Print(err)
		handler.RespondWithError(w, http.StatusInternalServerError, handler.TechnicalError, env.LogError)
		return
	}

	logInfo.Printf("Getting index page template")
	indexTemplate, err :=
		template.ParseFiles(templatesFolder+"header.html", templatesFolder+"indexPage.html", templatesFolder+"footer.html")
	if err != nil {
		logError.Print(err)
		handler.RespondWithError(w, http.StatusInternalServerError, handler.TechnicalError, env.LogError)
		return
	}

	logInfo.Printf("Setting index page template data")

	var data IndexPage

	for currentPostNum := 0; currentPostNum < len(posts) && currentPostNum < 10; currentPostNum++ {
		var blogPostData BlogPost
		post := posts[currentPostNum]
		blogPostData.PostTitle = post.Title
		blogPostData.PostAuthor = "Dmitry"
		blogPostData.PostCreationTime = post.Date.Format("Mon Jan 2 15:04:05")
		blogPostData.PostLink = fmt.Sprintf("/posts/%s", post.ID)
		if len(post.Content) < 160 {
			blogPostData.PostSnippet = post.Content
		} else {
			blogPostData.PostSnippet = post.Content[:160]
		}

		data.Posts = append(data.Posts, blogPostData)
	}

	if page != 0 {
		data.HasNewerPosts = true
		data.NewerPostsLink = fmt.Sprintf("/?page=%d", page-1)
	} else {
		data.HasNewerPosts = false
	}

	if len(posts) > 10 {
		data.HasOlderPosts = true
		data.OlderPostsLink = fmt.Sprintf("/?page=%d", page+1)
	} else {
		data.HasOlderPosts = false
	}

	logInfo.Printf("Executing index template")
	if err := indexTemplate.ExecuteTemplate(w, "indexPage", data); err != nil {
		if err != nil {
			logError.Print(err)
			handler.RespondWithError(w, http.StatusInternalServerError, handler.TechnicalError, env.LogError)
		}
	}
}

var handleHTMLFile = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	currentURLPath := r.URL.Path
	currentURLPath = strings.TrimSuffix(currentURLPath, ".html")

	var fileName string
	if currentURLPath == "" {
		GenerateIndexPage(w, r)
		return
	}

	fileName = currentURLPath + ".html"

	filePath := frontFolder + fileName

	http.ServeFile(w, r, filePath)
})

var handleHTMLPost = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	GeneratePostPage(w, r)
})

var handleHTMLAdminPage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// TODO: реализовать доступ к админке только админам
	// Пока это невозможно, так как я должен прикреплять к запросу токен. Сейчас у меня нет html элемента для входа
	// в админку, чтобы при нажатии по нему срабатывал скрипт.
	//userRole := r.Context().Value(handler.CtxKey).(string)
	//if userRole != "admin" {
	//	logInfo.Printf("User with role %s doesn't have access to admin page", userRole)
	//	handler.RespondWithError(w, http.StatusForbidden, handler.NoPermissions)
	//	return
	//}

	currentURLPath := r.URL.Path
	currentURLPath = strings.TrimSuffix(currentURLPath, ".html")

	var fileName string
	if currentURLPath == "" {
		fileName = "admin.html"
	} else {
		fileName = currentURLPath + ".html"
	}

	filePath := frontFolder + fileName

	logInfo.Printf("Current admin page path: %s", filePath)

	http.ServeFile(w, r, filePath)
})

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

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s",
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

	router.Handle("/api/posts", handler.GetPosts(env)).Queries("page", "{page}",
		"posts-per-page", "{posts-per-page}").Methods("GET")
	router.Handle("/api/posts", handler.GetPosts(env)).Queries("page", "{page}").Methods("GET")
	router.Handle("/api/posts/{id}", handler.GetCertainPost(env)).Methods("GET")
	router.Handle("/api/posts",
		jwtMiddleware.Handler(handler.JwtAuthentication(env, handler.CreatePost(env)))).Methods("POST")
	router.Handle("/api/posts/{id}",
		jwtMiddleware.Handler(handler.JwtAuthentication(env, handler.UpdatePost(env)))).Methods("PUT")
	router.Handle("/api/posts/{id}",
		jwtMiddleware.Handler(handler.JwtAuthentication(env, handler.DeletePost(env)))).Methods("DELETE")
	router.HandleFunc("/api/hc", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}).Methods("GET")
	router.Handle("/api/user/register", handler.RegisterUserHandler(env)).Methods("POST")
	router.Handle("/api/user/login", handler.LoginUserHandler(env)).Methods("POST")

	router.PathPrefix("/css").Handler(http.StripPrefix("/css", http.FileServer(http.Dir(frontFolder+"/css"))))
	router.PathPrefix("/scripts").Handler(http.StripPrefix("/scripts", http.FileServer(http.Dir(frontFolder+"/scripts"))))
	router.PathPrefix("/images").Handler(http.StripPrefix("/images", http.FileServer(http.Dir(frontFolder+"/images"))))
	router.PathPrefix("/posts/{id}").Handler(handleHTMLPost)
	//router.PathPrefix("/admin").Handler(http.StripPrefix("/admin",
	//	jwtMiddleware.Handler(handler.JwtAuthentication(handleHTMLAdminPage))))
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

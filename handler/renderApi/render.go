package renderApi

import (
	"database/sql"
	"fmt"
	"github.com/blinky-z/Blog/handler/restApi"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/service/commentService"
	"github.com/blinky-z/Blog/service/postService"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type RenderApiHandler struct {
	db       *sql.DB
	admins   *[]string
	logInfo  *log.Logger
	logError *log.Logger
}

func NewRenderApiHandler(db *sql.DB, admins *[]string, logInfo, logError *log.Logger) *RenderApiHandler {
	return &RenderApiHandler{
		db:       db,
		admins:   admins,
		logInfo:  logInfo,
		logError: logError,
	}
}

var (
	templatesFolder = filepath.FromSlash("front/templates/")

	siteMetadataDescription = "Blog about programming"
	siteMetadataKeywords    = []string{"Programming", "Linux"}
)

const (
	timeFormat              = "January 2 2006, 15:04:05"
	postsPerPageDefault int = 10
)

// pageSelector - represents page selector on index page
type pageSelector struct {
	HasNewerPosts  bool
	NewerPostsLink string
	HasOlderPosts  bool
	OlderPostsLink string
}

// indexPage - represents index page
type indexPage struct {
	Metadata     models.MetaData
	Posts        []models.Post
	PageSelector pageSelector
}

// postPage - represents /posts/{id} page
type postPage struct {
	Metadata      models.MetaData
	Post          models.Post
	Comments      []*models.CommentWithChilds
	CommentsCount int
	IsUserAdmin   bool
}

func countComments(comments []*models.CommentWithChilds) int {
	if len(comments) == 0 {
		return 0
	}

	l1 := len(comments)
	for _, currentComment := range comments {
		l1 += countComments(currentComment.Childs)
	}

	return l1
}

var convertTimeTemplateFunc = template.FuncMap{
	"convertTime": func(t time.Time) string {
		return t.Format(timeFormat)
	},
}

// RenderPostPageHandler - handler for server-side rendering of /posts/{id} page
func (renderApi *RenderApiHandler) RenderPostPageHandler() http.Handler {
	logError := renderApi.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := mux.Vars(r)["id"]
		validateIDError := restApi.ValidateID(postID)
		if validateIDError != restApi.NoError {
			restApi.Respond(w, http.StatusNotFound)
			return
		}

		var isUserAdmin bool

		usernameCookie, err := r.Cookie("Login")
		if err != nil {
			isUserAdmin = false
		} else {
			isUserAdmin = restApi.IsUserAdmin(usernameCookie.Value, renderApi.admins)
		}

		post, err := postService.GetById(renderApi.db, postID)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				restApi.Respond(w, http.StatusNotFound)
				return
			default:
				restApi.Respond(w, http.StatusInternalServerError)
				return
			}
		}

		incTemplateFunc := template.FuncMap{
			"inc": func(i int) int {
				return i + 1
			},
		}

		passArgsTemplateFunc := template.FuncMap{
			"args": func(vs ...interface{}) []interface{} {
				return vs
			},
		}

		postTemplate, err := template.New("").
			Funcs(incTemplateFunc).
			Funcs(passArgsTemplateFunc).
			Funcs(convertTimeTemplateFunc).
			ParseFiles(templatesFolder+"header.html", templatesFolder+"comments-list.html", templatesFolder+"comment.html",
				templatesFolder+"postPage.html", templatesFolder+"footer.html")
		if err != nil {
			restApi.Respond(w, http.StatusInternalServerError)
			return
		}

		var data postPage

		comments, err := commentService.GetAllByPostId(renderApi.db, postID)
		if err != nil {
			restApi.Respond(w, http.StatusInternalServerError)
			return
		}

		data.Post = post
		data.Comments = comments
		data.CommentsCount = countComments(comments)
		data.Metadata = post.Metadata
		data.IsUserAdmin = isUserAdmin

		if err := postTemplate. /*Funcs(incTemplateFunc).Funcs(passArgsTemplateFunc).Funcs(convertTimeTemplateFunc).*/
					ExecuteTemplate(w, "postPage", data); err != nil {
			logError.Printf("Error rendering single post page: %s", err)
			restApi.Respond(w, http.StatusInternalServerError)
		}
	})
}

// renderIndexPageHandler - handler for server-side rendering of index page
func (renderApi *RenderApiHandler) renderIndexPageHandler() http.Handler {
	logError := renderApi.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeParams := &restApi.GetPostsRequestQueryParams{
			Page:         r.FormValue("page"),
			PostsPerPage: "",
		}

		validateQueryParamsError := restApi.ValidateGetPostsRequestQueryParams(rangeParams)
		if validateQueryParamsError != restApi.NoError {
			restApi.Respond(w, http.StatusNotFound)
			return
		}
		page, _ := strconv.Atoi(rangeParams.Page)

		posts, err := postService.GetPostsInRange(renderApi.db, page, postsPerPageDefault+1)
		if err != nil {
			restApi.Respond(w, http.StatusInternalServerError)
			return
		}

		indexTemplate, err := template.New("").Funcs(convertTimeTemplateFunc).
			ParseFiles(templatesFolder+"header.html", templatesFolder+"indexPage.html", templatesFolder+"footer.html")
		if err != nil {
			restApi.Respond(w, http.StatusInternalServerError)
			return
		}

		var data indexPage

		if page != 0 {
			data.PageSelector.HasNewerPosts = true
			data.PageSelector.NewerPostsLink = fmt.Sprintf("/?page=%d", page-1)
		} else {
			data.PageSelector.HasNewerPosts = false
		}

		// hack here: if we were able to retrieve more posts than default value, then we have older posts
		if len(posts) > postsPerPageDefault {
			data.PageSelector.HasOlderPosts = true
			data.PageSelector.OlderPostsLink = fmt.Sprintf("/?page=%d", page+1)

			// remove very last post, as we need less posts
			posts = posts[:postsPerPageDefault]
		} else {
			data.PageSelector.HasOlderPosts = false
		}

		data.Posts = posts
		data.Metadata.Description = siteMetadataDescription
		data.Metadata.Keywords = siteMetadataKeywords

		if err := indexTemplate. /*.Funcs(convertTimeTemplateFunc)*/ ExecuteTemplate(w, "indexPage", data); err != nil {
			logError.Printf("Error rendering index page: %s", err)
			restApi.Respond(w, http.StatusInternalServerError)
		}
	})
}

// HandleHTMLFile - render html page
func HandleHTMLFile(renderApi *RenderApiHandler, frontFolder string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentURLPath := r.URL.Path
		currentURLPath = strings.TrimSuffix(currentURLPath, ".html")
		currentURLPath = filepath.FromSlash(currentURLPath)

		var fileName string
		if currentURLPath == "" || currentURLPath == "index" {
			renderApi.renderIndexPageHandler().ServeHTTP(w, r)
			return
		}

		fileName = currentURLPath + ".html"
		filePath := frontFolder + fileName
		http.ServeFile(w, r, filePath)
	})
}

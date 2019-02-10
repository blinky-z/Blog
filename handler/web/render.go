package web

import (
	"database/sql"
	"fmt"
	"github.com/blinky-z/Blog/commentService"
	"github.com/blinky-z/Blog/handler/api"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/postService"
	"github.com/gorilla/mux"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var (
	templatesFolder = filepath.FromSlash("front/templates/")
)

const (
	timeFormat            = "January 2 2006, 15:04:05"
	postsOnPageAmount int = 10
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

// GeneratePostPage - handler for server-side rendering /posts/{id} page
func GeneratePostPage(env *models.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.LogInfo.Printf("Rendering post page")

		postID := mux.Vars(r)["id"]
		validateIDError := api.ValidateID(postID)
		if validateIDError != api.NoError {
			env.LogInfo.Print("Can not GET post: post ID is invalid")
			api.Respond(w, http.StatusNotFound)
			return
		}

		var isUserAdmin bool

		usernameCookie, err := r.Cookie("Login")
		if err != nil {
			isUserAdmin = false
		} else {
			isUserAdmin = api.IsUserAdmin(usernameCookie.Value, api.UserEnv.Admins)
		}

		env.LogInfo.Printf("Getting post with id %s from database", postID)
		post, err := postService.GetCertainPost(env, postID)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				api.Respond(w, http.StatusNotFound)
				return
			default:
				api.Respond(w, http.StatusInternalServerError)
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

		env.LogInfo.Printf("Getting post page template")
		postTemplate, err := template.New("").Funcs(incTemplateFunc).Funcs(passArgsTemplateFunc).
			Funcs(convertTimeTemplateFunc).
			ParseFiles(templatesFolder+"header.html", templatesFolder+"comments-list.html", templatesFolder+"comment.html",
				templatesFolder+"postPage.html", templatesFolder+"footer.html")
		if err != nil {
			env.LogError.Print(err)
			api.Respond(w, http.StatusInternalServerError)
			return
		}

		env.LogInfo.Printf("Setting post page template data")

		var data postPage

		comments, err := commentService.GetComments(env, postID)
		if err != nil {
			env.LogError.Print(err)
			api.Respond(w, http.StatusInternalServerError)
			return
		}

		data.Post = post
		data.Comments = comments
		data.CommentsCount = countComments(comments)
		data.Metadata.Description = post.Metadata.Description
		data.Metadata.Keywords = post.Metadata.Keywords
		data.IsUserAdmin = isUserAdmin

		env.LogInfo.Printf("Executing post template")
		if err := postTemplate.Funcs(incTemplateFunc).Funcs(passArgsTemplateFunc).Funcs(convertTimeTemplateFunc).
			ExecuteTemplate(w, "postPage", data); err != nil {
			env.LogError.Print(err)
		}
	})
}

// GenerateIndexPage - handler for server-side rendering index page
func GenerateIndexPage(env *models.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.LogInfo.Printf("Rendering index page")

		params, validateError := api.ValidateGetPostsParams(r)
		if validateError != api.NoError {
			env.LogInfo.Print("Can not GET range of posts : get posts Query params are invalid")
			api.Respond(w, http.StatusNotFound)
			return
		}
		page := params.Page

		env.LogInfo.Printf("Getting posts from database")
		posts, err := postService.GetPosts(env, page, 11)
		if err != nil {
			env.LogError.Print(err)
			api.Respond(w, http.StatusInternalServerError)
			return
		}

		if len(posts) == 0 {
			api.Respond(w, http.StatusNotFound)
			return
		}

		env.LogInfo.Printf("Getting index page template")
		indexTemplate, err :=
			template.New("").Funcs(convertTimeTemplateFunc).
				ParseFiles(templatesFolder+"header.html", templatesFolder+"indexPage.html", templatesFolder+"footer.html")
		if err != nil {
			env.LogError.Print(err)
			api.Respond(w, http.StatusInternalServerError)
			return
		}

		env.LogInfo.Printf("Setting index page template data")

		var data indexPage

		if page != 0 {
			data.PageSelector.HasNewerPosts = true
			data.PageSelector.NewerPostsLink = fmt.Sprintf("/?page=%d", page-1)
		} else {
			data.PageSelector.HasNewerPosts = false
		}

		if len(posts) > postsOnPageAmount {
			data.PageSelector.HasOlderPosts = true
			data.PageSelector.OlderPostsLink = fmt.Sprintf("/?page=%d", page+1)
		} else {
			data.PageSelector.HasOlderPosts = false
		}

		if len(posts) > postsOnPageAmount {
			posts = posts[:postsOnPageAmount]
		}
		data.Posts = posts

		data.Metadata.Description = "Blog about programming"
		data.Metadata.Keywords = []string{"Programming"}

		env.LogInfo.Printf("Executing index template")
		if err := indexTemplate.Funcs(convertTimeTemplateFunc).ExecuteTemplate(w, "indexPage", data); err != nil {
			env.LogError.Print(err)
		}
	})
}

// HandleHTMLFile - handle html page. If requested page is index page or post page then render it on server-side and
// return rendered page, otherwise return empty html page that will be rendered on client-side
func HandleHTMLFile(env *models.Env, frontFolder string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentURLPath := r.URL.Path
		currentURLPath = strings.TrimSuffix(currentURLPath, ".html")
		currentURLPath = filepath.FromSlash(currentURLPath)

		var fileName string
		if currentURLPath == "" || currentURLPath == "index" {
			GenerateIndexPage(env).ServeHTTP(w, r)
			return
		}

		fileName = currentURLPath + ".html"

		filePath := frontFolder + fileName

		http.ServeFile(w, r, filePath)
	})
}

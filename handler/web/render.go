package web

import (
	"database/sql"
	"fmt"
	"github.com/blinky-z/Blog/commentService"
	"github.com/blinky-z/Blog/handler/api"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/postService"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

var (
	templatesFolder = filepath.FromSlash("front/templates/")
)

const (
	timeFormat = "January 2 2006, 15:04:05"
)

// postsList - represents posts list on index page
type postsList struct {
	Posts []blogPost
}

// pageSelector - represents page selector on index page
type pageSelector struct {
	HasNewerPosts  bool
	NewerPostsLink string
	OlderPostsLink string
	HasOlderPosts  bool
}

// indexPage - represents index page
type indexPage struct {
	models.MetaData
	postsList
	pageSelector
}

// commentWithChilds - represents comment in comments section
type commentWithChilds struct {
	CommentID      string
	Username       string
	CreationTime   string
	CommentContent string
	Childs         []*commentWithChilds
}

// postCommentsList - represents comments section below post
type postCommentsList struct {
	Comments []commentWithChilds
}

// blogPost - represents blog post on index and /posts/{id} pages
type blogPost struct {
	PostLink         string
	PostTitle        string
	PostAuthor       string
	PostCreationTime string
	PostSnippet      string
	PostContent      string
}

// postPage - represents /posts/{id} page
type postPage struct {
	models.MetaData
	blogPost
	CommentsCount int
	postCommentsList
	IsUserAdmin bool
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
			ParseFiles(templatesFolder+"header.html", templatesFolder+"comments-list.html", templatesFolder+"comment.html",
				templatesFolder+"postPage.html", templatesFolder+"footer.html")
		if err != nil {
			env.LogError.Print(err)
			api.Respond(w, http.StatusInternalServerError)
			return
		}

		env.LogInfo.Printf("Setting post page template data")

		var data postPage

		data.PostTitle = post.Title
		data.PostAuthor = "Dmitry"
		data.PostCreationTime = post.Date.Format(timeFormat)
		data.PostContent = post.Content

		var postMetadata models.MetaData
		postMetadata.Description = post.Metadata.Description
		postMetadata.Keywords = post.Metadata.Keywords

		data.MetaData = postMetadata

		comments, err := commentService.GetComments(env, postID)
		if err != nil {
			env.LogError.Print(err)
			api.Respond(w, http.StatusInternalServerError)
			return
		}

		commentWithChildsAsMap := make(map[string]*commentWithChilds)
		var parentComments []string

		for _, comment := range comments {
			commentWithChilds := &commentWithChilds{}
			commentWithChilds.CommentID = comment.ID
			commentWithChilds.CommentContent = comment.Content
			commentWithChilds.CreationTime = comment.Date.Format(timeFormat)
			commentWithChilds.Username = comment.Author

			commentWithChildsAsMap[comment.ID] = commentWithChilds

			if !comment.ParentID.Valid {
				parentComments = append(parentComments, comment.ID)
			}
		}

		for _, comment := range comments {
			if comment.ParentID.Valid {
				parent := commentWithChildsAsMap[comment.ParentID.Value().(string)]
				parent.Childs = append(parent.Childs, commentWithChildsAsMap[comment.ID])
				commentWithChildsAsMap[comment.ParentID.Value().(string)] = parent
			}
		}

		var parentCommentWithChilds []commentWithChilds
		for _, parentCommendID := range parentComments {
			parentCommentWithChilds = append(parentCommentWithChilds, *commentWithChildsAsMap[parentCommendID])
		}

		data.Comments = parentCommentWithChilds
		data.CommentsCount = len(commentWithChildsAsMap)

		data.IsUserAdmin = isUserAdmin

		env.LogInfo.Printf("Executing post template")
		if err := postTemplate.Funcs(incTemplateFunc).Funcs(passArgsTemplateFunc).
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
			template.ParseFiles(templatesFolder+"header.html", templatesFolder+"indexPage.html", templatesFolder+"footer.html")
		if err != nil {
			env.LogError.Print(err)
			api.Respond(w, http.StatusInternalServerError)
			return
		}

		env.LogInfo.Printf("Setting index page template data")

		var data indexPage

		for currentPostNum := 0; currentPostNum < len(posts) && currentPostNum < 10; currentPostNum++ {
			var blogPostData blogPost
			post := posts[currentPostNum]
			blogPostData.PostTitle = post.Title
			blogPostData.PostAuthor = "Dmitry"
			blogPostData.PostCreationTime = post.Date.Format(timeFormat)
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

		var indexMetadata models.MetaData
		indexMetadata.Description = "Blog about programming"
		indexMetadata.Keywords = []string{"Programming"}

		data.MetaData = indexMetadata

		env.LogInfo.Printf("Executing index template")
		if err := indexTemplate.ExecuteTemplate(w, "indexPage", data); err != nil {
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

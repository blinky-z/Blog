package web

import (
	"database/sql"
	"fmt"
	"github.com/blinky-z/Blog/handler/api"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/postService"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

var (
	templatesFolder = filepath.FromSlash("front/templates/")
)

// BlogPost - represents blog post on index and /posts/{id} pages
type BlogPost struct {
	PostLink         string
	PostTitle        string
	PostAuthor       string
	PostCreationTime string
	PostSnippet      string
	PostContent      string
}

// PostsList - represents posts list on index page
type PostsList struct {
	Posts []BlogPost
}

// PageSelector - represents page selector on index page
type PageSelector struct {
	HasNewerPosts  bool
	NewerPostsLink string
	OlderPostsLink string
	HasOlderPosts  bool
}

// IndexPage - represents index page
type IndexPage struct {
	PostsList
	PageSelector
	models.MetaData
}

// PostPage - represents /posts/{id} page
type PostPage struct {
	BlogPost
	models.MetaData
}

// GeneratePostPage - handler for server-side rendering /posts/{id} page
func GeneratePostPage(env *models.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.LogInfo.Printf("Rendering post page")

		id, validateIDError := api.ValidatePostID(r)
		if validateIDError != api.NoError {
			env.LogInfo.Print("Can not GET post: post ID is invalid")
			api.RespondWithError(w, http.StatusNotFound, validateIDError, env.LogError)
			return
		}

		env.LogInfo.Printf("Getting post with id %s from database", id)
		post, err := postService.GetCertainPost(env, id)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				api.RespondWithError(w, http.StatusNotFound, api.NoSuchPost, env.LogError)
				return
			default:
				api.RespondWithError(w, http.StatusInternalServerError, api.TechnicalError, env.LogError)
				return
			}
		}

		env.LogInfo.Printf("Getting post page template")
		postTemplate, err :=
			template.ParseFiles(templatesFolder+"header.html", templatesFolder+"postPage.html", templatesFolder+"footer.html")
		if err != nil {
			env.LogError.Print(err)
			api.RespondWithError(w, http.StatusInternalServerError, api.TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("Setting post page template data")

		var data PostPage

		data.PostTitle = post.Title
		data.PostAuthor = "Dmitry"
		data.PostCreationTime = post.Date.Format("Mon Jan 2 15:04:05")
		data.PostContent = post.Content
		data.Description = post.Metadata.Description
		data.Keywords = post.Metadata.Keywords

		env.LogInfo.Printf("Executing post template")
		if err := postTemplate.ExecuteTemplate(w, "postPage", data); err != nil {
			env.LogError.Print(err)
			api.RespondWithError(w, http.StatusInternalServerError, api.TechnicalError, env.LogError)
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
			api.RespondWithError(w, http.StatusNotFound, validateError, env.LogError)
			return
		}
		page := params.Page

		env.LogInfo.Printf("Getting posts from database")
		posts, err := postService.GetPosts(env, page, 11)
		if err != nil {
			env.LogError.Print(err)
			api.RespondWithError(w, http.StatusInternalServerError, api.TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("Getting index page template")
		indexTemplate, err :=
			template.ParseFiles(templatesFolder+"header.html", templatesFolder+"indexPage.html", templatesFolder+"footer.html")
		if err != nil {
			env.LogError.Print(err)
			api.RespondWithError(w, http.StatusInternalServerError, api.TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("Setting index page template data")

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

		var indexMetadata models.MetaData
		indexMetadata.Description = "Blog about programming"
		indexMetadata.Keywords = []string{"Programming"}

		data.Description = indexMetadata.Description
		data.Keywords = indexMetadata.Keywords

		env.LogInfo.Printf("Executing index template")
		if err := indexTemplate.ExecuteTemplate(w, "indexPage", data); err != nil {
			env.LogError.Print(err)
			api.RespondWithError(w, http.StatusInternalServerError, api.TechnicalError, env.LogError)
		}
	})
}

// HandleHTMLFile - handle html page. If requested page is index page or post page then render it on server-side and
// return rendered page, otherwise return empty html page that will be rendered on client-side
func HandleHTMLFile(env *models.Env, frontFolder string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentURLPath := r.URL.Path
		currentURLPath = strings.TrimSuffix(currentURLPath, ".html")

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

package renderapi

import (
	"database/sql"
	"fmt"
	"github.com/blinky-z/Blog/handler/restapi"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/service/postService"
	"github.com/blinky-z/Blog/service/tagService"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"text/template"
	"time"
)

type Handler struct {
	db          *sql.DB
	admins      *[]string
	layoutsPath string
	logInfo     *log.Logger
	logError    *log.Logger
}

func NewRenderAPIHandler(db *sql.DB, admins *[]string, layoutsPath string, logInfo, logError *log.Logger) *Handler {
	return &Handler{
		db:          db,
		admins:      admins,
		layoutsPath: layoutsPath,
		logInfo:     logInfo,
		logError:    logError,
	}
}

const (
	timeFormat           = "January 2 2006, 15:04:05"
	recentPostsCount int = 5
	postsPerPage     int = 10
	siteSuffix           = " | Progbloom - A blog about programming"
)

// pageSelector - represents page selector on index page
type pageSelector struct {
	NewerPostsLink string
	OlderPostsLink string
	HasNewerPosts  bool
	HasOlderPosts  bool
}

// SiteHead - represents <head> tag data
type SiteHead struct {
	Title    string
	Metadata models.MetaData
}

//SiteDescription - represents site description visible on front
type SiteDescription struct {
	Title       string
	Description string
}

var defaultSiteDescription = SiteDescription{
	Title:       "Progbloom",
	Description: "A blog about programming. I write about Linux, Java and low-level programming",
}

// Site - represents all site data
type Site struct {
	Head SiteHead
	Desc SiteDescription
	Data interface{}
}

// indexPageData - represents index page
type indexPageData struct {
	Posts []models.Post
}

// postPageData - represents a single post ("/posts/{id}") page
type postPageData struct {
	Post          models.Post
	Comments      []*models.CommentWithChilds
	CommentsCount int
	IsUserAdmin   bool
}

// postPageData - represents all posts ("/posts") or all posts tagged with ("tags/{tag}) page
type allPostsPageData struct {
	Posts        []models.Post
	PageSelector pageSelector
	Type         string
	Tag          string
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

var renderFuncs = template.FuncMap{
	"convertTime": convertTime,
}

func convertTime(t time.Time) string {
	return t.Format(timeFormat)
}

// RenderPostPageHandler - handler for server-side rendering of /posts/{id} page
func (renderApi *Handler) RenderPostPageHandler() http.Handler {
	logError := renderApi.logError
	layoutsPath := renderApi.layoutsPath
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := mux.Vars(r)["id"]
		if !restapi.IsPostIDValid(postID) {
			restapi.Respond(w, http.StatusNotFound)
			return
		}

		var isUserAdmin bool

		usernameCookie, err := r.Cookie("Login")
		if err != nil {
			isUserAdmin = false
		} else {
			isUserAdmin = restapi.IsUserAdmin(usernameCookie.Value, renderApi.admins)
		}

		returnedPost, err := postService.GetByID(renderApi.db, postID)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				restapi.Respond(w, http.StatusNotFound)
				return
			default:
				restapi.Respond(w, http.StatusInternalServerError)
				return
			}
		}

		tmpl, err := template.New("post").Funcs(renderFuncs).
			ParseFiles(
				layoutsPath+"partials/head.html",
				layoutsPath+"partials/header.html",
				layoutsPath+"partials/footer.html",
				layoutsPath+"post.html")
		if err != nil {
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		var data Site
		data.Head = SiteHead{
			Title:    returnedPost.Title + siteSuffix,
			Metadata: returnedPost.Metadata,
		}
		data.Desc = defaultSiteDescription
		data.Data = postPageData{
			Post:        returnedPost,
			IsUserAdmin: isUserAdmin,
		}

		if err := tmpl.ExecuteTemplate(w, "post", data); err != nil {
			logError.Printf("Error rendering single post page: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
		}
	})
}

// renderIndexPageHandler - handler for server-side rendering of index page
func (renderApi *Handler) RenderIndexPageHandler() http.Handler {
	logError := renderApi.logError
	layoutsPath := renderApi.layoutsPath
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts, err := postService.GetPostsInRange(renderApi.db, 0, recentPostsCount)
		if err != nil {
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		tmpl, err := template.New("index").Funcs(renderFuncs).
			ParseFiles(
				layoutsPath+filepath.FromSlash("partials/head.html"),
				layoutsPath+filepath.FromSlash("partials/header.html"),
				layoutsPath+filepath.FromSlash("partials/footer.html"),
				layoutsPath+"index.html")
		if err != nil {
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		var data Site
		data.Head = SiteHead{
			Title: "Recent Posts" + siteSuffix,
			Metadata: models.MetaData{
				Description: "blog about programming",
				Keywords:    []string{"Programming", "Linux"},
			},
		}
		data.Desc = defaultSiteDescription
		data.Data = indexPageData{
			Posts: posts,
		}

		if err := tmpl.ExecuteTemplate(w, "index", data); err != nil {
			logError.Printf("Error rendering index page: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
		}
	})
}

// RenderAllPostsPageHandler - handler for server-side rendering of all posts page
func (renderApi *Handler) RenderAllPostsPageHandler() http.Handler {
	logError := renderApi.logError
	layoutsPath := renderApi.layoutsPath
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeParams := &restapi.GetPostsRequestQueryParams{
			Page:         r.FormValue("page"),
			PostsPerPage: "",
		}

		validateQueryParamsError := restapi.ValidateGetPostsRequestQueryParams(rangeParams)
		if validateQueryParamsError != nil {
			restapi.Respond(w, http.StatusNotFound)
			return
		}
		page, _ := strconv.Atoi(rangeParams.Page)

		var tag string
		vars := mux.Vars(r)
		if vars != nil {
			tag = vars["tag"]
		}

		var posts []models.Post
		var err error
		if tag != "" {
			posts, err = postService.GetPostsInRangeByTag(renderApi.db, page, postsPerPage+1, tag)
		} else {
			posts, err = postService.GetPostsInRange(renderApi.db, page, postsPerPage+1)
		}
		if err != nil {
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		tmpl, err := template.New("all-posts").Funcs(renderFuncs).
			ParseFiles(
				layoutsPath+filepath.FromSlash("partials/head.html"),
				layoutsPath+filepath.FromSlash("partials/header.html"),
				layoutsPath+filepath.FromSlash("partials/footer.html"),
				layoutsPath+"all-posts.html")
		if err != nil {
			logError.Printf("Error rendering all posts/tagged page: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		var data Site

		var Title string
		if tag != "" {
			Title = "Posts tagged with " + tag + siteSuffix
		} else {
			Title = "All Posts" + " | Progbloom - A blog about programming"
		}

		data.Head = SiteHead{
			Title: Title,
			Metadata: models.MetaData{
				Description: "blog about programming",
				Keywords:    []string{"Programming", "Linux"},
			},
		}
		data.Desc = defaultSiteDescription

		pageSelector := pageSelector{}
		if page != 0 {
			pageSelector.HasNewerPosts = true
			if tag != "" {
				pageSelector.NewerPostsLink = fmt.Sprintf("/tags/%s?page=%d", tag, page-1)
			} else {
				pageSelector.NewerPostsLink = fmt.Sprintf("/posts?page=%d", page-1)
			}
		} else {
			pageSelector.HasNewerPosts = false
		}

		// hack here: if we were able to retrieve more posts than default value, then we have older posts
		if len(posts) > postsPerPage {
			pageSelector.HasOlderPosts = true
			if tag != "" {
				pageSelector.OlderPostsLink = fmt.Sprintf("/tags/%s?page=%d", tag, page+1)
			} else {
				pageSelector.OlderPostsLink = fmt.Sprintf("/posts?page=%d", page+1)
			}

			// remove very last post, as we need less posts
			posts = posts[:postsPerPage]
		} else {
			pageSelector.HasOlderPosts = false
		}

		allPostsPageData := allPostsPageData{}
		allPostsPageData.Posts = posts
		allPostsPageData.PageSelector = pageSelector
		if tag != "" {
			allPostsPageData.Type = "Tagged"
			allPostsPageData.Tag = tag
		}

		data.Data = allPostsPageData

		if err := tmpl.ExecuteTemplate(w, "all-posts", data); err != nil {
			logError.Printf("Error rendering all posts/tagged page: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
		}
	})
}

// RenderAllTagsPageHandler - handler for server-side rendering of all tags page
func (renderApi *Handler) RenderAllTagsPageHandler() http.Handler {
	logError := renderApi.logError
	layoutsPath := renderApi.layoutsPath
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("all-tags").Funcs(renderFuncs).
			ParseFiles(
				layoutsPath+filepath.FromSlash("partials/head.html"),
				layoutsPath+filepath.FromSlash("partials/header.html"),
				layoutsPath+filepath.FromSlash("partials/footer.html"),
				layoutsPath+"all-tags.html")
		if err != nil {
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		tags, err := tagService.GetAll(renderApi.db)
		if err != nil {
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		var data Site
		data.Head = SiteHead{
			Title: "Tags Cloud" + siteSuffix,
			Metadata: models.MetaData{
				Description: "blog about programming",
				Keywords:    []string{"Programming", "Linux"},
			},
		}
		data.Desc = defaultSiteDescription
		data.Data = struct {
			Tags []string
		}{
			Tags: tags,
		}

		if err := tmpl.ExecuteTemplate(w, "all-tags", data); err != nil {
			logError.Printf("Error rendering all tags page: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
		}
	})
}

//RenderAboutPageHandler - handler for server-side rendering of about page
func (renderApi *Handler) RenderAboutPageHandler() http.Handler {
	logError := renderApi.logError
	layoutsPath := renderApi.layoutsPath
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("about").Funcs(renderFuncs).
			ParseFiles(
				layoutsPath+filepath.FromSlash("partials/head.html"),
				layoutsPath+filepath.FromSlash("partials/header.html"),
				layoutsPath+filepath.FromSlash("partials/footer.html"),
				layoutsPath+"about.html")
		if err != nil {
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		var data Site
		data.Head = SiteHead{
			Title: "All Posts" + siteSuffix,
			Metadata: models.MetaData{
				Description: "blog about programming",
				Keywords:    []string{"Programming", "Linux"},
			},
		}
		data.Desc = defaultSiteDescription
		data.Data = struct {
			Content string
		}{
			Content: "Приветствую на моем сайте!",
		}

		if err := tmpl.ExecuteTemplate(w, "about", data); err != nil {
			logError.Printf("Error rendering about page: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
		}
	})
}

//RenderAdminPageHandler - handler for server-side rendering of admin dashboard page
func (renderApi *Handler) RenderAdminPageHandler() http.Handler {
	logError := renderApi.logError
	layoutsPath := renderApi.layoutsPath
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("about").Funcs(renderFuncs).
			ParseFiles(
				layoutsPath+filepath.FromSlash("partials/head.html"),
				layoutsPath+filepath.FromSlash("partials/header.html"),
				layoutsPath+filepath.FromSlash("partials/footer.html"),
				layoutsPath+"about.html")
		if err != nil {
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		var data Site

		data.Head = SiteHead{
			Title: "Admin Dashboard" + siteSuffix,
			Metadata: models.MetaData{
				Description: "blog about programming",
				Keywords:    []string{"Programming", "Linux"},
			},
		}
		data.Desc = defaultSiteDescription
		data.Data = struct {
			Content string
		}{
			Content: "Приветствую на моем сайте!",
		}

		if err := tmpl.ExecuteTemplate(w, "about", data); err != nil {
			logError.Printf("Error rendering about page: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
		}
	})
}

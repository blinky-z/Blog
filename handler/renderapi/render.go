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
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type Handler struct {
	db          *sql.DB
	admins      *[]string
	layoutsPath string
	domain      *url.URL
	logInfo     *log.Logger
	logError    *log.Logger
}

func NewRenderAPIHandler(db *sql.DB, layoutsPath string, domain *url.URL, logInfo, logError *log.Logger) *Handler {
	return &Handler{
		db:          db,
		layoutsPath: layoutsPath,
		domain:      domain,
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

// Site - represents all site data
type Site struct {
	Head   SiteHead
	Desc   SiteDescription
	Domain *url.URL
	Data   interface{}
}

// pageSelector - represents page selector (older and newer posts links)
type pageSelector struct {
	NewerPostsLink string
	OlderPostsLink string
	HasNewerPosts  bool
	HasOlderPosts  bool
}

// indexPageData - represents index page data
type indexPageData struct {
	Posts []models.Post
}

// postPageData - represents a single post ("/posts/{id}") page data
type postPageData struct {
	Post models.Post
}

// postPageData - represents all posts ("/posts") or all posts tagged with ("tags/{tag}) page data
type allPostsPageData struct {
	Posts        []models.Post
	PageSelector pageSelector
	Type         string
	Tag          string // set if it's the tags/{tag} page
}

// adminEditorPageData - represents data for admin dashboard editor
type adminEditorPageData struct {
	Post        models.Post
	Tags        []string
	PostPresent bool
}

var defaultMetadata = models.MetaData{
	Description: "blog about programming and linux",
	Keywords:    []string{"programming", "coding", "Linux", "Java", "C", "C++", "low-level programming"},
}

var defaultSiteDescription = SiteDescription{
	Title:       "Progbloom 🌻",
	Description: "A blog about programming. I write about Linux, Java and low-level programming",
}

// functions for use in go templates
var renderFuncs = template.FuncMap{
	"formatTime":    formatTime,
	"sliceToString": sliceToString,
}

// formatTime - formats time.Time and returns formatted time as string
func formatTime(t time.Time) string {
	return t.Format(timeFormat)
}

func sliceToString(a []string) string {
	var sb strings.Builder
	aLen := len(a) - 1
	for index, elem := range a {
		sb.WriteString(elem)
		if index < aLen {
			sb.WriteByte(',')
		}
	}
	return sb.String()
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

		post, err := postService.GetByID(renderApi.db, postID)
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
			Title:    post.Title + siteSuffix,
			Metadata: post.Metadata,
		}
		data.Domain = renderApi.domain
		data.Desc = defaultSiteDescription
		data.Data = postPageData{
			Post: post,
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
			renderApi.logInfo.Printf("Error retrieving posts: %s", err)
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
			logError.Printf("Error allocating index template: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		var data Site
		data.Head = SiteHead{
			Title:    "Home" + siteSuffix,
			Metadata: defaultMetadata,
		}
		data.Domain = renderApi.domain
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

// RenderAllPostsPageHandler - handler for server-side rendering of all posts and tagged posts page
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
			Title:    Title,
			Metadata: defaultMetadata,
		}
		data.Domain = renderApi.domain
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

		allTags, err := tagService.GetAll(renderApi.db)
		if err != nil {
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}
		allTagsAsStringSlice := make([]string, len(allTags))
		for tagIndex, tag := range allTags {
			allTagsAsStringSlice[tagIndex] = tag.Name
		}

		var data Site
		data.Head = SiteHead{
			Title:    "Tags Cloud" + siteSuffix,
			Metadata: defaultMetadata,
		}
		data.Domain = renderApi.domain
		data.Desc = defaultSiteDescription
		data.Data = struct {
			Tags []string
		}{
			Tags: allTagsAsStringSlice,
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
			Title:    "About" + siteSuffix,
			Metadata: defaultMetadata,
		}
		data.Domain = renderApi.domain
		data.Desc = defaultSiteDescription
		data.Data = struct {
			Content string
		}{
			Content: `Приветствую на моем сайте! Я пишу о Linux, Java и низкоуровневом программировании.
				<br>
				<hr>
				<b>Мои труды:</b>
				<ul>
			<li><a href="https://habr.com/ru/post/460257/">Hello, World! Глубокое погружение в Терминалы</a></li>
			</ul>`,
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
		tmpl, err := template.New("admin").Funcs(renderFuncs).
			ParseFiles(
				layoutsPath+filepath.FromSlash("partials/head.html"),
				layoutsPath+filepath.FromSlash("partials/header.html"),
				layoutsPath+filepath.FromSlash("partials/footer.html"),
				layoutsPath+"admin.html")
		if err != nil {
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		var data Site

		data.Head = SiteHead{
			Title:    "Admin Dashboard" + siteSuffix,
			Metadata: defaultMetadata,
		}
		data.Domain = renderApi.domain
		data.Desc = defaultSiteDescription
		data.Data = nil

		if err := tmpl.ExecuteTemplate(w, "admin", data); err != nil {
			logError.Printf("Error rendering about page: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
		}
	})
}

//RenderAdminPageHandler - handler for server-side rendering of admin dashboard editor page
func (renderApi *Handler) RenderAdminEditorPageHandler() http.Handler {
	logError := renderApi.logError
	layoutsPath := renderApi.layoutsPath
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adminEditorPageData := adminEditorPageData{}
		postID := r.FormValue("id")
		if postID != "" {
			if !restapi.IsPostIDValid(postID) {
				restapi.Respond(w, http.StatusNotFound)
				return
			}

			post, err := postService.GetByID(renderApi.db, postID)
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
			adminEditorPageData.Post = post
			adminEditorPageData.PostPresent = true
		} else {
			adminEditorPageData.Post = models.Post{}
			adminEditorPageData.PostPresent = false
		}

		allTags, err := tagService.GetAll(renderApi.db)
		if err != nil {
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}
		allTagsAsStringSlice := make([]string, len(allTags))
		for tagIndex, tag := range allTags {
			allTagsAsStringSlice[tagIndex] = tag.Name
		}
		adminEditorPageData.Tags = allTagsAsStringSlice

		tmpl, err := template.New("admin-editor").Funcs(renderFuncs).
			ParseFiles(
				layoutsPath+filepath.FromSlash("partials/head.html"),
				layoutsPath+filepath.FromSlash("partials/header.html"),
				layoutsPath+filepath.FromSlash("partials/footer.html"),
				layoutsPath+"admin/editor.html")
		if err != nil {
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		var data Site

		data.Head = SiteHead{
			Title:    "Admin Dashboard - Editor" + siteSuffix,
			Metadata: defaultMetadata,
		}
		data.Domain = renderApi.domain
		data.Desc = defaultSiteDescription
		data.Data = adminEditorPageData

		if err := tmpl.ExecuteTemplate(w, "admin-editor", data); err != nil {
			logError.Printf("Error rendering admin editor page: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
		}
	})
}

// RenderAllPostsPageHandler - handler for server-side rendering of admin dashboard posts managing page
func (renderApi *Handler) RenderAdminManagePostsPageHandler() http.Handler {
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

		posts, err := postService.GetPostsInRange(renderApi.db, page, postsPerPage+1)
		if err != nil {
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		tmpl, err := template.New("admin-manage-posts").Funcs(renderFuncs).
			ParseFiles(
				layoutsPath+filepath.FromSlash("partials/head.html"),
				layoutsPath+filepath.FromSlash("partials/header.html"),
				layoutsPath+filepath.FromSlash("partials/footer.html"),
				layoutsPath+"admin/manage-posts.html")
		if err != nil {
			logError.Printf("Error allocating admin dashboard posts managing page: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		var data Site

		data.Head = SiteHead{
			Title:    "Admin Dashboard - Manage posts" + " | Progbloom - A blog about programming",
			Metadata: defaultMetadata,
		}
		data.Domain = renderApi.domain
		data.Desc = defaultSiteDescription

		pageSelector := pageSelector{}
		if page != 0 {
			pageSelector.HasNewerPosts = true
			pageSelector.NewerPostsLink = fmt.Sprintf("/manage-posts?page=%d", page-1)
		} else {
			pageSelector.HasNewerPosts = false
		}

		// hack here: if we were able to retrieve more posts than default value, then we have older posts
		if len(posts) > postsPerPage {
			pageSelector.HasOlderPosts = true
			pageSelector.OlderPostsLink = fmt.Sprintf("/manage-posts?page=%d", page+1)

			// remove very last post, as we need less posts
			posts = posts[:postsPerPage]
		} else {
			pageSelector.HasOlderPosts = false
		}

		allPostsPageData := allPostsPageData{}
		allPostsPageData.Posts = posts
		allPostsPageData.PageSelector = pageSelector

		data.Data = allPostsPageData

		if err := tmpl.ExecuteTemplate(w, "admin-manage-posts", data); err != nil {
			logError.Printf("Error rendering admin dashboard posts managing page: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
		}
	})
}

func (renderApi *Handler) RenderAdminManageTagsPageHandler() http.Handler {
	logError := renderApi.logError
	layoutsPath := renderApi.layoutsPath
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("admin-manage-tags").Funcs(renderFuncs).
			ParseFiles(
				layoutsPath+filepath.FromSlash("partials/head.html"),
				layoutsPath+filepath.FromSlash("partials/header.html"),
				layoutsPath+filepath.FromSlash("partials/footer.html"),
				layoutsPath+"admin/manage-tags.html")
		if err != nil {
			logError.Printf("Error allocating admin dashboard tags managing page: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		tags, err := tagService.GetAll(renderApi.db)
		if err != nil {
			logError.Printf("Error retrieving all tags: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
			return
		}

		var data Site
		data.Head = SiteHead{
			Title:    "Admin Dashboard - Manage tags" + " | Progbloom - A blog about programming",
			Metadata: defaultMetadata,
		}
		data.Domain = renderApi.domain
		data.Desc = defaultSiteDescription
		data.Data = struct {
			Tags []models.Tag
		}{
			Tags: tags,
		}

		if err := tmpl.ExecuteTemplate(w, "admin-manage-tags", data); err != nil {
			logError.Printf("Error rendering admin dashboard tags managing page: %s", err)
			restapi.Respond(w, http.StatusInternalServerError)
		}
	})
}

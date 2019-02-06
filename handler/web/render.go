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
	models.MetaData
	PostsList
	PageSelector
}

// CommentWithChilds - represents comment in comments section
type CommentWithChilds struct {
	CommentID      string
	Username       string
	CreationTime   string
	CommentContent string
	Childs         []CommentWithChilds
}

// PostCommentsList - represents comments section below post
type PostCommentsList struct {
	Comments []CommentWithChilds
}

// BlogPost - represents blog post on index and /posts/{id} pages
type BlogPost struct {
	PostLink         string
	PostTitle        string
	PostAuthor       string
	PostCreationTime string
	PostSnippet      string
	PostContent      string
}

// PostPage - represents /posts/{id} page
type PostPage struct {
	models.MetaData
	BlogPost
	CommentsCount int
	PostCommentsList
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

		env.LogInfo.Printf("Getting post page template")
		postTemplate, err :=
			template.ParseFiles(templatesFolder+"header.html", templatesFolder+"comment.html",
				templatesFolder+"postPage.html", templatesFolder+"footer.html")
		if err != nil {
			env.LogError.Print(err)
			api.Respond(w, http.StatusInternalServerError)
			return
		}

		env.LogInfo.Printf("Setting post page template data")

		var data PostPage

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
		// TODO: написать это за O(N) или хотя бы за O(N^2)
		// идея: сначала сохраним все комменты в мапу, где ключ будет ID коммента
		// когда мы встречаем коммент, у которого есть родитель, мы находим в мапе этот коммент и добавляем к списку
		// вложенных комментов этот коммент
		// вложенные комменты в мапе не сохраняем
		commentWithChildsAsMap := make(map[string]CommentWithChilds)
		var parentComments []string

		// на этом шаге мы поместили в мапу абсолютно все комменты
		// каждый коммент может вмещать в себя еще комментарии
		// теперь, мы имеем ID коммента и сам коммент по этому ID в мапе commentWithChildsAsMap
		for _, comment := range comments {
			var commentWithChilds CommentWithChilds
			commentWithChilds.CommentID = comment.ID
			commentWithChilds.CommentContent = comment.Content
			commentWithChilds.CreationTime = comment.Date.Format(timeFormat)
			commentWithChilds.Username = comment.Author

			commentWithChildsAsMap[comment.ID] = commentWithChilds

			if !comment.ParentID.Valid {
				parentComments = append(parentComments, comment.ID)
			}
		}

		// теперь у нас есть мапа со всеми комментами
		// теперь заполним всех детей у всех комментов
		// при этом, теперь в мапе будут лежать как родители со всеми детьми, так и дети
		// у детей также есть дети
		for _, comment := range comments {
			if comment.ParentID.Valid {
				// достанем родителя текущего ребенка
				parent := commentWithChildsAsMap[comment.ParentID.Value()]
				// достанем детей текущего родителя
				// добавим текущий коммент как ребенка к остальным детям родителя
				parent.Childs = append(parent.Childs, commentWithChildsAsMap[comment.ID])
				// обновим запись в мапе
				commentWithChildsAsMap[comment.ParentID.Value()] = parent
			}
		}

		// TODO
		// ошибка: здесь мы просто берем детей родителей, у которых нет родителей, т.е. мы берем комменты первого уровня
		// и ложим их вместе с их детьми в слайс
		// но ведь нам нужно, чтобы не только дети комментариев первого уровня были, но и также дети детей
		var parentCommentWithChilds []CommentWithChilds
		for _, parentCommendID := range parentComments {
			parentCommentWithChilds = append(parentCommentWithChilds, commentWithChildsAsMap[parentCommendID])
		}

		fmt.Printf("\n%+v\n", parentComments)
		fmt.Printf("\n%+v\n", commentWithChildsAsMap)
		fmt.Printf("\n%+v\n", parentCommentWithChilds)
		data.Comments = parentCommentWithChilds

		data.CommentsCount = len(data.Comments)

		env.LogInfo.Printf("Executing post template")
		if err := postTemplate.ExecuteTemplate(w, "postPage", data); err != nil {
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

		var data IndexPage

		for currentPostNum := 0; currentPostNum < len(posts) && currentPostNum < 10; currentPostNum++ {
			var blogPostData BlogPost
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

package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/blinky-z/Blog/server/models"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type getPostsRangeParams struct {
	page         int
	postsPerPage int
}

const (
	// InvalidTitle - incorrect user input - invalid title of post
	InvalidTitle PostErrorCode = "INVALID_TITLE"
	// InvalidID - incorrect user input - invalid id of post
	InvalidID PostErrorCode = "INVALID_ID"
	// InvalidContent - incorrect user input - invalid content of post
	InvalidContent PostErrorCode = "INVALID_CONTENT"
	// BadRequestBody - incorrect user post - invalid json post
	BadRequestBody PostErrorCode = "BAD_BODY"
	// NoSuchPost - incorrect user input - requested post does not exist in database
	NoSuchPost PostErrorCode = "NO_SUCH_POST"
	// InvalidRange - user inputs invalid range of posts to get from database
	InvalidRange PostErrorCode = "INVALID_POSTS_RANGE"
	// NoPermissions - user doesn't permissions to create/update/delete post
	NoPermissions PostErrorCode = "NO_PERMISSIONS"

	// MaxPostTitleLen - maximum length of post title
	MaxPostTitleLen int = 120

	// MaxPostsPerPage - maximum posts can be displayed on one page
	MaxPostsPerPage int = 40

	defaultMaxPostsPerPage string = "10"
)

func validateGetPostsParams(r *http.Request) (params getPostsRangeParams, validateError PostErrorCode) {
	validateError = NoError

	var page int
	var postsPerPage int
	var err error

	if len(r.FormValue("page")) == 0 {
		validateError = InvalidRange
		return
	}

	if page, err = strconv.Atoi(r.FormValue("page")); err != nil || page < 0 {
		validateError = InvalidRange
		return
	}

	postsPerPageAsString := r.FormValue("posts-per-page")
	if len(postsPerPageAsString) == 0 {
		postsPerPageAsString = defaultMaxPostsPerPage
	}

	postsPerPage, err = strconv.Atoi(postsPerPageAsString)
	if err != nil {
		validateError = InvalidRange
		return
	}

	if postsPerPage < 0 || postsPerPage > MaxPostsPerPage {
		validateError = InvalidRange
		return
	}

	params.page = page
	params.postsPerPage = postsPerPage

	return
}

func validatePost(r *http.Request) (post models.Post, validateError PostErrorCode) {
	validateError = NoError
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		validateError = BadRequestBody
		return
	}

	if len(post.Title) > MaxPostTitleLen || len(post.Title) == 0 {
		validateError = InvalidTitle
		return
	}

	if len(post.Content) == 0 {
		validateError = InvalidContent
		return
	}

	return
}

func validatePostID(r *http.Request) (id string, validateError PostErrorCode) {
	validateError = NoError
	vars := mux.Vars(r)

	if _, err := strconv.Atoi(vars["id"]); err != nil {
		validateError = InvalidID
		return
	}

	id = vars["id"]
	return
}

// CreatePost - create post http handler
func CreatePost(w http.ResponseWriter, r *http.Request) {
	LogInfo.Print("Got new Post CREATE job")

	userRole := r.Context().Value(ctxKey).(string)
	if userRole != "admin" {
		LogInfo.Printf("User with role %s doesn't have permissions to CREATE post", userRole)
		respondWithError(w, http.StatusForbidden, NoPermissions)
		return
	}

	post, validatePostError := validatePost(r)
	if validatePostError != NoError {
		LogInfo.Print("Can not create post: post is invalid")
		respondWithError(w, http.StatusBadRequest, validatePostError)
		return
	}

	var createdPost models.Post

	LogInfo.Printf("Inserting post with Title %s into database", post.Title)

	if _, err := Db.Exec("BEGIN TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if err := Db.QueryRow("insert into posts(title, content) values($1, $2) RETURNING id, title, date, content",
		post.Title, post.Content).
		Scan(&createdPost.ID, &createdPost.Title, &createdPost.Date, &createdPost.Content); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if _, err := Db.Exec("END TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Post with Title %s successfully created", createdPost.Title)

	respondWithBody(w, http.StatusCreated, createdPost)
}

// UpdatePost - update post http handler
func UpdatePost(w http.ResponseWriter, r *http.Request) {
	LogInfo.Print("Got new Post UPDATE job")

	userRole := r.Context().Value(ctxKey).(string)
	if userRole != "admin" {
		LogInfo.Printf("User with role %s doesn't have permissions to UPDATE post", userRole)
		respondWithError(w, http.StatusForbidden, NoPermissions)
		return
	}

	id, validateIDError := validatePostID(r)
	if validateIDError != NoError {
		LogInfo.Print("Can not UPDATE post: ID of Post to update is invalid")
		respondWithError(w, http.StatusBadRequest, validateIDError)
		return
	}

	post, validatePostError := validatePost(r)
	if validatePostError != NoError {
		LogInfo.Printf("Can not UPDATE post with ID %s : New Post is invalid", id)
		respondWithError(w, http.StatusBadRequest, validatePostError)
		return
	}

	if err := Db.QueryRow("select from posts where id = $1", id).Scan(); err != nil {
		if err == sql.ErrNoRows {
			LogInfo.Printf("Can not UPDATE post with ID %s : post does not exist", id)
			respondWithError(w, http.StatusNotFound, NoSuchPost)
			return
		}

		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	var updatedPost models.Post

	LogInfo.Printf("Updating post with ID %s in database", id)

	if _, err := Db.Exec("BEGIN TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if err := Db.QueryRow("UPDATE posts SET title = $1, content = $2 WHERE id = $3 RETURNING id, title, date, content",
		post.Title, post.Content, id).
		Scan(&updatedPost.ID, &updatedPost.Title, &updatedPost.Date, &updatedPost.Content); err != nil {
		if err != nil {
			LogError.Print(err)
			respondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}
	}
	if _, err := Db.Exec("END TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Post with ID %s successfully updated", id)

	respondWithBody(w, http.StatusOK, updatedPost)
}

// DeletePost - delete post http handler
func DeletePost(w http.ResponseWriter, r *http.Request) {
	LogInfo.Print("Got new Post DELETE job")

	userRole := r.Context().Value(ctxKey).(string)
	if userRole != "admin" {
		LogInfo.Printf("User with role %s doesn't have permissions to DELETE post", userRole)
		respondWithError(w, http.StatusForbidden, NoPermissions)
		return
	}

	id, validateIDError := validatePostID(r)
	if validateIDError != NoError {
		LogInfo.Print("Can not DELETE post: post ID is invalid")
		respondWithError(w, http.StatusBadRequest, validateIDError)
		return
	}

	LogInfo.Printf("Deleting post with ID %s from database", id)

	if _, err := Db.Exec("BEGIN TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if _, err := Db.Exec("DELETE FROM posts WHERE id = $1", id); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if _, err := Db.Exec("END TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Post with ID %s successfully deleted", id)

	respond(w, http.StatusOK)
}

// GetCertainPost - get single post from database http handler
func GetCertainPost(w http.ResponseWriter, r *http.Request) {
	LogInfo.Print("Got new Post GET job")

	var post models.Post

	id, validateIDError := validatePostID(r)
	if validateIDError != NoError {
		LogInfo.Print("Can not GET post: post ID is invalid")
		respondWithError(w, http.StatusBadRequest, validateIDError)
		return
	}

	LogInfo.Printf("Getting post with ID %s from database", id)

	if err := Db.QueryRow("select id, title, date, content from posts where id = $1", id).
		Scan(&post.ID, &post.Title, &post.Date, &post.Content); err != nil {
		if err == sql.ErrNoRows {
			LogInfo.Printf("Can not GET post with ID %s : post does not exist", id)
			respondWithError(w, http.StatusNotFound, NoSuchPost)
			return
		}

		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Post with ID %s succesfully arrived from database", id)

	respondWithBody(w, 200, post)
}

// GetPosts - get one page of posts from database http handler
func GetPosts(w http.ResponseWriter, r *http.Request) {
	LogInfo.Print("Got new Range of Posts GET job")

	params, validateError := validateGetPostsParams(r)
	if validateError != NoError {
		LogInfo.Print("Can not GET range of posts : get posts Query params are invalid")
		respondWithError(w, http.StatusBadRequest, validateError)
		return
	}

	page := params.page
	postsPerPage := params.postsPerPage

	var posts []models.Post

	LogInfo.Printf("Getting Range of Posts with following params: (page: %d, posts per page: %d) from database",
		page, postsPerPage)

	rows, err := Db.Query("select id, title, date, content from posts order by id DESC offset $1 limit $2",
		page*postsPerPage, postsPerPage)
	if err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	for rows.Next() {
		var currentPost models.Post
		if err = rows.Scan(&currentPost.ID, &currentPost.Title, &currentPost.Date, &currentPost.Content); err != nil {
			LogError.Print(err)
			respondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}
		posts = append(posts, currentPost)
	}

	LogInfo.Print("Range of Posts successfully arrived from database")

	respondWithBody(w, 200, posts)
}

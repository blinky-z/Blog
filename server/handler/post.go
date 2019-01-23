package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/blinky-z/Blog/server/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// LogInfo - log for writing+ messages
	LogInfo *log.Logger
	// LogError - log for writing server errors
	LogError *log.Logger

	// Db - database connection. This variable is set by main function
	Db *sql.DB

	// SigningKey - secret key for creating token
	SigningKey string
)

// Response - behaves like Either Monad
// 'Error' field is set while error occurred.
// Otherwise 'Body' field is used to return post from database
type Response struct {
	Error PostErrorCode `json:"error"`
	Body  interface{}   `json:"body"`
}

type getPostsRangeParams struct {
	page         int
	postsPerPage int
}

type userCredentials struct {
	login    string
	email    string
	password string
}

// PostErrorCode - represents error occurred while handling request
type PostErrorCode string

const (
	// TechnicalError - server error
	TechnicalError PostErrorCode = "TECHNICAL_ERROR"
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
	// WrongCredentials - user inputs wrong password or login or email while logging in
	WrongCredentials PostErrorCode = "WRONG_CREDENTIALS"
	// InvalidEmail - user inputs invalid email while registration or logging in
	InvalidEmail PostErrorCode = "INVALID_EMAIL"
	// InvalidLogin - user inputs invalid login while registration
	InvalidLogin PostErrorCode = "INVALID_LOGIN"
	// InvalidPassword - user inputs invalid password while registration
	InvalidPassword PostErrorCode = "INVALID_PASSWORD"
	// NoError - no error occurred while handling request
	NoError PostErrorCode = ""

	maxPostTitleLen int = 120

	maxPostsPerPage int = 40

	defaultMaxPostsPerPage string = "10"

	minPwdLen int = 8
	maxPwdLen int = 40

	minLoginLen int = 6
	maxLoginLen int = 20

	postFields = "id, title, date, content"
)

func respond(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

func respondWithJSON(w http.ResponseWriter, code int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_, err := w.Write(body)
	if err != nil {
		LogError.Print(err)
	}
}

func respondWithError(w http.ResponseWriter, code int, errorCode PostErrorCode) {
	var response Response
	response.Error = errorCode
	encodedResponse, _ := json.Marshal(response)

	respondWithJSON(w, code, encodedResponse)
}

func respondWithBody(w http.ResponseWriter, code int, payload interface{}) {
	var response Response
	response.Body = payload
	encodedResponse, _ := json.Marshal(response)

	respondWithJSON(w, code, encodedResponse)
}

func validateUserGetPostsParams(r *http.Request) (params getPostsRangeParams, validateError PostErrorCode) {
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

	if postsPerPage < 0 || postsPerPage > maxPostsPerPage {
		validateError = InvalidRange
		return
	}

	params.page = page
	params.postsPerPage = postsPerPage

	return
}

func validateUserRegistrationCredentials(r *http.Request) (credentials userCredentials, validateError PostErrorCode) {
	validateError = NoError

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		if err != nil {
			validateError = BadRequestBody
			return
		}
	}

	login := credentials.login
	password := credentials.password
	email := credentials.email

	if len(login) == 0 || len(password) == 0 || len(email) == 0 {
		validateError = BadRequestBody
		return
	}

	if !checkEmail(email) {
		validateError = InvalidEmail
		return
	}

	pwdLen := len(password)
	if pwdLen < minPwdLen || pwdLen > maxPwdLen || !strings.Contains(password, "[A-Z]") {
		validateError = InvalidPassword
		return
	}

	loginLen := len(login)
	if loginLen < minLoginLen || loginLen > maxLoginLen {
		validateError = InvalidLogin
		return
	}

	return
}

func validateUserLoginCredentials(r *http.Request) (credentials userCredentials, validateError PostErrorCode) {
	validateError = NoError

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		if err != nil {
			validateError = BadRequestBody
			return
		}
	}

	login := credentials.login
	password := credentials.password
	email := credentials.email

	if (len(login) == 0 && len(email) == 0) || len(password) == 0 {
		validateError = BadRequestBody
		return
	}

	if len(email) != 0 && !checkEmail(email) {
		validateError = InvalidEmail
		return
	}

	return
}

func validateUserPost(r *http.Request) (post models.Post, validateError PostErrorCode) {
	validateError = NoError
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		validateError = BadRequestBody
		return
	}

	if len(post.Title) > maxPostTitleLen || len(post.Title) == 0 {
		validateError = InvalidTitle
		return
	}

	if len(post.Content) == 0 {
		validateError = InvalidContent
		return
	}

	return
}

func checkEmail(email string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?" +
		"(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	return re.MatchString(email)
}

func validateUserID(r *http.Request) (id string, validateError PostErrorCode) {
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
	post, validatePostError := validateUserPost(r)
	if validatePostError != NoError {
		respondWithError(w, http.StatusBadRequest, validatePostError)
		return
	}

	LogInfo.Printf("Got new post creation job. New post: %v", post)

	var createdPost models.Post

	if _, err := Db.Exec("BEGIN TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if err := Db.QueryRow("insert into posts(title, content) values($1, $2) RETURNING $3",
		post.Title, post.Content, postFields).
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

	LogInfo.Printf("Created new post: %v", createdPost)

	respondWithBody(w, http.StatusCreated, createdPost)
}

// UpdatePost - update post http handler
func UpdatePost(w http.ResponseWriter, r *http.Request) {
	id, validateIDError := validateUserID(r)
	if validateIDError != NoError {
		respondWithBody(w, http.StatusBadRequest, validateIDError)
		return
	}

	post, validatePostError := validateUserPost(r)
	if validatePostError != NoError {
		respondWithError(w, http.StatusBadRequest, validatePostError)
		return
	}

	LogInfo.Printf("Got new post update job. New post: %v", post)

	if err := Db.QueryRow("select from posts where id = $1", id).Scan(); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, NoSuchPost)
			return
		}

		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	var updatedPost models.Post

	if _, err := Db.Exec("BEGIN TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if err := Db.QueryRow("UPDATE posts SET title = $1, content = $2 WHERE id = $3 RETURNING id, title, date, content",
		post.Title, post.Content, id).Scan(&updatedPost.ID, &updatedPost.Title, &updatedPost.Date, &updatedPost.Content); err != nil {
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

	LogInfo.Printf("Updated post: %v", updatedPost)

	respondWithBody(w, http.StatusOK, updatedPost)
}

// DeletePost - delete post http handler
func DeletePost(w http.ResponseWriter, r *http.Request) {
	id, validateIDError := validateUserID(r)
	if validateIDError != NoError {
		respondWithError(w, http.StatusBadRequest, validateIDError)
		return
	}

	LogInfo.Printf("Got new post deletion job. Post id: %s", id)

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

	LogInfo.Printf("Post with id %s deleted", id)

	respond(w, http.StatusOK)
}

// GetCertainPost - get single post from database http handler
func GetCertainPost(w http.ResponseWriter, r *http.Request) {
	var response Response

	var post models.Post

	id, validateIDError := validateUserID(r)
	if validateIDError != NoError {
		respondWithError(w, http.StatusBadRequest, validateIDError)
		return
	}

	LogInfo.Printf("Got new get certain post job. Post id: %s", id)

	if err := Db.QueryRow("select $1 from posts where id = $2", postFields, id).Scan(
		&post.ID, &post.Title, &post.Date, &post.Content); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusBadRequest, NoSuchPost)
			return
		}

		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Post with id %s arrived from database", id)

	response.Body = post
	respondWithBody(w, 200, response)
}

// GetPosts - get one page of posts from database http handler
func GetPosts(w http.ResponseWriter, r *http.Request) {
	var response Response

	params, validateError := validateUserGetPostsParams(r)
	if validateError != NoError {
		respondWithError(w, http.StatusBadRequest, validateError)
		return
	}

	page := params.page
	postsPerPage := params.postsPerPage

	var posts []models.Post

	LogInfo.Printf("Got new get range of posts job. Page: %d. Posts per page: %d", page, postsPerPage)

	rows, err := Db.Query("select $1 from posts order by id DESC offset $2 limit $3",
		postFields, page*postsPerPage, postsPerPage)
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

	LogInfo.Print("Posts arrived from database")

	response.Body = posts
	respondWithBody(w, 200, response)
}

// RegisterUserHandler - checks user registration credentials and inserts login and password into database
func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	credentials, validateError := validateUserRegistrationCredentials(r)
	if validateError != NoError {
		respondWithError(w, http.StatusBadRequest, validateError)
		return
	}

	login := credentials.login
	email := credentials.email
	password := []byte(credentials.password)

	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	if _, err = Db.Exec("BEGIN TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	_, err = Db.Exec("insert into users (login, email, password) values($1, $2, $3)",
		login, email, string(hashedPassword))
	if err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if _, err := Db.Exec("END TRANSACTION"); err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	respond(w, http.StatusOK)
}

func getToken(credentials userCredentials) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["name"] = credentials.login
	claims["exp"] = time.Now().Add(1 * time.Hour).Unix()

	tokenString, err := token.SignedString(SigningKey)

	return tokenString, err
}

// LoginUserHandler - checks user credentials and returns authorization token
func LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	credentials, validateError := validateUserLoginCredentials(r)
	if validateError != NoError {
		respondWithError(w, http.StatusBadRequest, validateError)
		return
	}

	login := credentials.login
	email := credentials.email
	password := credentials.password

	var hashedPassword string

	var identifierField string
	var identifierFieldValue string
	if len(email) != 0 {
		identifierField = "email"
		identifierFieldValue = email
	} else {
		identifierField = "login"
		identifierFieldValue = login
	}

	if err := Db.QueryRow("select password from users where $1 = $2", identifierField, identifierFieldValue).
		Scan(&hashedPassword); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, WrongCredentials)
			return
		}
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, WrongCredentials)
		return
	}

	token, err := getToken(credentials)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	respondWithBody(w, http.StatusAccepted, token)
}

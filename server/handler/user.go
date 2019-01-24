package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/blinky-z/Blog/server/models"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

var (
	// SigningKey - secret key for creating token
	SigningKey []byte
)

const (
	// WrongCredentials - user inputs wrong password or login or email while logging in
	WrongCredentials PostErrorCode = "WRONG_CREDENTIALS"
	// InvalidEmail - user inputs invalid email while registration or logging in
	InvalidEmail PostErrorCode = "INVALID_EMAIL"
	// InvalidLogin - user inputs invalid login while registration
	InvalidLogin PostErrorCode = "INVALID_LOGIN"
	// InvalidPassword - user inputs invalid password while registration
	InvalidPassword PostErrorCode = "INVALID_PASSWORD"
	// AlreadyRegistered - user trying to register account while already registered
	AlreadyRegistered PostErrorCode = "ALREADY_REGISTERED"
	// IncompleteCredentials - user do not input full credentials: login, email, password
	IncompleteCredentials PostErrorCode = "INCOMPLETE_CREDENTIALS"

	// MinPwdLen - minimum length of user password
	MinPwdLen int = 8
	// MaxPwdLen - maximum length of user password
	MaxPwdLen int = 38

	// MinLoginLen - minimum length of user login
	MinLoginLen int = 6
	// MaxLoginLen - maximum length of user login
	MaxLoginLen int = 36

	// MaxEmailLen - maximum length of email
	MaxEmailLen int = 255
)

func validateUserRegistrationCredentials(r *http.Request) (credentials models.User, validateError PostErrorCode) {
	validateError = NoError

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		if err != nil {
			validateError = BadRequestBody
			return
		}
	}

	login := credentials.Login
	password := credentials.Password
	email := credentials.Email

	if len(login) == 0 || len(email) == 0 || len(password) == 0 {
		validateError = IncompleteCredentials
		return
	}

	if !checkEmail(email) {
		validateError = InvalidEmail
		return
	}

	loginLen := len(login)
	if loginLen < MinLoginLen || loginLen > MaxLoginLen {
		validateError = InvalidLogin
		return
	}

	passwordLen := len(password)
	if passwordLen < MinPwdLen || passwordLen > MaxPwdLen || password == login {
		validateError = InvalidPassword
		return
	}

	return
}

func validateUserLoginCredentials(r *http.Request) (credentials models.User, validateError PostErrorCode) {
	validateError = NoError

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		if err != nil {
			validateError = BadRequestBody
			return
		}
	}

	login := credentials.Login
	password := credentials.Password
	email := credentials.Email

	if (len(login) == 0 && len(email) == 0) || len(password) == 0 {
		validateError = IncompleteCredentials
		return
	}

	if len(email) != 0 && !checkEmail(email) {
		validateError = InvalidEmail
		return
	}

	return
}

func checkEmail(email string) bool {
	return strings.Count(email, "@") == 1 && len(email) <= MaxEmailLen && email[0] != '@' && email[len(email)-1] != '@'
}

// RegisterUserHandler - checks user registration credentials and inserts login and password into database
func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	LogInfo.Print("Got new user Registration job")

	credentials, validateError := validateUserRegistrationCredentials(r)
	if validateError != NoError {
		LogInfo.Print("User Registration credentials are invalid")
		respondWithError(w, http.StatusBadRequest, validateError)
		return
	}

	login := credentials.Login
	email := credentials.Email
	password := []byte(credentials.Password)

	var userExists bool
	if err := Db.QueryRow("select exists(select from users where email = $1 or login = $2)", email, login).
		Scan(&userExists); err != nil && err != sql.ErrNoRows {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	if userExists {
		LogInfo.Printf("User with following credentials: (login: %s; email: %s) already registered", login, email)
		respondWithError(w, http.StatusBadRequest, AlreadyRegistered)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Inserting credentials of user with following credentials: (login: %s; email: %s) into database",
		login, email)

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

	LogInfo.Printf("User with following credentials: (login: %s; email: %s) successfully registered", login, email)

	respond(w, http.StatusOK)
}

func getToken(credentials models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["name"] = credentials.Login
	claims["exp"] = time.Now().Add(1 * time.Hour).Unix()

	tokenString, err := token.SignedString(SigningKey)

	return tokenString, err
}

// LoginUserHandler - checks user credentials and returns authorization token
func LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	LogInfo.Printf("Got new user Log In job")

	credentials, validateError := validateUserLoginCredentials(r)
	if validateError != NoError {
		LogInfo.Print("User Log In credentials are invalid")
		respondWithError(w, http.StatusBadRequest, validateError)
		return
	}

	login := credentials.Login
	email := credentials.Email
	password := credentials.Password

	var hashedPassword string

	LogInfo.Printf("Getting hashed password from database of user with following credentials: (login: %s; email: %s)",
		login, email)

	var err error
	if len(email) != 0 {
		err = Db.QueryRow("select password from users where email = $1", email).
			Scan(&hashedPassword)
	} else {
		err = Db.QueryRow("select password from users where login = $1", login).
			Scan(&hashedPassword)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			LogInfo.Printf("Can not get hashed password: user with following credentials: (login: %s; email: %s) "+
				"does not exist", login, email)
			respondWithError(w, http.StatusUnauthorized, WrongCredentials)
			return
		}
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		LogInfo.Printf("Inputted password by user with following credentials: (login: %s; email: %s) "+
			"does not match hashed one",
			login, email)
		respondWithError(w, http.StatusUnauthorized, WrongCredentials)
		return
	}

	token, err := getToken(credentials)
	if err != nil {
		LogError.Print(err)
		respondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Token successfully generated. Sending token to user with following credentials: "+
		"(login: %s; email: %s)", login, email)

	respondWithBody(w, http.StatusAccepted, token)
}

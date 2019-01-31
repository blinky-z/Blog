package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"github.com/blinky-z/Blog/server/models"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

// ctxRoleKey - type that represents context key type for getting user's role
type ctxRoleKey string

var (
	// SigningKey - secret key for creating token
	SigningKey []byte

	// Admins - List of admins that own permissions to create, update, delete posts
	Admins []models.User

	// CtxKey - context key for getting user's role
	CtxKey = ctxRoleKey("role")
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
	// InvalidToken - user provided non-authentic or malformed token
	InvalidToken PostErrorCode = "INVALID_TOKEN"

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

	loginLen := len(login)
	if loginLen != 0 && (loginLen < MinLoginLen || loginLen > MaxLoginLen) {
		validateError = InvalidLogin
		return
	}

	passwordLen := len(password)
	if passwordLen < MinPwdLen || passwordLen > MaxPwdLen {
		validateError = InvalidPassword
		return
	}

	return
}

func checkEmail(email string) bool {
	return strings.Count(email, "@") == 1 && len(email) <= MaxEmailLen && email[0] != '@' && email[len(email)-1] != '@'
}

// JwtAuthentication - middleware for checking JWT tokens
var JwtAuthentication = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LogInfo.Printf("Checking fingerprint")

		token := r.Context().Value("user").(*jwt.Token)

		fingerprintCookie, err := r.Cookie("Secure-Fgp")
		if err != nil {
			LogInfo.Printf("Request missing fingerprint")
			RespondWithError(w, http.StatusUnauthorized, InvalidToken)
			return
		}
		rawFingerprint := fingerprintCookie.Value

		claims := token.Claims.(jwt.MapClaims)
		tokenFingerprint := claims["fingerprint"].(string)

		if err = bcrypt.CompareHashAndPassword([]byte(tokenFingerprint), []byte(rawFingerprint)); err != nil {
			LogInfo.Printf("Error checking fingeprint: raw fingerprint does not match fingeprint containing in JWT Token")
			RespondWithError(w, http.StatusUnauthorized, InvalidToken)
			return
		}

		role := claims["role"].(string)

		LogInfo.Printf("Fingerprint is valid. Serving next http handler")

		ctx := context.WithValue(r.Context(), CtxKey, role)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// RegisterUserHandler - checks user registration credentials and inserts login and password into database
func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	LogInfo.Print("Got new user Registration job")

	credentials, validateError := validateUserRegistrationCredentials(r)
	if validateError != NoError {
		LogInfo.Print("User Registration credentials are invalid")
		RespondWithError(w, http.StatusBadRequest, validateError)
		return
	}

	login := credentials.Login
	email := credentials.Email
	password := []byte(credentials.Password)

	var userExists bool
	if err := Db.QueryRow("select exists(select from users where email = $1 or login = $2)", email, login).
		Scan(&userExists); err != nil {
		LogError.Print(err)
		RespondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	if userExists {
		LogInfo.Printf("User with following credentials: (login: %s; email: %s) already registered", login, email)
		RespondWithError(w, http.StatusBadRequest, AlreadyRegistered)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		LogError.Print(err)
		RespondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Inserting credentials of user with following credentials: (login: %s; email: %s) into database",
		login, email)

	if _, err = Db.Exec("BEGIN TRANSACTION"); err != nil {
		LogError.Print(err)
		RespondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	_, err = Db.Exec("insert into users (login, email, password) values($1, $2, $3)",
		login, email, string(hashedPassword))
	if err != nil {
		LogError.Print(err)
		RespondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}
	if _, err := Db.Exec("END TRANSACTION"); err != nil {
		LogError.Print(err)
		RespondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("User with following credentials: (login: %s; email: %s) successfully registered", login, email)

	respond(w, http.StatusOK)
}

func generateRandomContext() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

func isUserAdmin(login string) bool {
	for _, currentAdmin := range Admins {
		if currentAdmin.Login == login {
			return true
		}
	}
	return false
}

func getToken(login, ctx string) (string, error) {
	var claims models.TokenClaims
	claims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
	claims.Fingerprint = ctx

	if isUserAdmin(login) {
		claims.Role = "admin"
	} else {
		claims.Role = "user"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(SigningKey)

	return tokenString, err
}

// LoginUserHandler - checks user credentials and returns authorization token
func LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	LogInfo.Printf("Got new user Log In job")

	credentials, validateError := validateUserLoginCredentials(r)
	if validateError != NoError {
		LogInfo.Print("User Log In credentials are invalid")
		RespondWithError(w, http.StatusBadRequest, validateError)
		return
	}

	login := credentials.Login
	email := credentials.Email
	password := credentials.Password

	var hashedPassword string

	LogInfo.Printf("Getting hashed password from database of user with following credentials: (login: %s; email: %s)",
		login, email)

	var err error
	if len(login) == 0 {
		err = Db.QueryRow("select login, password from users where email = $1", email).Scan(&login, &hashedPassword)
	} else {
		err = Db.QueryRow("select password from users where login = $1", login).Scan(&hashedPassword)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			LogInfo.Printf("Can not get hashed password: user with following credentials: (login: %s; email: %s) "+
				"does not exist", login, email)
			RespondWithError(w, http.StatusUnauthorized, WrongCredentials)
			return
		}
		LogError.Print(err)
		RespondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		LogInfo.Printf("Inputted password by user %s does not match hashed one", login)
		RespondWithError(w, http.StatusUnauthorized, WrongCredentials)
		return
	}

	LogInfo.Printf("Generating hashed fingerprint")

	ctx, err := generateRandomContext()
	if err != nil {
		LogError.Print(err)
		RespondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	// set raw fingerprint in cookie
	ctxCookie := &http.Cookie{Name: "Secure-Fgp", Value: ctx, SameSite: http.SameSiteStrictMode, HttpOnly: true,
		Expires: time.Now().Add(time.Hour * 1), Path: "/"}
	http.SetCookie(w, ctxCookie)

	// generate hashed fingerprint
	hashedCtx, err := bcrypt.GenerateFromPassword([]byte(ctx), bcrypt.DefaultCost)
	if err != nil {
		LogError.Print(err)
		RespondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Generating JWT Token with hashed fingerprint")

	token, err := getToken(login, string(hashedCtx))
	if err != nil {
		LogError.Print(err)
		RespondWithError(w, http.StatusInternalServerError, TechnicalError)
		return
	}

	LogInfo.Printf("Token successfully generated. Sending token to user %s", login)

	RespondWithBody(w, http.StatusAccepted, token)
}

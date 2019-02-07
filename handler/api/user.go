package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/settings"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

// ctxRoleKey - type that represents context key type for getting user's role
type ctxRoleKey string

var (
	// CtxKey - context key for getting user's role
	CtxKey = ctxRoleKey("role")
)

// UserAPI - environment container struct to declare all auth handlers as methods
type UserAPI struct {
	Env        *models.Env
	SigningKey []byte
	Admins     []settings.Admin
}

var (
	// UserEnv - instance of UserAPI struct. Initialized by main
	UserEnv UserAPI
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

func checkEmail(email string) bool {
	return strings.Count(email, "@") == 1 && len(email) <= MaxEmailLen && email[0] != '@' && email[len(email)-1] != '@'
}

func validateUserRegistrationCredentials(r *http.Request) (
	credentials models.RegistrationRequest, validateError PostErrorCode) {
	validateError = NoError

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		if err != nil {
			validateError = BadRequestBody
			return
		}
	}

	username := credentials.Username
	password := credentials.Password
	email := credentials.Email

	if len(username) == 0 || len(email) == 0 || len(password) == 0 {
		validateError = IncompleteCredentials
		return
	}

	if !checkEmail(email) {
		validateError = InvalidEmail
		return
	}

	loginLen := len(username)
	if loginLen < MinLoginLen || loginLen > MaxLoginLen {
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

func validateUserLoginCredentials(r *http.Request) (credentials models.LoginRequest, validateError PostErrorCode) {
	validateError = NoError

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		if err != nil {
			validateError = BadRequestBody
			return
		}
	}

	username := credentials.Username
	email := credentials.Email
	password := credentials.Password

	if (len(username) == 0 && len(email) == 0) || len(password) == 0 {
		validateError = IncompleteCredentials
		return
	}

	if len(email) != 0 {
		if !checkEmail(email) {
			validateError = InvalidEmail
			return
		}
	} else {
		loginLen := len(username)
		if loginLen != 0 && (loginLen < MinLoginLen || loginLen > MaxLoginLen) {
			validateError = InvalidLogin
			return
		}
	}

	passwordLen := len(password)
	if passwordLen < MinPwdLen || passwordLen > MaxPwdLen {
		validateError = InvalidPassword
		return
	}

	return
}

// FgpAuthentication - middleware for checking user's fingerprint
// This handler should be served only after JWT token checking
func FgpAuthentication(env *models.Env, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.LogInfo.Printf("Checking fingerprint")

		token := r.Context().Value("user").(*jwt.Token)

		fingerprintCookie, err := r.Cookie("Secure-Fgp")
		if err != nil {
			env.LogInfo.Printf("Request missing fingerprint")
			RespondWithError(w, http.StatusUnauthorized, InvalidToken, env.LogError)
			return
		}
		rawFingerprint := fingerprintCookie.Value

		claims := token.Claims.(jwt.MapClaims)
		tokenFingerprint := claims["fingerprint"].(string)

		if err = bcrypt.CompareHashAndPassword([]byte(tokenFingerprint), []byte(rawFingerprint)); err != nil {
			env.LogInfo.Printf(
				"Error checking fingeprint: raw fingerprint does not match fingeprint containing in JWT Token")
			RespondWithError(w, http.StatusUnauthorized, InvalidToken, env.LogError)
			return
		}

		role := claims["role"].(string)

		env.LogInfo.Printf("Fingerprint is valid. Serving next http handler")

		ctx := context.WithValue(r.Context(), CtxKey, models.UserRole(role))
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// RegisterUserHandler - checks user registration credentials and inserts login and password into database
func (api *UserAPI) RegisterUserHandler() http.Handler {
	env := api.Env
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.LogInfo.Print("Got new user Registration job")

		credentials, validateError := validateUserRegistrationCredentials(r)
		if validateError != NoError {
			env.LogInfo.Print("User Registration credentials are invalid")
			RespondWithError(w, http.StatusBadRequest, validateError, env.LogError)
			return
		}

		username := credentials.Username
		email := credentials.Email
		password := []byte(credentials.Password)

		var userExists bool
		if err := env.Db.QueryRow(
			"select exists(select from users where email = $1 or username = $2)", email, username).
			Scan(&userExists); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		if userExists {
			env.LogInfo.Printf("User with following credentials: (username: %s; email: %s) already registered",
				username, email)
			RespondWithError(w, http.StatusBadRequest, AlreadyRegistered, env.LogError)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
		if err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("Inserting credentials of user with following credentials: (username: %s; email: %s) "+
			"into database", username, email)

		if _, err = env.Db.Exec("BEGIN TRANSACTION"); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}
		_, err = env.Db.Exec("insert into users (username, email, password) values($1, $2, $3)",
			username, email, string(hashedPassword))
		if err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}
		if _, err := env.Db.Exec("END TRANSACTION"); err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("User with following credentials: (username: %s; email: %s) successfully registered",
			username, email)

		Respond(w, http.StatusOK)
	})
}

func generateRandomContext() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// IsUserAdmin - check if user is admin
func IsUserAdmin(login string, admins []settings.Admin) bool {
	for _, currentAdmin := range admins {
		if currentAdmin.Login == login {
			return true
		}
	}
	return false
}

func createToken(login, fgp string, api *UserAPI) (string, error) {
	var claims models.TokenClaims
	claims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
	claims.Fingerprint = fgp

	if IsUserAdmin(login, api.Admins) {
		claims.Role = roleAdmin
	} else {
		claims.Role = roleUser
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(api.SigningKey)

	return tokenString, err
}

// LoginUserHandler - checks user credentials and returns authorization token
func (api *UserAPI) LoginUserHandler() http.Handler {
	env := api.Env
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.LogInfo.Printf("Got new user Log In job")

		credentials, validateError := validateUserLoginCredentials(r)
		if validateError != NoError {
			env.LogInfo.Print("User Log In credentials are invalid")
			RespondWithError(w, http.StatusBadRequest, validateError, env.LogError)
			return
		}

		username := credentials.Username
		email := credentials.Email
		password := credentials.Password

		var hashedPassword string

		env.LogInfo.Printf("Getting hashed password from database of user with following credentials: "+
			"(username: %s; email: %s)", username, email)

		var err error
		if len(username) == 0 {
			err = env.Db.QueryRow("select username, password from users where email = $1", email).
				Scan(&username, &hashedPassword)
		} else {
			err = env.Db.QueryRow("select password from users where username = $1", username).
				Scan(&hashedPassword)
		}
		if err != nil {
			if err == sql.ErrNoRows {
				env.LogInfo.Printf("Can not get hashed password: user with following credentials: "+
					"(username: %s; email: %s) "+"does not exist", username, email)
				RespondWithError(w, http.StatusUnauthorized, WrongCredentials, env.LogError)
				return
			}
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
			env.LogInfo.Printf("Inputted password by user %s does not match hashed one", username)
			RespondWithError(w, http.StatusUnauthorized, WrongCredentials, env.LogError)
			return
		}

		env.LogInfo.Printf("Generating hashed fingerprint")

		ctx, err := generateRandomContext()
		if err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		// set raw fingerprint cookie
		ctxCookie := &http.Cookie{Name: "Secure-Fgp", Value: ctx, SameSite: http.SameSiteStrictMode, HttpOnly: true,
			Expires: time.Now().Add(time.Hour * 1), Path: "/"}
		http.SetCookie(w, ctxCookie)

		// set username cookie for detecting is user admin
		usernameCookie := &http.Cookie{Name: "Login", Value: username, SameSite: http.SameSiteStrictMode, HttpOnly: true,
			Path: "/"}
		http.SetCookie(w, usernameCookie)

		// generate hashed fingerprint
		hashedFgp, err := bcrypt.GenerateFromPassword([]byte(ctx), bcrypt.DefaultCost)
		if err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("Generating JWT Token with hashed fingerprint")

		token, err := createToken(username, string(hashedFgp), api)
		if err != nil {
			env.LogError.Print(err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError, env.LogError)
			return
		}

		env.LogInfo.Printf("Token successfully generated. Sending token to user %s", username)

		RespondWithBody(w, http.StatusAccepted, token, env.LogError)
	})
}

package restapi

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"github.com/blinky-z/Blog/models"
	"github.com/blinky-z/Blog/service/userService"
	"github.com/dgrijalva/jwt-go"
	pg "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
	"time"
)

// ctxRoleKey - special type for getting user's role from request context
type ctxRoleKey string

const (
	// CtxRoleKey - used to get user's role from request context
	CtxRoleKey = ctxRoleKey("role")
)

// UserAPIHandler - environment container struct to declare all auth handlers as methods
type UserAPIHandler struct {
	db              *sql.DB
	jwtSecret       []byte
	admins          *[]string
	logInfo         *log.Logger
	logError        *log.Logger
	jwtUserProperty string
}

func NewUserAPIHandler(db *sql.DB, jwtSecret []byte, admins *[]string, jwtUserProperty string,
	logInfo, logError *log.Logger) *UserAPIHandler {
	return &UserAPIHandler{
		db:              db,
		jwtSecret:       jwtSecret,
		admins:          admins,
		logInfo:         logInfo,
		logError:        logError,
		jwtUserProperty: jwtUserProperty,
	}
}

// error codes for this API
var (
	// WrongCredentials - user inputs wrong password or login or email while logging in
	WrongCredentials = models.NewRequestErrorCode("WRONG_CREDENTIALS")
	// InvalidEmail - user inputs invalid email while registration or logging in
	InvalidEmail = models.NewRequestErrorCode("INVALID_EMAIL")
	// InvalidUsername - user inputs invalid username while registration
	InvalidUsername = models.NewRequestErrorCode("INVALID_LOGIN")
	// InvalidPassword - user inputs invalid password while registration
	InvalidPassword = models.NewRequestErrorCode("INVALID_PASSWORD")
	// UserAlreadyRegistered - user trying to register account while already registered
	UserAlreadyRegistered = models.NewRequestErrorCode("USER_ALREADY_REGISTERED")
	// IncompleteCredentials - user do not input full credentials: login, email, password
	IncompleteCredentials = models.NewRequestErrorCode("INCOMPLETE_CREDENTIALS")
	// InvalidFingerprint - user provided non-authentic or malformed fingerprint
	InvalidFingerprint = models.NewRequestErrorCode("INVALID_FINGERPRINT")
	// InvalidToken - user provided non-authentic or malformed token
	InvalidToken = models.NewRequestErrorCode("INVALID_TOKEN")
)

// constants for use in validator methods
const (
	// MinPwdLen - minimum length of user password
	MinPwdLen int = 8
	// MaxPwdLen - maximum length of user password
	MaxPwdLen int = 38

	// MinUsernameLen - minimum username length
	MinUsernameLen int = 6
	// MaxUsernameLen - maximum username length
	MaxUsernameLen int = 36

	// MaxEmailLen - maximum length of email
	MaxEmailLen int = 255
)

// user roles
const (
	roleAdmin = models.UserRole("admin")
	roleUser  = models.UserRole("user")
)

func validateEmail(email string) models.RequestErrorCode {
	email = strings.TrimSpace(email)
	if strings.Count(email, "@") != 1 || len(email) > MaxEmailLen || email[0] == '@' || email[len(email)-1] == '@' {
		return InvalidEmail
	}
	return nil
}

func validateUsername(username string) models.RequestErrorCode {
	authorLen := len(strings.TrimSpace(username))
	if authorLen > MaxUsernameLen || authorLen < MinUsernameLen {
		return InvalidUsername
	}

	return nil
}

func validatePassword(password string) models.RequestErrorCode {
	passwordLen := len(password)
	if passwordLen < MinPwdLen || passwordLen > MaxPwdLen {
		return InvalidPassword
	}
	return nil
}

func validateRegistrationRequest(request models.RegistrationRequest) models.RequestErrorCode {
	username := request.Username
	email := request.Email
	password := request.Password

	if username == "" || email == "" || password == "" {
		return IncompleteCredentials
	}
	if err := validateEmail(email); err != nil {
		return err
	}
	if err := validateUsername(username); err != nil {
		return err
	}
	return validatePassword(password)
}

func validateLoginRequest(request models.LoginRequest) models.RequestErrorCode {
	username := request.Username
	email := request.Email
	password := request.Password

	// user can provide only either username or email, but not both
	if (username == "" && email == "") || password == "" {
		return IncompleteCredentials
	}

	if email != "" {
		if err := validateEmail(email); err != nil {
			return err
		}
	} else {
		if err := validateUsername(username); err != nil {
			return err
		}
	}
	return validatePassword(password)
}

// FgpAuthentication - middleware for checking fingerprint
// This handler should be a next step after JWT token checking
func (api *UserAPIHandler) FgpAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value(api.jwtUserProperty).(*jwt.Token)
		tokenClaims := token.Claims.(jwt.MapClaims)

		fgpCookie, err := r.Cookie("Secure-Fgp")
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, InvalidFingerprint)
			return
		}

		rawFgp := []byte(fgpCookie.Value)
		hashedFgp := []byte(tokenClaims["fingerprint"].(string))

		if err = bcrypt.CompareHashAndPassword(hashedFgp, rawFgp); err != nil {
			RespondWithError(w, http.StatusUnauthorized, InvalidFingerprint)
			return
		}

		// create a new context with user role in it. We will pass this context to the next handler
		userRole := tokenClaims["role"].(string)
		ctx := context.WithValue(r.Context(), CtxRoleKey, models.UserRole(userRole))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// unused
// RegisterUserHandler - serves registration requests
// This function generates hashed password with bcrypt library before saving user in database
// use bcrypt.CompareHashAndPassword function to compare password from login form with actual hashed password
func (api *UserAPIHandler) RegisterUserHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request models.RegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			RespondWithError(w, http.StatusBadRequest, BadRequestBody)
			return
		}

		logInfo.Printf("Got new user registration request. Request: %+v", request)

		validateError := validateRegistrationRequest(request)
		if validateError != nil {
			logError.Printf("Can't register user: invalid request. Error: %s", validateError)
			RespondWithError(w, http.StatusBadRequest, validateError)
			return
		}

		email := strings.TrimSpace(request.Email)
		username := strings.TrimSpace(request.Username)
		password := []byte(request.Password)

		// generate hashed password
		hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
		if err != nil {
			logError.Printf("Can't register user: error generating hashed password. Username: %s. Error: %s",
				username, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		if err := userService.Save(api.db, username, email, string(hashedPassword)); err != nil {
			logError.Printf("Error saving user in database. Username: %s. Error: %s", username, err)

			// check for duplicate error
			if err.(*pg.Error).Code == "23505" {
				RespondWithError(w, http.StatusBadRequest, UserAlreadyRegistered)
				return
			}
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		logInfo.Printf("User registered. User info: Username: %s, Email: %s", username, email)
		Respond(w, http.StatusOK)
	})
}

// generateFingerprint - generates fingerprint
// This function uses cryptographically secure random number generator
func generateFingerprint() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// IsUserAdmin - check if the given user is an admin
func IsUserAdmin(username string, admins *[]string) bool {
	for _, currentAdminUsername := range *admins {
		if username == currentAdminUsername {
			return true
		}
	}
	return false
}

// generateJwtToken - generates JWT token
// Generate and hash fingerprint before calling this function
func generateJwtToken(login, fgp string, api *UserAPIHandler) (string, error) {
	var claims models.TokenClaims

	// set required claims
	claims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
	claims.Fingerprint = fgp
	if IsUserAdmin(login, api.admins) {
		claims.Role = roleAdmin
	} else {
		claims.Role = roleUser
	}

	// generate and sign the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(api.jwtSecret)
}

// unused
// LoginUserHandler - serves user login request
// This function sends back generated JWT token as payload
func (api *UserAPIHandler) LoginUserHandler() http.Handler {
	logInfo := api.logInfo
	logError := api.logError
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			RespondWithError(w, http.StatusBadRequest, BadRequestBody)
			return
		}

		logInfo.Printf("Got new user login request. Request: %+v", request)

		username := strings.TrimSpace(request.Username)
		email := strings.TrimSpace(request.Email)
		password := request.Password

		validateError := validateLoginRequest(request)
		if validateError != nil {
			logInfo.Printf("Bad login: invalid credentials. Username: %s, email: %s", username, email)
			RespondWithError(w, http.StatusBadRequest, validateError)
			return
		}

		logInfo.Printf("Getting hashed password from database for user login. Username: %s, email: %s", username, email)

		var hashedPassword string
		var err error

		// get hashed password from database
		if username != "" {
			err = api.db.QueryRow("select password from users where username = $1", username).
				Scan(&hashedPassword)
		} else {
			// also get username for this user as we will need this later to set username in cookie
			err = api.db.QueryRow("select username, password from users where email = $1", email).
				Scan(&username, &hashedPassword)
		}
		if err != nil {
			if err == sql.ErrNoRows {
				logInfo.Printf("Login failed: user does not exist. Username: %s, email: %s", username, email)
				RespondWithError(w, http.StatusUnauthorized, WrongCredentials)
				return
			}
			logError.Printf("Bad login: error retrieving user info from database. Username: %s, email: %s. Error: %s",
				username, email, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		// compare given password and the hashed one from database
		if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
			logInfo.Printf("Login failed: inputted password does not match hashed one. Username: %s, email: %s",
				username, email)
			RespondWithError(w, http.StatusUnauthorized, WrongCredentials)
			return
		}

		logInfo.Printf("Generating fingerprint for user login. Username: %s, email: %s", username, email)
		rawFgp, err := generateFingerprint()
		if err != nil {
			logError.Printf("Bad login: error generating fingerprint. Username: %s, email: %s. Error: %s",
				username, email, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		// set raw fingerprint cookie. This cookie is inaccessible from JS
		ctxCookie := &http.Cookie{Name: "Secure-Fgp", Value: rawFgp, SameSite: http.SameSiteStrictMode, HttpOnly: true,
			Secure: true, Expires: time.Now().Add(time.Hour * 1), Path: "/"}
		http.SetCookie(w, ctxCookie)

		// hash generated fingerprint for storing in JWT token
		hashedFgp, err := bcrypt.GenerateFromPassword([]byte(rawFgp), bcrypt.DefaultCost)
		if err != nil {
			logError.Printf("Bad login: error hashing generated fingerprint. Username: %s, email: %s. Error: %s",
				username, email, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		// create JWT token with username and hashed fingerprint in it
		// we will save the token on the client side
		token, err := generateJwtToken(username, string(hashedFgp), api)
		if err != nil {
			logError.Printf("Bad login: error generating JWT token. Username: %s, email: %s. Error: %s",
				username, email, err)
			RespondWithError(w, http.StatusInternalServerError, TechnicalError)
			return
		}

		logInfo.Printf("Successful login. Username: %s, email: %s", username, email)
		RespondWithBody(w, http.StatusOK, token)
	})
}

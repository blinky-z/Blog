package tests

import (
	"github.com/blinky-z/Blog/server"
	"github.com/google/uuid"
	"net/http"
	"os"
	"testing"
	"time"
)

// performs required initialization and runs tests
func TestMain(m *testing.M) {
	loginUsername = uuid.New().String()
	loginEmail = loginUsername + "@gmail.com"
	loginPassword = uuid.New().String() + "Z"
	admins := loginUsername
	serverPort := "8080"
	address = "localhost:" + serverPort

	_ = os.Setenv("JWT_SECRET_KEY", "testSecretKey")
	_ = os.Setenv("ADMINS", admins)
	_ = os.Setenv("SERVER_PORT", serverPort)

	go server.RunServer()
	for {
		resp, err := http.Get("http://" + address + "/api/hc")
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	db = server.Db

	// register a new user for tests
	{
		r := registerUser(loginUsername, loginEmail, loginPassword)
		assertNiceResponse(r, http.StatusOK)
	}

	// login with registered user and save auth data
	{
		r := loginUser("", loginEmail, loginPassword)
		assertNiceResponse(r, http.StatusOK)
		setNewAuthData(r)
	}

	code := m.Run()
	os.Exit(code)
}

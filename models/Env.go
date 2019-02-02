package models

import (
	"database/sql"
	"log"
)

// Env - environment structure for providing other packages db connection, admins list and other stuff
type Env struct {
	// LogInfo - log for writing usual messages
	LogInfo *log.Logger
	// LogError - log for writing server error messages
	LogError *log.Logger
	// Db - database connection pointer
	Db *sql.DB
	// SigningKey - secret key for creating token
	SigningKey []byte
	// Admins - List of admins that own permissions to create, update, delete posts
	Admins []User
}

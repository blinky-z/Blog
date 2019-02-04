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
}

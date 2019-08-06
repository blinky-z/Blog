package userService

import (
	"database/sql"
)

const (
	// usersInsertFields - fields that should be filled while inserting a new entity
	usersInsertFields = "username, email, password"
)

// Save - saves a new user in database
func Save(db *sql.DB, username, email, password string) error {
	_, err := db.Exec("insert into users ("+usersInsertFields+") values ($1, $2, $3)", username, email, password)
	return err
}

// ExistsByUsernameOrEmail - check if user with the given username OR email exists
// returns boolean indicating whether user exists or not and error
func ExistsByUsernameOrEmail(db *sql.DB, username, email string) (bool, error) {
	var isUserExists bool
	err := db.QueryRow(
		"select exists(select from users where email = $1 or username = $2)", email, username).
		Scan(&isUserExists)
	return isUserExists, err
}

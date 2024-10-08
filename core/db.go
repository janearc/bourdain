package core

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // PostgreSQL driver
	"net/url"
	"strings"
)

// ConnectDB establishes a connection to the PostgreSQL database
func ConnectDB(config *Config) (*sql.DB, error) {
	dsn := getDSN(config) // Use the centralized DSN function

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	// Check if the connection is working
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return db, nil
}

// getDSN constructs the database connection string with escaped credentials
func getDSN(config *Config) string {
	escapedUser := url.QueryEscape(config.Database.User)
	escapedPassword := url.QueryEscape(config.Database.Password)
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		escapedUser,
		escapedPassword,
		config.Database.Host,
		config.Database.Port,
		config.Database.DbName)
}

// formatArrayForPostgres formats an array for PostgreSQL, escaping elements and wrapping them in curly braces
func FormatArrayForPostgres(arr []string) string {
	for i, elem := range arr {
		arr[i] = fmt.Sprintf("\"%s\"", elem) // Escape elements with quotes
	}
	return fmt.Sprintf("{%s}", strings.Join(arr, ",")) // Wrap in curly braces
}

package core

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
	"net/url"
	"strings"
)

func ConnectDB(config *Config) (*sql.DB, error) {
	dsn := getDSN(config)

	// Log the DSN for verification
	logrus.Infof("Connecting to database with DSN: %s", dsn)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return db, nil
}

func getDSN(config *Config) string {
	logrus.Info("getDSN is being called") // This will crash the app to ensure it runs this code
	escapedUser := url.QueryEscape(config.Database.User)
	escapedPassword := url.QueryEscape(config.Database.Password)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		escapedUser,
		escapedPassword,
		config.Database.Host,
		config.Database.Port,
		config.Database.DbName)
	logrus.Infof("Constructed DSN: %s", dsn)
	return dsn
}

// FormatArrayForPostgres formats an array for PostgreSQL, escaping elements and wrapping them in curly braces
func FormatArrayForPostgres(arr []string) string {
	for i, elem := range arr {
		arr[i] = fmt.Sprintf("\"%s\"", elem) // Escape elements with quotes
	}
	return fmt.Sprintf("{%s}", strings.Join(arr, ",")) // Wrap in curly braces
}

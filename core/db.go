package core

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
	"net/url"
	"os"
	"path/filepath"
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

// LoadSQLFile loads a SQL file from the tooling/queries directory
func LoadSQLFile(queryName string) (string, error) {
	filePath := filepath.Join("tooling", "queries", queryName)
	sqlBytes, err := os.ReadFile(filepath.Clean(filePath)) // Use os.ReadFile instead of ioutil.ReadFile
	if err != nil {
		return "", err
	}
	return string(sqlBytes), nil
}

// ExecSQLFromFile executes a SQL file from the queries directory (when expecting no rows)
func ExecSQLFromFile(db *sql.DB, queryName string, args ...interface{}) (sql.Result, error) {
	sqlQuery, err := LoadSQLFile(queryName)
	if err != nil {
		return nil, err
	}
	return db.Exec(sqlQuery, args...)
}

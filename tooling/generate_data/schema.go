package main

import (
	"database/sql"
	"github.com/janearc/bourdain/core"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// buildSchema builds the schema from static SQL files to keep that out of the Golang code
func buildSchema(db *sql.DB) {
	remoteDB, err := core.GetCurrentDatabase(db)
	if err != nil {
		logrus.Fatalf("Error getting current database: %v", err)
	} else {
		logrus.Infof("[buildschema] Current database: %s", remoteDB)
	}

	// Define the path to the SQL directory (inside docker)
	sqlDir := "/config/queries"

	// Read all the files from the SQL directory
	files, err := os.ReadDir(sqlDir)
	if err != nil {
		logrus.Fatalf("Error reading directory: %v", err)
	}

	// Filter for SQL files
	var sqlFiles []os.DirEntry
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			sqlFiles = append(sqlFiles, file)
		}
	}

	// Sort the files numerically based on their prefix
	sort.Slice(sqlFiles, func(i, j int) bool {
		num1 := extractFilePrefix(sqlFiles[i].Name())
		num2 := extractFilePrefix(sqlFiles[j].Name())
		return num1 < num2
	})

	// Iterate over each sorted file and execute them
	for _, file := range sqlFiles {
		sqlFilePath := filepath.Join(sqlDir, file.Name())

		// Execute each SQL file normally for schema creation
		_, err := core.ExecSQLFromFile(db, sqlFilePath)
		if err != nil {
			logrus.Fatalf("Error executing SQL file %s: %v", file.Name(), err)
		}
		logrus.Infof("Successfully executed SQL file: %s", file.Name())
	}

	logrus.Info("All SQL entities created.")
}

// runPopulateTops runs the populate_tops function after schema creation
func runPopulateTops(db *sql.DB) error {
	// Log the action to ensure visibility
	logrus.Info("Attempting to populate tops...")

	// Check if the connection is still valid
	err := db.Ping()
	if err != nil {
		logrus.Fatalf("Database connection error during tops population: %v", err)
	}

	// Log before executing the populate_tops function
	logrus.Info("Executing populate_tops() stored procedure...")

	// Call populate_tops function using SELECT, as it's a stored procedure
	_, err = db.Exec(`SELECT populate_tops();`)
	if err != nil {
		logrus.Errorf("Error executing populate_tops(): %v", err)
		return err
	}

	// Confirm successful execution
	logrus.Info("Tops populated successfully after schema creation.")
	return nil
}

// extractFilePrefix extracts the numeric prefix from a file name (e.g., "00_restaurants.sql" -> 0)
func extractFilePrefix(fileName string) int {
	parts := strings.Split(fileName, "_")
	if len(parts) > 0 {
		if num, err := strconv.Atoi(parts[0]); err == nil {
			return num
		}
	}
	return 0 // Default to 0 if no prefix found
}

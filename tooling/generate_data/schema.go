package main

import (
	"database/sql"
	"github.com/janearc/bourdain/core"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// buildSchema builds the schema from static sql files to keep that out of the golang puddin
func buildSchema(db *sql.DB) {
	remoteDB, err := getCurrentDatabase(db)
	if err != nil {
		log.Fatalf("Error getting current database: %v", err)
	} else {
		log.Printf("[buildschema] Current database: %s", remoteDB)
	}

	// Define the path to the SQL directory (inside docker)
	sqlDir := "/config/queries"

	// Read all the files from the SQL directory
	files, err := os.ReadDir(sqlDir)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
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
		// Extract the numeric prefix from the file name
		num1 := extractFilePrefix(sqlFiles[i].Name())
		num2 := extractFilePrefix(sqlFiles[j].Name())
		return num1 < num2
	})

	// Iterate over each sorted file and execute them
	for _, file := range sqlFiles {
		sqlFilePath := filepath.Join(sqlDir, file.Name())

		// Execute each SQL file
		_, err := core.ExecSQLFromFile(db, sqlFilePath)
		if err != nil {
			log.Fatalf("Error executing SQL file %s: %v", file.Name(), err)
		}
		log.Printf("Successfully executed SQL file: %s", file.Name())
	}

	log.Println("All SQL entities created successfully")
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

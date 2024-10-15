package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/janearc/bourdain/core"
	"math/rand"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Initialize a local random generator
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// createDatabase enables necessary extensions and configures the database.
func createDatabase(db *sql.DB) {
	dbName, err := core.GetCurrentDatabase(db)
	if err != nil {
		logrus.Fatalf("Error getting current database: %v", err)
	}

	// Enable extensions: uuid-ossp, PostGIS
	extensions := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		`CREATE EXTENSION IF NOT EXISTS postgis;`,
	}
	for _, ext := range extensions {
		if _, err := db.Exec(ext); err != nil {
			logrus.Fatalf("Error creating extension: %v", err)
		}
	}

	// Set logging for better insights
	dbSettings := []string{
		fmt.Sprintf(`ALTER DATABASE %s SET log_statement = 'all';`, dbName),
		`ALTER SYSTEM SET client_min_messages TO 'NOTICE';`,
	}
	for _, setting := range dbSettings {
		if _, err := db.Exec(setting); err != nil {
			logrus.Fatalf("Error setting database parameters: %v", err)
		}
	}

	logrus.Info("Database extensions and settings applied")
}

// insertRestaurants inserts random restaurant data into the database.
func insertRestaurants(count int, stdout bool, db *sql.DB) {
	for i := 0; i < count; i++ {
		name := RandomRestaurantName(rng)
		lat, lon := randomLocation()
		capacity := map[string]int{
			"two-top":  rng.Intn(10) + 1,
			"four-top": rng.Intn(10) + 1,
			"six-top":  rng.Intn(5) + 1,
		}
		endors := randomEndorsements()

		capacityJSON, _ := json.Marshal(capacity)
		endorsJSON, _ := json.Marshal(endors)

		openingTime, closingTime := randomBusinessHours()

		sqlStmt := `
			INSERT INTO restaurants (name, capacity, endorsements, location, opening_time, closing_time)
			VALUES ($1, $2::jsonb, $3::jsonb, ST_SetSRID(ST_MakePoint($4, $5), 4326), $6, $7)
			RETURNING id;`

		if stdout {
			logrus.Infof("Would execute: %s", sqlStmt)
		} else {
			var id string
			if err := db.QueryRow(sqlStmt, name, string(capacityJSON), string(endorsJSON), lon, lat, openingTime, closingTime).Scan(&id); err != nil {
				logrus.Errorf("Error inserting restaurant: %v", err)
			}
		}
	}
}

// insertDiners inserts a specified number of diners into the database
func insertDiners(count int, stdout bool, db *sql.DB) {
	for i := 0; i < count; i++ {
		name := RandomName(rng)       // Generate a random diner name
		lat, lon := randomLocation()  // Generate random latitude and longitude
		prefs := randomEndorsements() // Generate random preferences

		// Marshal preferences to JSON (since it's stored as JSONB in the database)
		prefsJSON, err := json.Marshal(prefs)
		if err != nil {
			logrus.Errorf("Error marshaling preferences JSON: %v", err)
			continue
		}

		// SQL statement to insert a diner
		sqlStmt := `
			INSERT INTO diners (name, preferences, location)
			VALUES ($1, $2::jsonb, ST_SetSRID(ST_MakePoint($3, $4), 4326))
			RETURNING id;
		`

		// Log SQL for stdout mode, or execute the insert if not in stdout mode
		if stdout || db == nil {
			logrus.Infof("Would execute: %s", sqlStmt)
		} else {
			var id string
			err := db.QueryRow(sqlStmt, name, string(prefsJSON), lon, lat).Scan(&id)
			if err != nil {
				logrus.Errorf("Error inserting diner: %v", err)
			}
		}
	}
}

// randomBusinessHours returns randomly selected opening and closing times for restaurants.
func randomBusinessHours() (string, string) {
	switch r := rng.Float64(); {
	case r < 0.1: // 24-hour restaurant (10%)
		return "00:00", "23:59"
	case r < 0.35: // 10am to 10pm (25%)
		return "10:00", "22:00"
	default: // Dinner place (5:30pm to 11:30pm)
		return "17:30", "23:30"
	}
}

// main is the entry point of the application. It handles different modes like DB initialization, SQL stdout, and name generation.
func main() {
	stdout := flag.Bool("stdout", false, "Print SQL statements to stdout instead of executing")
	initdb := flag.Bool("initdb", false, "Initialize the database with test data")
	configFile := flag.String("config", "/config/config.json", "Path to the config file")
	properName := flag.Bool("proper-name", false, "Generate a random proper name")
	restaurantName := flag.Bool("restaurant-name", false, "Generate a random restaurant name")
	flag.Parse()

	// Configure Logrus
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetLevel(logrus.InfoLevel)

	if *properName {
		fmt.Println(RandomName(rng))
		return
	}
	if *restaurantName {
		fmt.Println(RandomRestaurantName(rng))
		return
	}

	// Load configuration
	config, err := core.LoadConfig(*configFile)
	if err != nil {
		logrus.Fatalf("Error loading config: %v", err)
	}

	// Connect to the database
	db, err := core.ConnectDB(config)
	if err != nil {
		logrus.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	if *stdout {
		logrus.Info("Generating SQL statements...")
		insertRestaurants(10, true, nil)
	} else if *initdb {
		logrus.Info("Initializing database...")
		createDatabase(db)
		buildSchema(db)

		// if stuff gets slow, turn these down a little
		insertRestaurants(15000, false, db)
		insertDiners(250, false, db)

		if err = runPopulateTops(db); err != nil {
			logrus.Fatalf("Error populating tops: %v", err)
		}
		logrus.Info("Database initialized successfully with sample data.")
	} else {
		logrus.Warn("Please specify --stdout, --initdb, --proper-name, or --restaurant-name.")
	}
}

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

func createDatabase(db *sql.DB) {
	remoteDB, rderr := getCurrentDatabase(db)

	if rderr != nil {
		logrus.Fatalf("Error getting current database: %v", rderr)
	} else {
		logrus.Infof("[createdb] Current database: %s", remoteDB)
	}

	// Enable the uuid-ossp extension for generating UUIDs
	_, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		logrus.Fatalf("Error creating uuid-ossp extension: %v", err)
	}

	// Enable the PostGIS extension for geography support
	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS postgis;`)
	if err != nil {
		logrus.Fatalf("Error creating PostGIS extension: %v", err)
	}

	logrus.Info("Database extensions and setup complete")
}

func createTables(db *sql.DB) {
	remoteDB, rderr := getCurrentDatabase(db)

	if rderr != nil {
		logrus.Fatalf("Error getting current database: %v", rderr)
	} else {
		logrus.Infof("[createtables] Current database: %s", remoteDB)
	}

	// Create diners table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS diners (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			preferences JSONB NOT NULL,
			location GEOGRAPHY(POINT, 4326)
		);
	`)
	if err != nil {
		logrus.Fatalf("Error creating diners table: %v", err)
	}

	// Create restaurants table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS restaurants (
		    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		    name VARCHAR(255) NOT NULL,
		    capacity JSONB NOT NULL,
		    endorsements JSONB NOT NULL,
		    location GEOGRAPHY(POINT, 4326),
		    opening_time TIME NOT NULL,
		    closing_time TIME NOT NULL
		);
	`)
	if err != nil {
		logrus.Fatalf("Error creating restaurants table: %v", err)
	}

	// Create reservations table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS reservations (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			restaurant_id UUID REFERENCES restaurants(id),
			diner_id UUID REFERENCES diners(id),
			reservation_time TIMESTAMP NOT NULL,
			num_diners INTEGER NOT NULL
		);
	`)
	if err != nil {
		logrus.Fatalf("Error creating reservations table: %v", err)
	}

	logrus.Info("Tables created successfully")
}

func randomLocation() (float64, float64) {
	latBounds := [2]float64{40.5774, 40.9176} // Rough latitude range for Manhattan, Brooklyn, Bronx
	lonBounds := [2]float64{-74.15, -73.7004} // Rough longitude range

	lat := latBounds[0] + (rng.Float64() * (latBounds[1] - latBounds[0]))
	lon := lonBounds[0] + (rng.Float64() * (lonBounds[1] - lonBounds[0]))
	return lat, lon
}

func randomEndorsements() []string {
	num := rng.Intn(3) + 1
	var selected []string
	for i := 0; i < num; i++ {
		selected = append(selected, Endorsements[rng.Intn(len(Endorsements))])
	}
	return selected
}

func randomPreferences() []string {
	return randomEndorsements()
}

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

		// Marshal capacity and endorsements to JSON
		capacityJSON, err := json.Marshal(capacity)
		if err != nil {
			logrus.Errorf("Error marshaling capacity JSON: %v", err)
			continue
		}
		endorsJSON, err := json.Marshal(endors)
		if err != nil {
			logrus.Errorf("Error marshaling endorsements JSON: %v", err)
			continue
		}

		// Randomly assign hours based on the given probabilities
		var openingTime, closingTime string

		// 10% chance of being a 24-hour restaurant
		if rng.Float64() < 0.1 {
			openingTime = "00:00"
			closingTime = "23:59"
		} else if rng.Float64() < 0.25 { // 25% chance of being open from 10am to 10pm
			openingTime = "10:00"
			closingTime = "22:00"
		} else { // The rest will be dinner places (5:30pm to 11:30pm)
			openingTime = "17:30"
			closingTime = "23:30"
		}

		// Insert the restaurant and return the generated UUID
		sqlStmt := `
			INSERT INTO restaurants (name, capacity, endorsements, location, opening_time, closing_time)
			VALUES ($1, $2::jsonb, $3::jsonb, ST_SetSRID(ST_MakePoint($4, $5), 4326), $6, $7)
			RETURNING id;`

		if stdout || db == nil {
			logrus.Infof("Would execute: %s", sqlStmt)
		} else {
			var id string
			err := db.QueryRow(sqlStmt, name, string(capacityJSON), string(endorsJSON), lon, lat, openingTime, closingTime).Scan(&id)
			if err != nil {
				logrus.Errorf("Error inserting restaurant: %v", err)
			} else {
				logrus.Infof("Inserted restaurant with UUID: %s", id)
			}
		}
	}
}

func insertDiners(count int, stdout bool, db *sql.DB) {
	for i := 0; i < count; i++ {
		name := RandomName(rng)
		lat, lon := randomLocation()
		prefs := randomPreferences()

		// Marshal preferences to JSON (since it's a JSONB field)
		prefsJSON, err := json.Marshal(prefs)
		if err != nil {
			logrus.Errorf("Error marshaling preferences JSON: %v", err)
			continue
		}

		// Insert the diner and return the generated UUID
		sqlStmt := `
			INSERT INTO diners (name, preferences, location)
			VALUES ($1, $2::jsonb, ST_SetSRID(ST_MakePoint($3, $4), 4326))
			RETURNING id;`

		if stdout || db == nil {
			logrus.Infof("Would execute: %s", sqlStmt)
		} else {
			var id string
			err := db.QueryRow(sqlStmt, name, string(prefsJSON), lon, lat).Scan(&id)
			if err != nil {
				logrus.Errorf("Error inserting diner: %v", err)
			} else {
				logrus.Infof("Inserted diner with UUID: %s", id)
			}
		}
	}
}

func getCurrentDatabase(db *sql.DB) (string, error) {
	var dbName string
	err := db.QueryRow("SELECT current_database();").Scan(&dbName)
	if err != nil {
		return "", fmt.Errorf("error fetching current database: %v", err)
	}
	return dbName, nil
}

func main() {
	// Flags to determine mode of operation
	stdout := flag.Bool("stdout", false, "Print SQL statements to stdout instead of executing")
	initdb := flag.Bool("initdb", false, "Initialize the database with test data")
	configFile := flag.String("config", "/config/config.json", "Path to the config file")
	properName := flag.Bool("proper-name", false, "Generate a random proper name")
	restaurantName := flag.Bool("restaurant-name", false, "Generate a random restaurant name")
	flag.Parse()

	// Configure Logrus
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	// Handle the fun name generation flags
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

	// Connect to the database using core.ConnectDB
	db, err := core.ConnectDB(config)
	if err != nil {
		logrus.Fatalf("Error connecting to the database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logrus.Errorf("Error closing the database connection: %v", err)
		}
	}()

	if *stdout {
		// Generate and print SQL statements instead of inserting into DB
		logrus.Info("Generating SQL statements for restaurants and diners...")
		insertRestaurants(100, true, nil)
		insertDiners(500, true, nil)
	} else if *initdb {
		logrus.Info("Initializing database...")

		// Create the database extensions (UUID, PostGIS)
		createDatabase(db)

		// Create tables if they don't exist
		createTables(db)

		// Create the helper functions
		createPlSqlFunctions(db)

		// Insert data into the tables
		insertRestaurants(100, false, db)
		insertDiners(500, false, db)

		logrus.Info("Database initialized successfully with sample data.")
	} else {
		logrus.Warn("Please specify either --stdout, --initdb, --proper-name, or --restaurant-name.")
	}
}

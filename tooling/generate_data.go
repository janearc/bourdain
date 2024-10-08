package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/janearc/bourdain/core"
	"math/rand"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Initialize a local random generator
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func createTables(db *sql.DB) {
	// Enable PostGIS extension
	_, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS postgis;`)
	if err != nil {
		logrus.Fatalf("Error enabling PostGIS extension: %v", err)
	}

	// Create the restaurants table
	restaurantTable := `
	CREATE TABLE IF NOT EXISTS restaurants (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		capacity JSONB NOT NULL,
		endorsements TEXT[] NOT NULL,
		location GEOGRAPHY(POINT, 4326)
	);`
	_, err = db.Exec(restaurantTable)
	if err != nil {
		logrus.Fatalf("Error creating restaurants table: %v", err)
	}

	// Create the diners table
	dinersTable := `
	CREATE TABLE IF NOT EXISTS diners (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		preferences TEXT[] NOT NULL,
		location GEOGRAPHY(POINT, 4326)
	);`
	_, err = db.Exec(dinersTable)
	if err != nil {
		logrus.Fatalf("Error creating diners table: %v", err)
	}
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
		name := RandomRestaurantName(rng) // No need to escape, handled by parameterized queries
		lat, lon := randomLocation()
		capacity := fmt.Sprintf(`{"two-top": %d, "four-top": %d, "six-top": %d}`, rng.Intn(10)+1, rng.Intn(10)+1, rng.Intn(5)+1)
		endors := randomEndorsements()
		endorsArray := formatArrayForPostgres(endors)

		sqlStmt := "INSERT INTO restaurants (name, capacity, endorsements, location) VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326));"

		if stdout || db == nil {
			// For stdout mode, still need to display the query with formatted values
			logrus.Infof("Would execute: %s", fmt.Sprintf(sqlStmt, name, capacity, endorsArray, lon, lat))
		} else {
			_, err := db.Exec(sqlStmt, name, capacity, endorsArray, lon, lat)
			if err != nil {
				logrus.Errorf("Error executing insert: %v", err)
			}
		}
	}
}

func insertDiners(count int, stdout bool, db *sql.DB) {
	for i := 0; i < count; i++ {
		name := RandomName(rng) // No need to escape, handled by parameterized queries
		lat, lon := randomLocation()
		prefs := randomPreferences()
		prefsArray := formatArrayForPostgres(prefs)

		sqlStmt := "INSERT INTO diners (name, preferences, location) VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326));"

		if stdout || db == nil {
			// For stdout mode, still need to display the query with formatted values
			logrus.Infof("Would execute: %s", fmt.Sprintf(sqlStmt, name, prefsArray, lon, lat))
		} else {
			_, err := db.Exec(sqlStmt, name, prefsArray, lon, lat)
			if err != nil {
				logrus.Errorf("Error executing insert: %v", err)
			}
		}
	}
}

func formatArrayForPostgres(arr []string) string {
	// Escape each element and join them with commas, then wrap in curly braces
	for i, elem := range arr {
		arr[i] = fmt.Sprintf("\"%s\"", elem)
	}
	return fmt.Sprintf("{%s}", strings.Join(arr, ","))
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

	// Connect to the database using core.ConnectDB, which now includes the DSN logic
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
		logrus.Info("Generating SQL statements...")
		insertRestaurants(100, true, nil)
		insertDiners(500, true, nil)
	} else if *initdb {
		logrus.Info("Initializing database...")

		// Create tables if they don't exist
		createTables(db)

		// Insert data into the tables
		insertRestaurants(100, false, db)
		insertDiners(500, false, db)
	} else {
		logrus.Warn("Please specify either --stdout, --initdb, --proper-name, or --restaurant-name.")
	}
}

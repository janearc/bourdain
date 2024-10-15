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
	remoteDB, err := getCurrentDatabase(db)

	if err != nil {
		logrus.Fatalf("Error getting current database: %v", err)
	} else {
		logrus.Infof("[createdb] Current database: %s", remoteDB)
	}

	// Enable the uuid-ossp extension for generating UUIDs
	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		logrus.Fatalf("Error creating uuid-ossp extension: %v", err)
	}

	// Enable the PostGIS extension for geography support
	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS postgis;`)
	if err != nil {
		logrus.Fatalf("Error creating PostGIS extension: %v", err)
	}

	// set logging to horrifying
	_, err = db.Exec(`ALTER DATABASE ` + remoteDB + ` SET log_statement = 'all';`)
	if err != nil {
		logrus.Fatalf("Error setting log_statement: %v", err)
	}

	_, err = db.Exec("ALTER SYSTEM SET client_min_messages TO 'NOTICE';")
	if err != nil {
		logrus.Fatalf("Error setting client_min_messages: %v", err)
	}

	logrus.Info("Database extensions and setup complete")
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
	// Create a slice of preferences based on the weights
	allPreferences := make([]string, 0, len(endorsementWeights))
	for pref, weight := range endorsementWeights {
		// Add preference to the list multiple times based on its weight
		count := int(weight * 100) // Adjust weight scaling as needed
		for i := 0; i < count; i++ {
			allPreferences = append(allPreferences, pref)
		}
	}

	// 25% chance to have no preferences at all
	if rand.Float64() < 0.25 {
		return []string{}
	}

	// 60% chance to have exactly one preference
	if rand.Float64() < 0.60 {
		return []string{allPreferences[rand.Intn(len(allPreferences))]}
	}

	// Only 15% chance to have two preferences
	if rand.Float64() < 0.15 {
		pref1 := allPreferences[rand.Intn(len(allPreferences))]
		pref2 := allPreferences[rand.Intn(len(allPreferences))]
		// Ensure the two preferences are distinct
		for pref1 == pref2 {
			pref2 = allPreferences[rand.Intn(len(allPreferences))]
		}
		return []string{pref1, pref2}
	}

	// Even rarer chance (5%) to have three or more preferences
	numPrefs := rand.Intn(3) + 1 // Randomly pick 1 to 3 preferences
	rand.Shuffle(len(allPreferences), func(i, j int) { allPreferences[i], allPreferences[j] = allPreferences[j], allPreferences[i] })
	return allPreferences[:numPrefs]
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
		insertRestaurants(10, true, nil)
		insertDiners(10, true, nil)
	} else if *initdb {
		logrus.Info("Initializing database...")

		// Create the database extensions (UUID, PostGIS)
		createDatabase(db)

		// Build the schema
		buildSchema(db)

		// Insert data into the tables
		insertRestaurants(5000, false, db)
		insertDiners(1000, false, db)

		// Now that restaurants are inserted, populate the tops
		err = runPopulateTops(db)
		if err != nil {
			logrus.Fatalf("Error populating tops: %v", err)
		}

		logrus.Info("Database initialized successfully with sample data.")
	} else {
		logrus.Warn("Please specify either --stdout, --initdb, --proper-name, or --restaurant-name.")
	}
}

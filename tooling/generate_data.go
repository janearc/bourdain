package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Initialize a local random generator
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

type Config struct {
	Database struct {
		User     string `json:"user"`
		Password string `json:"password"`
		DbName   string `json:"dbname"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
	} `json:"database"`
}

func loadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func getDSN(config *Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.DbName)
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

func insertRestaurants(count int, stdout bool) {
	for i := 0; i < count; i++ {
		name := RandomRestaurantName(rng)
		lat, lon := randomLocation()
		capacity := fmt.Sprintf(`{"two-top": %d, "four-top": %d, "six-top": %d}`, rng.Intn(10)+1, rng.Intn(10)+1, rng.Intn(5)+1)
		endors := randomEndorsements()

		sql := fmt.Sprintf(
			"INSERT INTO restaurants (name, capacity, endorsements, location) VALUES ('%s', '%s', '%s', ST_SetSRID(ST_MakePoint(%f, %f), 4326));",
			name, capacity, fmt.Sprintf(`%q`, endors), lon, lat,
		)

		if stdout {
			logrus.Info(sql)
		} else {
			logrus.Infof("Would execute: %s", sql)
		}
	}
}

func insertDiners(count int, stdout bool) {
	for i := 0; i < count; i++ {
		name := RandomName(rng)
		lat, lon := randomLocation()
		prefs := randomPreferences()

		sql := fmt.Sprintf(
			"INSERT INTO diners (name, preferences, location) VALUES ('%s', '%s', ST_SetSRID(ST_MakePoint(%f, %f), 4326));",
			name, fmt.Sprintf(`%q`, prefs), lon, lat,
		)

		if stdout {
			logrus.Info(sql)
		} else {
			logrus.Infof("Would execute: %s", sql)
		}
	}
}

func main() {
	// Flags to determine mode of operation
	stdout := flag.Bool("stdout", false, "Print SQL statements to stdout instead of executing")
	initdb := flag.Bool("initdb", false, "Initialize the database with test data")
	configFile := flag.String("config", "config.json", "Path to the config file")
	flag.Parse()

	// Configure Logrus
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	// Load configuration
	config, err := loadConfig(*configFile)
	if err != nil {
		logrus.Fatalf("Error loading config: %v", err)
	}

	// Get the DSN from the configuration
	dsn := getDSN(config)

	if *stdout {
		insertRestaurants(100, true)
		insertDiners(500, true)
	} else if *initdb {
		logrus.Info("Initializing database...")

		// Connect to the database
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			logrus.Fatalf("Error connecting to the database: %v", err)
		}
		// Handle error while closing the database connection
		defer func() {
			if err := db.Close(); err != nil {
				logrus.Errorf("Error closing the database connection: %v", err)
			}
		}()

		// Insert data into the database
		insertRestaurants(100, false)
		insertDiners(500, false)
	} else {
		logrus.Warn("Please specify either --stdout or --initdb.")
	}
}

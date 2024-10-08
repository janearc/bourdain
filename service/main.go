package main

import (
	"net/http"
	"strconv"

	"github.com/janearc/bourdain/core"
	"github.com/sirupsen/logrus"
)

func main() {
	// Use absolute path for config since it's inside Docker
	config, err := core.LoadConfig("/config/config.json")
	if err != nil {
		logrus.Fatalf("Could not load config: %v", err)
	}

	// Connect to the database
	db, err := core.ConnectDB(config)
	if err != nil {
		logrus.Fatalf("Could not connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logrus.Fatalf("Error closing database connection: %v", err)
		}
	}()

	// Pass `db` explicitly to handlers
	http.HandleFunc("/restaurant/available", func(w http.ResponseWriter, r *http.Request) {
		restaurantAvailability(w, r, db)
	})
	http.HandleFunc("/restaurant/book", func(w http.ResponseWriter, r *http.Request) {
		restaurantBook(w, r, db)
	})

	// Start the web server using the port from config.json
	port := strconv.Itoa(config.Server.Port)
	logrus.Infof("Server starting on port %s", port)
	logrus.Fatal(http.ListenAndServe(":"+port, nil))
}

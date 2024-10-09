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

	// Define HTTP handlers with closure to pass `db` into handlers
	http.HandleFunc("/restaurant/available", func(w http.ResponseWriter, r *http.Request) {
		restaurantAvailability(w, r, db)
	})

	http.HandleFunc("/restaurant/book", func(w http.ResponseWriter, r *http.Request) {
		restaurantBook(w, r, db)
	})

	// these are private functions which are required for keeping "solution"
	// code out of golang. essentially, the tool that validates the http endpoints
	// feels to me to be "cheating" to have database calls in it. so while it kind
	// of breaks the idea of a "reservation booking system" to have "party building logic"
	// etc in the database, for purposes of having a solution that cleanly addresses
	// the presented problem (build http endpoints), it is the right design choice.

	http.HandleFunc("/private/build_party", func(w http.ResponseWriter, r *http.Request) {
		buildParty(w, r, db)
	})

	// Start the web server using the port from config.json
	port := strconv.Itoa(config.Server.Port)
	logrus.Infof("Server starting on port %s", port)
	logrus.Fatal(http.ListenAndServe(":"+port, nil))
}

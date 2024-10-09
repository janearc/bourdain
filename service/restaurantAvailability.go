package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// restaurantAvailability returns a list of restaurants that can accommodate the number of diners and are open during the specified time
func restaurantAvailability(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Retrieve query parameters from the URL
	dinersStr := r.URL.Query().Get("diners")
	startTimeStr := r.URL.Query().Get("startTime")
	endTimeStr := r.URL.Query().Get("endTime")

	// Convert the number of diners to an integer
	diners, err := strconv.Atoi(dinersStr)
	if err != nil || diners <= 0 {
		http.Error(w, "Invalid number of diners", http.StatusBadRequest)
		return
	}

	// Convert the start and end times to a time.Time object
	startTime, err := time.Parse("15:04", startTimeStr) // Using the "HH:mm" format
	if err != nil {
		http.Error(w, "Invalid start time", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse("15:04", endTimeStr)
	if err != nil {
		http.Error(w, "Invalid end time", http.StatusBadRequest)
		return
	}

	// SQL query to find restaurants that can accommodate the diners and are open during the specified time
	query := `
		SELECT name 
		FROM restaurants 
		WHERE 
			(cast(capacity->>'two-top' as integer) * 2) +
			(cast(capacity->>'four-top' as integer) * 4) +
			(cast(capacity->>'six-top' as integer) * 6) >= $1
			AND opening_time <= $2
			AND closing_time >= $3;`

	rows, err := db.Query(query, diners, startTime.Format("15:04"), endTime.Format("15:04"))
	if err != nil {
		logrus.Errorf("Error querying restaurants: %v", err)
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logrus.Errorf("Error closing rows: %v", err)
		}
	}()

	// Collect the results
	var availableRestaurants []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			logrus.Errorf("Error scanning result: %v", err)
			http.Error(w, "Error scanning result", http.StatusInternalServerError)
			return
		}
		availableRestaurants = append(availableRestaurants, name)
	}

	// Check if there are any available restaurants
	if len(availableRestaurants) == 0 {
		http.Error(w, "No restaurants available for the given number of diners and time window", http.StatusNotFound)
		return
	}

	// Return the available restaurants as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(availableRestaurants); err != nil {
		logrus.Errorf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

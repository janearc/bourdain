package main

import (
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// restaurantAvailability returns a list of restaurants that can accommodate the number of diners and are open during the specified time
func restaurantAvailability(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get query parameters
	dinersStr := r.URL.Query().Get("diners")
	dinersUUIDStr := r.URL.Query().Get("dinerUUIDs")
	startTimeStr := r.URL.Query().Get("start")
	endTimeStr := r.URL.Query().Get("end")

	// Convert parameters to usable types
	diners, err := strconv.Atoi(dinersStr)
	if err != nil || diners <= 0 {
		http.Error(w, "Invalid number of diners", http.StatusBadRequest)
		return
	}
	dinerUUIDs := strings.Split(dinersUUIDStr, ",")
	// Log the raw query parameters
	logrus.Infof("Received start time: %s, end time: %s", startTimeStr, endTimeStr)

	startTime, err := time.Parse("15:04", startTimeStr)
	if err != nil {
		logrus.Errorf("Error parsing start time: %v", err)
		http.Error(w, "Invalid start time", http.StatusBadRequest)
		return
	}
	endTime, err := time.Parse("15:04", endTimeStr)
	if err != nil {
		http.Error(w, "Invalid end time", http.StatusBadRequest)
		return
	}

	// Add validation for UUIDs
	if len(dinerUUIDs) == 0 || dinerUUIDs[0] == "" {
		http.Error(w, "No valid UUIDs provided", http.StatusBadRequest)
		return
	}

	// Logging to help debug any potential issues
	logrus.Infof("Diners: %d, UUIDs: %v, Start: %v, End: %v", diners, dinerUUIDs, startTime, endTime)

	// Execute the SQL query
	query := "SELECT * FROM check_restaurant_availability($1, $2::uuid[], $3, $4);"
	rows, err := db.Query(query, diners, pq.Array(dinerUUIDs), startTime, endTime)
	if err != nil {
		logrus.Errorf("Error querying availability: %v", err)
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logrus.Errorf("Error closing rows: %v", err)
		}
	}()

	// Collect results
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

	// Check if any restaurants were found
	if len(availableRestaurants) == 0 {
		http.Error(w, "No restaurants available for the given parameters", http.StatusNotFound)
		return
	}

	// Return the results as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(availableRestaurants); err != nil {
		logrus.Errorf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

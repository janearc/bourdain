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
	dinersUUIDStr := r.URL.Query().Get("dinersUUID")
	startTimeStr := r.URL.Query().Get("startTime")
	endTimeStr := r.URL.Query().Get("endTime")

	// Convert parameters to usable types
	diners, err := strconv.Atoi(dinersStr)
	if err != nil || diners <= 0 {
		http.Error(w, "Invalid number of diners", http.StatusBadRequest)
		return
	}
	dinerUUIDs := strings.Split(dinersUUIDStr, ",")
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		http.Error(w, "Invalid start time", http.StatusBadRequest)
		return
	}
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		http.Error(w, "Invalid end time", http.StatusBadRequest)
		return
	}

	// Execute the SQL query
	query := "SELECT * FROM check_restaurant_availability($1, $2::uuid[], $3, $4);"
	rows, err := db.Query(query, diners, pq.Array(dinerUUIDs), startTime, endTime)
	if err != nil {
		logrus.Errorf("Error querying availability: %v", err)
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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

	// Return the results as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(availableRestaurants); err != nil {
		logrus.Errorf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

package main

import (
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

// restaurantAvailability returns a list of restaurants that can accommodate the number of diners and are open during the specified time
func restaurantAvailability(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get query parameters
	dinersUUIDStr := r.URL.Query().Get("dinerUUIDs")
	startTimeStr := r.URL.Query().Get("startTime")
	endTimeStr := r.URL.Query().Get("endTime")

	dinerUUIDs := strings.Split(dinersUUIDStr, ",")
	// Log the raw query parameters
	logrus.Infof("Received start time: %s, end time: %s", startTimeStr, endTimeStr)

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		logrus.Errorf("Error parsing start time: %v", err)
		http.Error(w, "Invalid start time", http.StatusBadRequest)
		return
	}
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		logrus.Errorf("Error parsing end time: %v", err)
		http.Error(w, "Invalid end time", http.StatusBadRequest)
		return
	}

	// Add validation for UUIDs
	if len(dinerUUIDs) == 0 || dinerUUIDs[0] == "" {
		http.Error(w, "No valid UUIDs provided", http.StatusBadRequest)
		return
	}

	// Logging to help debug any potential issues
	logrus.Infof("Diners: %v, UUIDs: %v, Start: %v, End: %v", len(dinerUUIDs), dinerUUIDs, startTime, endTime)

	// Execute the SQL query to retrieve restaurant availability
	query := `
		SELECT r.restaurant_id, r.restaurant_name, r.matched_endorsements::text, r.message
		FROM check_restaurant_availability($1::uuid[], $2, $3) AS r;
	`
	rows, err := db.Query(query, pq.Array(dinerUUIDs), startTime, endTime)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "raise_exception" {
			// Treat this as no results found instead of an error
			logrus.Warnf("No restaurants matched the given endorsements: %v", pqErr)
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode([]map[string]string{}); err != nil {
				logrus.Errorf("Error encoding empty response: %v", err)
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}
			return
		}

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
	var availableRestaurants []map[string]string
	for rows.Next() {
		var restaurantID string
		var name string
		var matchedEndorsements string // Holds endorsements in jsonb
		var message string

		if err := rows.Scan(&restaurantID, &name, &matchedEndorsements, &message); err != nil {
			logrus.Errorf("Error scanning result: %v", err)
			http.Error(w, "Error scanning result", http.StatusInternalServerError)
			return
		}
		logrus.Infof("Found restaurant: %s (ID: %s), Endorsements: %s, Message: %s", name, restaurantID, matchedEndorsements, message)

		restaurantInfo := map[string]string{
			"id":                  restaurantID,
			"name":                name,
			"matchedEndorsements": matchedEndorsements,
			"message":             message,
		}
		availableRestaurants = append(availableRestaurants, restaurantInfo)
	}

	// Return a 200 response with an empty list if no restaurants are found
	w.Header().Set("Content-Type", "application/json")
	if len(availableRestaurants) == 0 {
		logrus.Infof("No restaurants available for the given parameters")
		if err := json.NewEncoder(w).Encode([]map[string]string{}); err != nil {
			logrus.Errorf("Error encoding empty response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
		return
	}

	// Return the results as JSON
	if err := json.NewEncoder(w).Encode(availableRestaurants); err != nil {
		logrus.Errorf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

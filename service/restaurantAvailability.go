package main

import (
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
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

	if len(dinerUUIDs) == 0 || dinerUUIDs[0] == "" {
		http.Error(w, "No valid UUIDs provided", http.StatusBadRequest)
		return
	}

	// Execute the SQL query to retrieve restaurant availability
	query := `
		SELECT r.restaurant_id, r.restaurant_name, r.matched_endorsements::text, r.message
		FROM check_restaurant_availability($1::uuid[], $2, $3) AS r;
	`
	rows, err := db.Query(query, pq.Array(dinerUUIDs), startTime, endTime)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "raise_exception" {
			// No restaurants matched the given endorsements
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]string{})
			return
		}
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var availableRestaurants []map[string]string
	for rows.Next() {
		var restaurantID, name, matchedEndorsements, message string
		if err := rows.Scan(&restaurantID, &name, &matchedEndorsements, &message); err != nil {
			http.Error(w, "Error scanning result", http.StatusInternalServerError)
			return
		}
		availableRestaurants = append(availableRestaurants, map[string]string{
			"id":                  restaurantID,
			"name":                name,
			"matchedEndorsements": matchedEndorsements,
			"message":             message,
		})
	}

	// Return the results as JSON
	w.Header().Set("Content-Type", "application/json")
	if len(availableRestaurants) == 0 {
		json.NewEncoder(w).Encode([]map[string]string{})
	} else {
		json.NewEncoder(w).Encode(availableRestaurants)
	}
}

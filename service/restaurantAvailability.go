package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// restaurantAvailability returns a list of restaurants that can accommodate the number of diners
func restaurantAvailability(w http.ResponseWriter, r *http.Request) {
	// startTime := r.URL.Query().Get("startTime") // Not used in this basic example, but could be part of a reservations table
	dinersStr := r.URL.Query().Get("diners")

	// Convert the number of diners to an integer
	diners, err := strconv.Atoi(dinersStr)
	if err != nil || diners <= 0 {
		http.Error(w, "Invalid number of diners", http.StatusBadRequest)
		return
	}

	// Query to find restaurants that can accommodate the specified number of diners
	query := `
		SELECT name 
		FROM restaurants 
		WHERE 
			(cast(capacity->>'two-top' as integer) * 2) +
			(cast(capacity->>'four-top' as integer) * 4) +
			(cast(capacity->>'six-top' as integer) * 6) >= $1;`

	rows, err := db.Query(query, diners)
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
		http.Error(w, "No restaurants available for the given number of diners", http.StatusNotFound)
		return
	}

	// Return the available restaurants as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(availableRestaurants); err != nil {
		logrus.Errorf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

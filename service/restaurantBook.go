package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// restaurantBook reserves a restaurant for the given number of diners
func restaurantBook(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// startTime := r.URL.Query().Get("startTime") // Can be used to track reservation time
	dinersStr := r.URL.Query().Get("diners")
	restaurant := r.URL.Query().Get("restaurant")

	// Convert the number of diners to an integer
	diners, err := strconv.Atoi(dinersStr)
	if err != nil || diners <= 0 {
		http.Error(w, "Invalid number of diners", http.StatusBadRequest)
		return
	}

	// Query to check if the restaurant can accommodate the diners
	query := `
		SELECT 
			(cast(capacity->>'two-top' as integer) * 2) +
			(cast(capacity->>'four-top' as integer) * 4) +
			(cast(capacity->>'six-top' as integer) * 6) AS total_capacity 
		FROM restaurants 
		WHERE name = $1;`

	var totalCapacity int
	err = db.QueryRow(query, restaurant).Scan(&totalCapacity)
	if err == sql.ErrNoRows {
		http.Error(w, "Restaurant not found", http.StatusNotFound)
		return
	} else if err != nil {
		logrus.Errorf("Error querying restaurant: %v", err)
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}

	// Check if the restaurant has enough capacity
	if totalCapacity < diners {
		http.Error(w, "Restaurant cannot accommodate the given number of diners", http.StatusBadRequest)
		return
	}

	// Update the restaurant's capacity after the reservation
	updateQuery := `
		UPDATE restaurants 
		SET 
			capacity = jsonb_set(capacity, '{two-top}', to_jsonb(cast(cast(capacity->>'two-top' as integer) - 1 as text)))
		WHERE name = $1;`
	_, err = db.Exec(updateQuery, restaurant)
	if err != nil {
		logrus.Errorf("Error updating restaurant capacity: %v", err)
		http.Error(w, "Error updating database", http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success", "message": "Reservation booked successfully"}`))
}

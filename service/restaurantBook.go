package main

import (
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
	"net/http"
	"strings"
	"time"
)

// restaurantBook reserves a restaurant for the given number of diners
func restaurantBook(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get query parameters
	startTimeStr := r.URL.Query().Get("startTime")
	endTimeStr := r.URL.Query().Get("endTime")
	dinerUUIDStr := r.URL.Query().Get("dinerUUIDs")
	restaurantUUID := r.URL.Query().Get("restaurantUUID")

	// Validate required parameters
	if startTimeStr == "" || endTimeStr == "" || dinerUUIDStr == "" || restaurantUUID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Parse start and end times
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		http.Error(w, "Invalid start time format", http.StatusBadRequest)
		return
	}
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		http.Error(w, "Invalid end time format", http.StatusBadRequest)
		return
	}

	// Split diner UUIDs into a slice
	dinerUUIDs := strings.Split(dinerUUIDStr, ",")

	// Prepare the SQL call to the stored procedure
	query := `SELECT public.restaurant_book($1::uuid, $2::uuid[], $3::timestamp, $4::timestamp)`

	// Call the stored procedure
	var reservationUUID string
	err = db.QueryRow(query, restaurantUUID, pq.Array(dinerUUIDs), startTime, endTime).Scan(&reservationUUID)
	if err != nil {
		http.Error(w, "Error creating reservation", http.StatusInternalServerError)
		return
	}

	// Respond with the new reservation UUID
	response := map[string]string{
		"status":         "success",
		"reservation_id": reservationUUID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

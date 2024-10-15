package main

import (
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// restaurantBook reserves a restaurant for the given number of diners
func restaurantBook(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get query parameters
	startTimeStr := r.URL.Query().Get("startTime")
	endTimeStr := r.URL.Query().Get("endTime")
	dinerUUIDStr := r.URL.Query().Get("dinerUUIDs")
	restaurantUUID := r.URL.Query().Get("restaurantUUID")

	// Check for missing parameters and log which one is missing
	if startTimeStr == "" {
		logrus.Error("Missing query parameter: startTime")
		http.Error(w, "Missing required parameter: startTime", http.StatusBadRequest)
		return
	}
	if endTimeStr == "" {
		logrus.Error("Missing query parameter: endTime")
		http.Error(w, "Missing required parameter: endTime", http.StatusBadRequest)
		return
	}
	if dinerUUIDStr == "" {
		logrus.Error("Missing query parameter: dinerUUIDs")
		http.Error(w, "Missing required parameter: dinerUUIDs", http.StatusBadRequest)
		return
	}
	if restaurantUUID == "" {
		logrus.Error("Missing query parameter: restaurantUUID")
		http.Error(w, "Missing required parameter: restaurantUUID", http.StatusBadRequest)
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
	query := `
		SELECT public.restaurant_book($1::uuid, $2::uuid[], $3::timestamp, $4::timestamp)`

	// Call the stored procedure
	var reservationUUID string
	err = db.QueryRow(query, restaurantUUID, pq.Array(dinerUUIDs), startTime, endTime).Scan(&reservationUUID)
	if err != nil {
		logrus.Errorf("Error executing stored procedure: %v", err)
		http.Error(w, "Error creating reservation", http.StatusInternalServerError)
		return
	}

	// Respond with success and the new reservation UUID
	response := map[string]string{
		"status":         "success",
		"reservation_id": reservationUUID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Handle error during response encoding
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logrus.Errorf("Error encoding JSON response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

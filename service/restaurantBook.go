package main

import (
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// restaurantBook reserves a restaurant for the given number of diners
func restaurantBook(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get query parameters
	startTimeStr := r.URL.Query().Get("startTime")
	endTimeStr := r.URL.Query().Get("endTime")
	dinersStr := r.URL.Query().Get("diners") // TODO: not real clear on what this is, suspicious
	dinerUUIDStr := r.URL.Query().Get("dinerUUIDs")
	restaurantUUID := r.URL.Query().Get("restaurantUUID")

	// Parse the number of diners
	diners, err := strconv.Atoi(dinersStr)
	if err != nil || diners <= 0 {
		http.Error(w, "Invalid number of diners", http.StatusBadRequest)
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
	if len(dinerUUIDs) != diners {
		http.Error(w, "Number of diners does not match number of diner UUIDs", http.StatusBadRequest)
		return
	}

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

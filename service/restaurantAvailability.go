package main

import (
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

// restaurantAvailability returns a list of restaurants that can accommodate the number of diners and are open during the specified time
func restaurantAvailability(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	dinersStr := r.URL.Query().Get("diners")
	dinersUUIDStr := r.URL.Query().Get("dinersUUID")

	// Convert the diners and UUIDs into a format we can work with
	diners, err := strconv.Atoi(dinersStr)
	if err != nil || diners <= 0 {
		http.Error(w, "Invalid number of diners", http.StatusBadRequest)
		return
	}
	dinerUUIDs := strings.Split(dinersUUIDStr, ",")

	// Fetch endorsements for all diners using the PL/pgSQL function
	var dinerEndorsements []string
	logrus.Infof("Fetching endorsements for party UUIDs: %v", dinerUUIDs)
	query := `SELECT endorsement FROM get_diner_endorsements($1)`
	rows, err := db.Query(query, pq.Array(dinerUUIDs))
	if err != nil {
		logrus.Errorf("Error fetching diner endorsements: %v", err)
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var endorsement string
		if err := rows.Scan(&endorsement); err != nil {
			logrus.Errorf("Error scanning endorsement: %v", err)
			http.Error(w, "Error scanning result", http.StatusInternalServerError)
			return
		}
		dinerEndorsements = append(dinerEndorsements, endorsement)
	}

	// Filter restaurants based on diner endorsements
	query = "SELECT * FROM check_restaurant_availability($1, $2::jsonb);"

	// Convert endorsements to JSONB
	endorsementsJSON, err := json.Marshal(dinerEndorsements)
	if err != nil {
		logrus.Errorf("Error marshaling endorsements JSON: %v", err)
		http.Error(w, "Error encoding endorsements", http.StatusInternalServerError)
		return
	}

	// Query restaurants based on diner size and endorsements
	rows, err = db.Query(query, diners, endorsementsJSON)
	if err != nil {
		logrus.Errorf("Error querying restaurants: %v", err)
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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

	if len(availableRestaurants) == 0 {
		http.Error(w, "No restaurants available for the given number of diners", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(availableRestaurants); err != nil {
		logrus.Errorf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

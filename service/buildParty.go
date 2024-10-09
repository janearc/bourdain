package main

import (
	"database/sql"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func buildParty(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get the party size from the request query parameters
	partySizeStr := r.URL.Query().Get("partySize")

	// Convert the party size to an integer
	partySize, err := strconv.Atoi(partySizeStr)
	if err != nil || partySize <= 0 {
		http.Error(w, "Invalid party size", http.StatusBadRequest)
		return
	}

	// Call the `generate_party` function in the database
	query := `SELECT diner_id FROM generate_party($1)`
	rows, err := db.Query(query, partySize)
	if err != nil {
		logrus.Errorf("Error querying party: %v", err)
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logrus.Errorf("Error closing rows: %v", err)
		}
	}()

	// Collect the results
	var dinerUUIDs []string
	for rows.Next() {
		var dinerID string
		if err := rows.Scan(&dinerID); err != nil {
			logrus.Errorf("Error scanning result: %v", err)
			http.Error(w, "Error scanning result", http.StatusInternalServerError)
			return
		}
		dinerUUIDs = append(dinerUUIDs, dinerID)
	}

	// Return the party UUIDs as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dinerUUIDs); err != nil {
		logrus.Errorf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

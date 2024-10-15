package main

import (
	"database/sql"
	"encoding/json"
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
	query := `SELECT diner_id::text FROM generate_party($1)`
	rows, err := db.Query(query, partySize)
	if err != nil {
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Collect the results
	var dinerUUIDs []string
	for rows.Next() {
		var dinerID string
		if err := rows.Scan(&dinerID); err != nil {
			http.Error(w, "Error scanning result", http.StatusInternalServerError)
			return
		}
		dinerUUIDs = append(dinerUUIDs, dinerID)
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		http.Error(w, "Error processing data", http.StatusInternalServerError)
		return
	}

	// Return the party UUIDs as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dinerUUIDs)
}

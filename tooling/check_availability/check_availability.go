package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"net/http"
	"time"
)

// Create a new random source and generator
var carng = rand.New(rand.NewSource(time.Now().UnixNano()))

// Generates a random number of diners for a party (between 2 and 24)
// Larger parties (greater than 10) will be rare
func generatePartySize() int {
	r := carng.Float64()
	switch {
	case r < 0.7: // 70% chance for small parties (2-6 diners)
		return carng.Intn(5) + 2
	case r < 0.95: // 25% chance for medium parties (7-10 diners)
		return carng.Intn(4) + 7
	default: // 5% chance for large parties (11-24 diners)
		return carng.Intn(14) + 11
	}
}

// randomReservationTime generates a random start time and end time for a reservation, in 15-minute increments
func randomReservationTime() (time.Time, time.Time) {
	// Define the earliest and latest possible times (10:00 AM to 10:00 PM for example)
	openTime := time.Date(0, 0, 0, 10, 0, 0, 0, time.UTC)
	closeTime := time.Date(0, 0, 0, 22, 0, 0, 0, time.UTC)

	// Generate a random number of 15-minute intervals between the open and close time
	totalMinutes := int(closeTime.Sub(openTime).Minutes())
	randomMinutes := rand.Intn(totalMinutes/15) * 15

	// Calculate start and end time based on the random interval
	startTime := openTime.Add(time.Duration(randomMinutes) * time.Minute)
	endTime := startTime.Add(45 * time.Minute) // Assume a 45-minute dining period

	return startTime, endTime
}

// Hit the availability endpoint with the party size and random reservation time
func checkAvailability(partySize int, dinerUUIDs []string, startTime, endTime time.Time) error {
	url := fmt.Sprintf("http://localhost:8080/restaurant/available?diners=%d&start=%s&end=%s",
		partySize, startTime.Format("15:04"), endTime.Format("15:04"))

	// Optionally, we can include diner UUIDs as well if necessary
	// You might want to encode them in the URL or pass them in the request body as JSON
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error hitting availability endpoint: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Errorf("Error closing response body: %v", err)
		}
	}()

	// Read and print the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		logrus.Infof("Party of %d diners found availability: %s", partySize, string(body))
	} else {
		logrus.Infof("Party of %d diners found no availability", partySize)
	}

	return nil
}

func findAvailableRestaurantsWithTime(partySize int, endorsements string, startTime, endTime time.Time, db *sql.DB) ([]string, error) {
	query := `
		SELECT name 
		FROM find_available_restaurants($1, $2::jsonb)
		WHERE opening_time <= $3::time 
		AND closing_time >= $4::time;`

	// Execute the query
	rows, err := db.Query(query, partySize, endorsements, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("error querying database for availability: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logrus.Errorf("Error closing rows: %v", err)
		}
	}()

	// Collect results
	var availableRestaurants []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("error scanning result: %v", err)
		}
		availableRestaurants = append(availableRestaurants, name)
	}

	// Check if any restaurants were found
	if len(availableRestaurants) == 0 {
		return nil, nil
	}

	return availableRestaurants, nil
}

func findAvailableRestaurants(partySize int, endorsements string, db *sql.DB) ([]string, error) {
	query := `
		SELECT name
		FROM find_available_restaurants($1, $2);`

	rows, err := db.Query(query, partySize, endorsements)
	if err != nil {
		return nil, fmt.Errorf("Error querying restaurants: %v", err)
	}
	defer rows.Close()

	var availableRestaurants []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("Error scanning restaurant name: %v", err)
		}
		availableRestaurants = append(availableRestaurants, name)
	}

	return availableRestaurants, nil
}

// create a hypothetical party that we're going to use to assess availability
func buildParty(partySize int) ([]string, error) {
	// Create the URL for the build party endpoint
	url := fmt.Sprintf("http://localhost:8080/private/build_party?partySize=%d", partySize)

	// Make the HTTP GET request to the endpoint
	resp, err := http.Get(url)
	if err != nil {
		logrus.Errorf("Error hitting build party endpoint: %v", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Errorf("Error closing response body: %v", err)
		}
	}()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("Build party endpoint returned status: %s", resp.Status)
		return nil, fmt.Errorf("failed to build party, status code: %d", resp.StatusCode)
	}

	// Parse the response body (JSON array of UUIDs)
	var dinerUUIDs []string
	if err := json.NewDecoder(resp.Body).Decode(&dinerUUIDs); err != nil {
		logrus.Errorf("Error decoding response: %v", err)
		return nil, err
	}

	logrus.Infof("Generated party with UUIDs: %v", dinerUUIDs)
	return dinerUUIDs, nil
}

func main() {
	for i := 1; i <= 5; i++ {
		// Build a random party and generate a random reservation time
		dinerUUIDs, err := buildParty(generatePartySize())
		if err != nil {
			logrus.Errorf("Error building party: %v", err)
			continue
		}

		startTime, endTime := randomReservationTime()
		logrus.Infof("Checking availability for a party from %s to %s...", startTime.Format("15:04"), endTime.Format("15:04"))

		// Call checkAvailability with the correct arguments
		err = checkAvailability(len(dinerUUIDs), dinerUUIDs, startTime, endTime)
		if err != nil {
			logrus.Errorf("Error checking availability: %v", err)
		}

		time.Sleep(1 * time.Second) // Sleep between requests to avoid hammering the server
	}
}

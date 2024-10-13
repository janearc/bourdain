package main

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// Create a new random source and generator
var carng = rand.New(rand.NewSource(time.Now().UnixNano()))

// Generates a random number of diners for a party with adjusted distribution
// 10 parties of 2-4, one party of 6, and one party of 10
func generatePartySize() int {
	r := carng.Float64()
	switch {
	case r < 0.9: // 90% chance for small parties (2-4 diners)
		return carng.Intn(3) + 2
	case r < 0.975: // 7.5% chance for a party of 6
		return 6
	default: // 2.5% chance for a party of 10
		return 10
	}
}

// randomReservationTime generates a random start and end time for a reservation, in 15-minute increments
func randomReservationTime() (time.Time, time.Time) {
	// Generate a random time during the day in 15-minute intervals
	startHour := rand.Intn(24)       // 0 to 23 hours
	startMinute := rand.Intn(4) * 15 // 0, 15, 30, or 45 minutes

	// Use the current date instead of year 0
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), startHour, startMinute, 0, 0, time.UTC)

	// Set the minimum and maximum dining duration (30 to 120 minutes)
	minDuration := 30
	maxDuration := 120
	randomDurationMinutes := rand.Intn((maxDuration-minDuration)/15+1)*15 + minDuration

	// Calculate the end time based on the random duration
	endTime := startTime.Add(time.Duration(randomDurationMinutes) * time.Minute)

	return startTime, endTime
}

// Hit the availability endpoint with the party size and random reservation time
func checkAvailability(dinerUUIDs []string, startTime, endTime time.Time) error {
	// Join the diner UUIDs for the URL query string
	dinerUUIDStr := strings.Join(dinerUUIDs, ",")
	logrus.Infof("Parsed startTime: %v, endTime: %v", startTime, endTime)
	url := fmt.Sprintf("http://localhost:8080/restaurant/available?startTime=%s&endTime=%s&dinerUUIDs=%s",
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
		dinerUUIDStr)

	// Log the request URL before making the HTTP request
	logrus.Infof("Sending request to availability endpoint: %s", url)

	// Make the HTTP GET request to the availability endpoint
	resp, err := http.Get(url)
	if err != nil {
		logrus.Errorf("Error hitting availability endpoint: %v", err)
		return fmt.Errorf("error hitting availability endpoint: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Errorf("Error closing response body: %v", err)
		}
	}()

	// Log the status code of the response
	logrus.Infof("Received response with status code: %d", resp.StatusCode)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Error reading response body: %v", err)
		return fmt.Errorf("error reading response: %v", err)
	}

	// Log the raw response body for debugging
	logrus.Infof("Raw response body: %s", string(body))

	// Handle the response based on the status code
	if resp.StatusCode == http.StatusOK {
		logrus.Infof("Found availability: %s", string(body))
	} else if resp.StatusCode == http.StatusInternalServerError {
		logrus.Errorf("Server error (500) while checking availability [%s]", string(body))
	} else if resp.StatusCode == http.StatusBadRequest {
		logrus.Errorf("Bad request (400) while checking availability [%s]", string(body))
	} else {
		logrus.Warnf("Unexpected status code %d for party [%s]", resp.StatusCode, string(body))
	}

	// Return nil if everything went fine
	return nil
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
	for i := 1; i <= 10; i++ {
		// Build a random party and generate a random reservation time
		dinerUUIDs, err := buildParty(generatePartySize())
		if err != nil {
			logrus.Errorf("Error building party: %v", err)
			continue
		}

		startTime, endTime := randomReservationTime()
		logrus.Infof("Checking availability for a party from %s to %s...", startTime.Format("15:04"), endTime.Format("15:04"))

		// Call checkAvailability with the correct arguments
		err = checkAvailability(dinerUUIDs, startTime, endTime)
		if err != nil {
			logrus.Errorf("Error checking availability: %v", err)
		}

		time.Sleep(1 * time.Second) // Sleep between requests to avoid hammering the server
	}
}

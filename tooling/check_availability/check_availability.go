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

// Struct for a restaurant
type Restaurant struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	MatchedEndorsements string `json:"matchedEndorsements"`
	Message             string `json:"message"`
}

// Struct for a diner
type Diner struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Struct for a reservation
type Reservation struct {
	ID           string   `json:"id"`
	RestaurantID string   `json:"restaurant_id"`
	DinerUUIDs   []string `json:"diner_uuids"`
	StartTime    string   `json:"start_time"`
	EndTime      string   `json:"end_time"`
}

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
	startHour := rand.Intn(24)       // 0 to 23 hours
	startMinute := rand.Intn(4) * 15 // 0, 15, 30, or 45 minutes

	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), startHour, startMinute, 0, 0, time.UTC)

	// Random dining duration (30 to 120 minutes)
	minDuration := 30
	maxDuration := 120
	randomDurationMinutes := rand.Intn((maxDuration-minDuration)/15+1)*15 + minDuration
	endTime := startTime.Add(time.Duration(randomDurationMinutes) * time.Minute)

	return startTime, endTime
}

// Hit the availability endpoint with the party size and random reservation time
func checkAvailability(dinerUUIDs []string, startTime, endTime time.Time) ([]Restaurant, error) {
	dinerUUIDStr := strings.Join(dinerUUIDs, ",")
	logrus.Infof("Parsed startTime: %v, endTime: %v", startTime, endTime)
	url := fmt.Sprintf("http://localhost:8080/restaurant/available?startTime=%s&endTime=%s&dinerUUIDs=%s",
		startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), dinerUUIDStr)

	resp, err := http.Get(url)
	if err != nil {
		logrus.Errorf("Error hitting availability endpoint: %v", err)
		return nil, fmt.Errorf("error hitting availability endpoint: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Error reading response body: %v", err)
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		restaurants := []Restaurant{}
		err := json.Unmarshal(body, &restaurants)
		if err != nil {
			logrus.Errorf("Error decoding response: %v", err)
			return nil, fmt.Errorf("error decoding response: %v", err)
		}
		logrus.Infof("Found availability at %d stores", len(restaurants))

		// Randomly select one restaurant from available
		if len(restaurants) > 0 {
			selectedRestaurant := restaurants[rand.Intn(len(restaurants))]
			logrus.Infof("Randomly selected restaurant: %s", selectedRestaurant.Name)
			return []Restaurant{selectedRestaurant}, nil
		}
	} else {
		logrus.Warnf("Unexpected status code %d for party [%s]", resp.StatusCode, string(body))
	}

	return nil, nil
}

// create a hypothetical party that we're going to use to assess availability
func buildParty(partySize int) ([]string, error) {
	url := fmt.Sprintf("http://localhost:8080/private/build_party?partySize=%d", partySize)

	resp, err := http.Get(url)
	if err != nil {
		logrus.Errorf("Error hitting build party endpoint: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("Build party endpoint returned status: %s", resp.Status)
		return nil, fmt.Errorf("failed to build party, status code: %d", resp.StatusCode)
	}

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
		dinerUUIDs, err := buildParty(generatePartySize())
		if err != nil {
			logrus.Errorf("Error building party: %v", err)
			continue
		}

		startTime, endTime := randomReservationTime()
		logrus.Infof("Checking availability for a party from %s to %s...", startTime.Format("15:04"), endTime.Format("15:04"))

		availableRestaurants, err := checkAvailability(dinerUUIDs, startTime, endTime)
		if err != nil {
			logrus.Errorf("Error checking availability: %v", err)
			continue
		}

		// Now you can use availableRestaurants to call the /restaurant/book endpoint
		if len(availableRestaurants) > 0 {
			// Example of booking logic, you can enhance it
			logrus.Infof("Proceeding to book a reservation at restaurant: %s", availableRestaurants[0].Name)
			// Call /restaurant/book here with availableRestaurants[0]
		}

		time.Sleep(1 * time.Second) // Sleep between requests to avoid hammering the server
	}
}

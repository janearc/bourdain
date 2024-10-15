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

// Struct for restaurant details with correct JSON mapping
type Restaurant struct {
	ID   string `json:"id"` // Ensure this matches the JSON field returned from the API
	Name string `json:"name"`
}

type Reservation struct {
	RestaurantID string   `json:"restaurant_id"`
	DinerUUIDs   []string `json:"diner_uuids"`
	StartTime    string   `json:"start_time"`
	EndTime      string   `json:"end_time"`
}

// Generates a random number of diners for a party
func generatePartySize() int {
	r := carng.Float64()
	switch {
	case r < 0.9:
		return carng.Intn(3) + 2 // 2-4 diners
	case r < 0.975:
		return 6 // 7.5% chance for 6 diners
	default:
		return 10 // 2.5% chance for 10 diners
	}
}

// Generate random reservation times
func randomReservationTime() (time.Time, time.Time) {
	startHour := rand.Intn(24)
	startMinute := rand.Intn(4) * 15
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), startHour, startMinute, 0, 0, time.UTC)

	minDuration := 30
	maxDuration := 120
	randomDurationMinutes := rand.Intn((maxDuration-minDuration)/15+1)*15 + minDuration
	endTime := startTime.Add(time.Duration(randomDurationMinutes) * time.Minute)

	return startTime, endTime
}

// Check availability by hitting the /restaurant/available endpoint
func checkAvailability(dinerUUIDs []string, startTime, endTime time.Time) ([]Restaurant, error) {
	dinerUUIDStr := strings.Join(dinerUUIDs, ",")
	url := fmt.Sprintf("http://localhost:8080/restaurant/available?startTime=%s&endTime=%s&dinerUUIDs=%s",
		startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), dinerUUIDStr)

	logrus.Infof("Checking availability at URL: %s", url)

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

	// Log the raw response body for debugging
	logrus.Infof("Raw response body: %s", string(body))

	if resp.StatusCode == http.StatusOK {
		var restaurants []Restaurant
		err := json.Unmarshal(body, &restaurants)
		if err != nil {
			logrus.Errorf("Error decoding response: %v", err)
			return nil, fmt.Errorf("error decoding response: %v", err)
		}
		logrus.Infof("Found availability at %d restaurants", len(restaurants))

		if len(restaurants) > 0 {
			selectedRestaurant := restaurants[rand.Intn(len(restaurants))]
			logrus.Infof("Randomly selected restaurant: %s (%s)", selectedRestaurant.Name, selectedRestaurant.ID)
			return restaurants, nil
		}
		logrus.Warnf("No restaurants available")
	} else {
		logrus.Warnf("Unexpected status code %d for party [%s]", resp.StatusCode, string(body))
	}

	return nil, nil
}

// Book a reservation via the /restaurant/book endpoint
func bookReservation(reservation Reservation) error {
	if reservation.RestaurantID == "" {
		logrus.Error("RestaurantID is empty. Cannot book reservation.")
		return fmt.Errorf("restaurant ID is missing")
	}

	// Construct URL for the booking request
	url := fmt.Sprintf(
		"http://localhost:8080/restaurant/book?startTime=%s&endTime=%s&dinerUUIDs=%s&restaurantUUID=%s",
		reservation.StartTime, reservation.EndTime, strings.Join(reservation.DinerUUIDs, ","), reservation.RestaurantID,
	)
	logrus.Infof("Booking reservation at URL: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		logrus.Errorf("Error booking reservation: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logrus.Errorf("Failed to book reservation, status code: %d, response: %s", resp.StatusCode, string(body))
		return fmt.Errorf("failed to book reservation, status code: %d", resp.StatusCode)
	}

	logrus.Infof("Successfully booked reservation for restaurant %s", reservation.RestaurantID)
	return nil
}

// Generate a party of diners
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

		// Proceed to book the reservation
		if len(availableRestaurants) > 0 {
			selectedRestaurant := availableRestaurants[0]
			reservation := Reservation{
				RestaurantID: selectedRestaurant.ID,
				DinerUUIDs:   dinerUUIDs,
				StartTime:    startTime.Format(time.RFC3339),
				EndTime:      endTime.Format(time.RFC3339),
			}

			err = bookReservation(reservation)
			if err != nil {
				logrus.Errorf("Error booking reservation: %v", err)
			}
		}

		time.Sleep(1 * time.Second) // Sleep between requests to avoid hammering the server
	}
}

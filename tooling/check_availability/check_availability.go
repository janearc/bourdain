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

type Restaurant struct {
	ID   string `json:"id"`
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
		return 6
	default:
		return 10
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

// Check availability via the /restaurant/available endpoint
func checkAvailability(dinerUUIDs []string, startTime, endTime time.Time) ([]Restaurant, error) {
	dinerUUIDStr := strings.Join(dinerUUIDs, ",")
	url := fmt.Sprintf("http://localhost:8080/restaurant/available?startTime=%s&endTime=%s&dinerUUIDs=%s",
		startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), dinerUUIDStr)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error hitting availability endpoint: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		var restaurants []Restaurant
		err := json.Unmarshal(body, &restaurants)
		if err != nil {
			return nil, fmt.Errorf("error decoding response: %v", err)
		}
		return restaurants, nil
	}

	return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
}

// Book a reservation via the /restaurant/book endpoint
func bookReservation(reservation Reservation) error {
	if reservation.RestaurantID == "" {
		return fmt.Errorf("restaurant ID is missing")
	}

	url := fmt.Sprintf(
		"http://localhost:8080/restaurant/book?startTime=%s&endTime=%s&dinerUUIDs=%s&restaurantUUID=%s",
		reservation.StartTime, reservation.EndTime, strings.Join(reservation.DinerUUIDs, ","), reservation.RestaurantID,
	)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to book reservation, status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Generate a party of diners
func buildParty(partySize int) ([]string, error) {
	url := fmt.Sprintf("http://localhost:8080/private/build_party?partySize=%d", partySize)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to build party, status code: %d", resp.StatusCode)
	}

	var dinerUUIDs []string
	if err := json.NewDecoder(resp.Body).Decode(&dinerUUIDs); err != nil {
		return nil, err
	}

	return dinerUUIDs, nil
}

func main() {
	for i := 1; i <= 10; i++ {
		partySize := generatePartySize()
		dinerUUIDs, err := buildParty(partySize)
		if err != nil {
			logrus.Errorf("Error building party: %v", err)
			continue
		}

		startTime, endTime := randomReservationTime()

		availableRestaurants, err := checkAvailability(dinerUUIDs, startTime, endTime)
		if err != nil {
			logrus.Errorf("Error checking availability: %v", err)
			continue
		}

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
			} else {
				logrus.Infof("Party of %d successfully booked at %s", partySize, selectedRestaurant.Name)
			}
		}

		time.Sleep(1 * time.Second)
	}
}

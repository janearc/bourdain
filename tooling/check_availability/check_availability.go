package main

import (
	"fmt"
	"io/ioutil"
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

// Hit the availability endpoint with the party size
func checkAvailability(partySize int) {
	url := fmt.Sprintf("http://localhost:8080/restaurant/available?diners=%d", partySize)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error hitting availability endpoint: %v\n", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	// Read and print the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Party of %d diners found availability: %s\n", partySize, string(body))
	} else {
		fmt.Printf("Party of %d diners found no availability\n", partySize)
	}
}

func main() {
	for i := 1; i <= 5; i++ {
		partySize := generatePartySize()
		fmt.Printf("Checking availability for party of %d diners...\n", partySize)
		checkAvailability(partySize)
		time.Sleep(1 * time.Second) // Sleep a bit between requests to avoid hammering the server
	}
}

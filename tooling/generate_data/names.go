package main

import (
	"fmt"
	"math/rand"
)

// Exported variables and functions
var (
	firstNames                 = []string{"Chester", "Camden", "Harriet", "Olivia", "Arthur", "Eleanor", "Percival", "Jasper", "Florence", "Theodore", "Boris", "Natasha", "Rocky", "Bullwinkle"}
	lastNames                  = []string{"Hackington", "Perspicacious", "Smith", "Wellington", "Crumble", "Burlington", "Peabody", "Fitzroy", "Wainwright", "Harrington"}
	californiaAdjectives       = []string{"Sunny", "Fresh", "New", "Inverted", "Golden", "Crispy", "Hearty"}
	californiaNouns            = []string{"Sprout", "Avocado", "Eggplant", "Citrus", "Kale", "Sunflower", "Olive", "Quinoa", "Berry"}
	californiaQualifiers       = []string{"on 32nd", "Joy", "Remoulade, Jr.", "Collective", "at the Park", "Kitchen"}
	mediterraneanSeasonings    = []string{"Zaatar", "Turmeric", "Dill", "Rosemary", "Basil", "Mint", "Saffron"}
	mediterraneanVerbs         = []string{"Strolls", "Whispers", "Dances", "Loves", "Sings", "Embraces", "Savors"}
	mediterraneanTypes         = []string{"Taverna", "Dining Cart", "Market", "Cafe", "Haven", "Oasis"}
	fineDiningAdjectives       = []string{"Humble", "New", "Secret", "Ethereal", "Mystic", "Timeless", "Hidden"}
	fineDiningPlaceDescriptors = []string{"on 32nd", "Compromise", "Watering Hole", "Gastronomy", "Transcendence", "Retreat"}
	frenchPhrases              = []string{"Le Rêve", "Maison", "Cuisine", "Gourmand", "Savoureux"}
	italianPhrases             = []string{"La Vita", "Il Gusto", "Osteria", "Bontà"}
	endorsementWeights         = map[string]float64{
		"gluten-free":          0.15,
		"kid-friendly":         0.15,
		"paleo":                0.05,
		"vegan":                0.05,
		"organic":              0.20,
		"halal":                0.10,
		"kosher":               0.05,
		"pet-friendly":         0.15,
		"molecular-gastronomy": 0.05,
	}
)

// RandomName generates a random name using the provided random number generator
func RandomName(rng *rand.Rand) string {
	firstName := firstNames[rng.Intn(len(firstNames))]
	lastName := lastNames[rng.Intn(len(lastNames))]
	return fmt.Sprintf("%s %s", firstName, lastName)
}

// RandomRestaurantName generates a random restaurant name using the provided random number generator
func RandomRestaurantName(rng *rand.Rand) string {
	vibe := rng.Intn(3)
	switch vibe {
	case 0: // California cuisine
		adj := californiaAdjectives[rng.Intn(len(californiaAdjectives))]
		noun := californiaNouns[rng.Intn(len(californiaNouns))]
		qualifier := ""
		if rng.Float32() < 0.5 {
			qualifier = " " + californiaQualifiers[rng.Intn(len(californiaQualifiers))]
		}
		return fmt.Sprintf("%s %s%s", adj, noun, qualifier)
	case 1: // Mediterranean
		seasoning := mediterraneanSeasonings[rng.Intn(len(mediterraneanSeasonings))]
		verb := mediterraneanVerbs[rng.Intn(len(mediterraneanVerbs))]
		typ := ""
		if rng.Float32() < 0.5 {
			typ = " " + mediterraneanTypes[rng.Intn(len(mediterraneanTypes))]
		}
		return fmt.Sprintf("%s %s%s", seasoning, verb, typ)
	case 2: // Fine dining
		phrase := frenchPhrases[rng.Intn(len(frenchPhrases))]
		if rng.Float32() < 0.5 {
			phrase = italianPhrases[rng.Intn(len(italianPhrases))]
		}
		name := firstNames[rng.Intn(len(firstNames))]
		adj := fineDiningAdjectives[rng.Intn(len(fineDiningAdjectives))]
		placeDescriptor := fineDiningPlaceDescriptors[rng.Intn(len(fineDiningPlaceDescriptors))]
		return fmt.Sprintf("%s's %s %s %s", name, adj, phrase, placeDescriptor)
	default:
		return "Unnamed Eatery"
	}
}

// randomLocation generates a random latitude and longitude within the bounds of Manhattan, Brooklyn, and Bronx.
func randomLocation() (float64, float64) {
	// Define rough latitude and longitude bounds for Manhattan, Brooklyn, and Bronx
	latBounds := [2]float64{40.5774, 40.9176} // Approx latitude range
	lonBounds := [2]float64{-74.15, -73.7004} // Approx longitude range

	// Generate random latitude and longitude within the defined bounds
	lat := latBounds[0] + rng.Float64()*(latBounds[1]-latBounds[0])
	lon := lonBounds[0] + rng.Float64()*(lonBounds[1]-lonBounds[0])

	return lat, lon
}

// randomEndorsements selects endorsements based on defined weights in endorsementWeights.
func randomEndorsements() []string {
	// Create a slice of endorsements where each appears multiple times based on its weight
	allEndorsements := make([]string, 0)
	for endorsement, weight := range endorsementWeights {
		count := int(weight * 100) // Scale the weight to an integer (out of 100)
		for i := 0; i < count; i++ {
			allEndorsements = append(allEndorsements, endorsement)
		}
	}

	// Shuffle the endorsements slice to ensure randomness
	rng.Shuffle(len(allEndorsements), func(i, j int) {
		allEndorsements[i], allEndorsements[j] = allEndorsements[j], allEndorsements[i]
	})

	// Randomly select 1-3 endorsements from the weighted list
	numEndorsements := rng.Intn(3) + 1 // Random number between 1 and 3
	selected := make([]string, numEndorsements)

	for i := 0; i < numEndorsements; i++ {
		selected[i] = allEndorsements[rng.Intn(len(allEndorsements))]
	}

	return selected
}

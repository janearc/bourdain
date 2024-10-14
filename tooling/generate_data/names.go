package main

import (
	"fmt"
	"math/rand"
)

// Exported variables and functions
var (
	Endorsements               = []string{"gluten-free", "kid-friendly", "paleo", "vegan", "organic", "halal", "kosher", "pet-friendly", "molecular-gastronomy"}
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

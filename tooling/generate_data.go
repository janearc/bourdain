package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var (
	endorsements               = []string{"gluten-free", "kid-friendly", "paleo", "vegan", "organic", "halal", "kosher"}
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
)

// Initialize a local random generator
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomName() string {
	firstName := firstNames[rng.Intn(len(firstNames))]
	lastName := lastNames[rng.Intn(len(lastNames))]
	return fmt.Sprintf("%s %s", firstName, lastName)
}

func randomRestaurantName() string {
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

func randomLocation() (float64, float64) {
	latBounds := [2]float64{40.5774, 40.9176} // Rough latitude range for Manhattan, Brooklyn, Bronx
	lonBounds := [2]float64{-74.15, -73.7004} // Rough longitude range

	lat := latBounds[0] + (rand.Float64() * (latBounds[1] - latBounds[0]))
	lon := lonBounds[0] + (rand.Float64() * (lonBounds[1] - lonBounds[0]))
	return lat, lon
}

func randomEndorsements() []string {
	num := rand.Intn(3) + 1
	var selected []string
	for i := 0; i < num; i++ {
		selected = append(selected, endorsements[rand.Intn(len(endorsements))])
	}
	return selected
}

func randomPreferences() []string {
	return randomEndorsements()
}

func insertRestaurants(count int, stdout bool) {
	for i := 0; i < count; i++ {
		name := randomRestaurantName()
		lat, lon := randomLocation()
		capacity := fmt.Sprintf(`{"two-top": %d, "four-top": %d, "six-top": %d}`, rand.Intn(10)+1, rand.Intn(10)+1, rand.Intn(5)+1)
		endors := randomEndorsements()

		sql := fmt.Sprintf(
			"INSERT INTO restaurants (name, capacity, endorsements, location) VALUES ('%s', '%s', '%s', ST_SetSRID(ST_MakePoint(%f, %f), 4326));",
			name, capacity, fmt.Sprintf(`%q`, endors), lon, lat,
		)

		if stdout {
			logrus.Info(sql)
		} else {
			// You would replace this with actual database insertion logic when not in stdout mode
			logrus.Infof("Would execute: %s", sql)
		}
	}
}

func insertDiners(count int, stdout bool) {
	for i := 0; i < count; i++ {
		name := randomName()
		lat, lon := randomLocation()
		prefs := randomPreferences()

		sql := fmt.Sprintf(
			"INSERT INTO diners (name, preferences, location) VALUES ('%s', '%s', ST_SetSRID(ST_MakePoint(%f, %f), 4326));",
			name, fmt.Sprintf(`%q`, prefs), lon, lat,
		)

		if stdout {
			logrus.Info(sql)
		} else {
			// You would replace this with actual database insertion logic when not in stdout mode
			logrus.Infof("Would execute: %s", sql)
		}
	}
}

func main() {
	// Flags to determine mode of operation
	stdout := flag.Bool("stdout", false, "Print SQL statements to stdout instead of executing")
	initdb := flag.Bool("initdb", false, "Initialize the database with test data")
	flag.Parse()

	// Configure Logrus
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	// Use the local rng to generate data
	if *stdout {
		insertRestaurants(100, true)
		insertDiners(500, true)
	} else if *initdb {
		logrus.Info("Initializing database...")
		insertRestaurants(100, false)
		insertDiners(500, false)
	} else {
		logrus.Warn("Please specify either --stdout or --initdb.")
	}
}

# bourdain

take-home exercise for stephen

# problem

given a set of restaurants with attributes:

```json
{
  "restaurants": [
    {
      "name": "string",
      "capacity": {
          "two-top":  1,    // integer
          "four-top": 2,    // integer,
          "six-top":  3,    // integer
      },
      "endorsements": [
        "gluten-free",
        "kid-friendly",
        "paleo"
      ],
      "location": [  // [ latitude, longitude ]
        37.7749,
        -122.4194
      ]
    }
  ]
}
```

and a set of diners with attributes:

```json
{
  "diners": [
    {
      "name": "string",
      "location": [  // [ latitude, longitude ]
        37.7749,
        -122.4194
      ],
      "preferences": [
        "gluten-free",
        "kid-friendly",
        "paleo"
      ]
    }
  ]
}
```

create two endpoints:

```golang
// restaurantAvailability returns a list of restaurants that can accommodate the number of diners at the given time
http.HandleFunc("/restaurant/available", func(w http.ResponseWriter, r *http.Request) {
	// reservations are assumed to be two hours
	startTime := r.URL.Query().Get("startTime")
	// how you implement this is up to you
	diners := r.URL.Query().Get("diners")
})
```

```golang
// restaurantBook reserves the correct number of tables at the given restaurant
http.HandleFunc("/restaurant/available", func(w http.ResponseWriter, r *http.Request) {
    // reservations are assumed to be two hours
    startTime := r.URL.Query().Get("startTime")
    // how you implement this is up to you
    diners := r.URL.Query().Get("diners")
    // how you implement this is up to you
    restaurant := r.URL.Query().Get("restaurant")
})
```

Only API is in scope; UI is out of scope. You may use a database. The solution does not need to be publicly deployed
but should be developed in a manner consistent with production code. Language is your choice but please do not use
befunge.
#!/bin/bash

# TOOLING defines the path for the test scripts
TOOLING="./tooling/verify_solution_shell"

SIZE=3

# Chain the scripts together:
# 1. Call test_restaurant_availability.sh to get available restaurants for a party of 3
# 2. Pass the output to random_restaurant.sh to pick a random restaurant from the list
# 3. Finally, pass that data to book_restaurant.sh to make the booking
${TOOLING}/test_restaurant_availability.sh ${SIZE} "2024-10-14T18:00:00Z" "2024-10-14T20:00:00Z" | \
  while IFS= read -r line; do
    # Extract the restaurant UUID and diner UUIDs using sed
    RESTAURANT_UUID=$(echo "$line" | sed -n 's/^Available restaurants: //p')
    DINER_UUIDS=$(echo "$line" | sed -n 's/^Diner UUIDs: //p')

    # Check if we have both the restaurant UUID and diner UUIDs
    if [[ -n "$RESTAURANT_UUID" && -n "$DINER_UUIDS" ]]; then
      # Now, use the available data to book the restaurant
      echo "Restaurant UUID: $RESTAURANT_UUID"
      echo "Diner UUIDs: $DINER_UUIDS"
      ${TOOLING}/book_restaurant.sh "$RESTAURANT_UUID" "$DINER_UUIDS" "2024-10-14T18:00:00Z" "2024-10-14T20:00:00Z"
    fi
  done
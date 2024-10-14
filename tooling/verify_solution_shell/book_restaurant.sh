#!/bin/bash

# Check if necessary arguments are provided
if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ] || [ -z "$4" ]; then
  echo "Usage: $0 <restaurantUUID> <dinerUUIDs> <start_time> <end_time>"
  exit 1
fi

# Set the necessary arguments
RESTAURANT_UUID=$1
DINER_UUIDS=$2
STARTTIME=$3
ENDTIME=$4

# Debugging output to verify the provided inputs
echo "Attempting to book restaurant with UUID: $RESTAURANT_UUID"
echo "Diner UUIDs: $DINER_UUIDS"
echo "Start time: $STARTTIME"
echo "End time: $ENDTIME"

# Execute the curl request to /restaurant/book endpoint
RESPONSE=$(curl -s -X POST http://localhost:8080/restaurant/book \
  -H "Content-Type: application/json" \
  -d "{\"restaurantUUID\":\"$RESTAURANT_UUID\", \"dinerUUIDs\":[$DINER_UUIDS], \"startTime\":\"$STARTTIME\", \"endTime\":\"$ENDTIME\"}")

# Check if the curl request was successful
if [ $? -ne 0 ]; then
  echo "Error calling /restaurant/book endpoint"
  exit 1
fi

# Output the response (UUID of the booked reservation)
echo "Response from restaurant/book endpoint:"
echo "$RESPONSE"
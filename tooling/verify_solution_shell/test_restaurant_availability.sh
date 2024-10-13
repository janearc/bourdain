#!/bin/bash

# Check if partysize, start_time, and end_time are provided as arguments
if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
  echo "Usage: $0 <partysize> <start_time> <end_time>"
  exit 1
fi

# Set the partysize from the first argument
PARTYSIZE=$1
STARTTIME=$2
ENDTIME=$3

# Get the diner UUIDs from the build_party script
DINERUUIDS=$(./tooling/verify_solution_shell/build_party.sh "$PARTYSIZE")

# Check if the build_party script executed successfully
if [ $? -ne 0 ]; then
  echo "Error calling build_party.sh"
  exit 1
fi

# Extract only the UUIDs by stripping out any extra text before the actual JSON array
DINERUUIDS=$(echo "$DINERUUIDS" | sed -n 's/.*\[\(.*\)\].*/\1/p')

# Format the dinerUUIDs as a comma-separated string for the curl request
DINERUUIDS_CSV=$(echo "$DINERUUIDS" | tr -d '"' | tr ',' ',')

# Debugging output to verify the formatted UUIDs
echo "Attempting to feed dinerUUIDs: $DINERUUIDS_CSV"

# Execute the curl request directly without eval
RESPONSE=$(curl -s "http://localhost:8080/restaurant/available?dinerUUIDs=${DINERUUIDS_CSV}&startTime=${STARTTIME}&endTime=${ENDTIME}")

# Check if curl executed successfully
if [ $? -ne 0 ]; then
  echo "Error calling /restaurant/available endpoint"
  exit 1
fi

# Output the response (list of available restaurants in JSON format)
echo "Response from restaurant/available endpoint:"
echo "$RESPONSE"
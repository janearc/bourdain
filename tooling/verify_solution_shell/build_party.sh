#!/bin/bash

# Check if partysize is provided as an argument
if [ -z "$1" ]; then
  echo "Usage: $0 <partysize>"
  exit 1
fi

# Set the partysize from the first argument
PARTYSIZE=$1

# Call the /private/build_party endpoint with the partysize
RESPONSE=$(curl -s "http://localhost:8080/private/build_party?partySize=${PARTYSIZE}")

# Check if curl executed successfully
if [ $? -ne 0 ]; then
  echo "Error calling /private/build_party endpoint"
  exit 1
fi

# Output the response (list of dinerUUIDs in JSON format)
echo "Response from build_party endpoint:"
echo "$RESPONSE"
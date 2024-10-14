#!/bin/bash

# Read diner UUIDs and available restaurants from stdin
read DINER_UUIDS
read RESTAURANTS

# Extract the actual values from the prefixed lines
DINER_UUIDS=$(echo "$DINER_UUIDS" | sed 's/Diner UUIDs: //')
RESTAURANTS=$(echo "$RESTAURANTS" | sed 's/Available restaurants: //')

# Use jq to pick a random restaurant from the JSON array
SELECTED_RESTAURANT=$(echo "$RESTAURANTS" | jq -r '.[|random]')

# Output the selected restaurant and diner UUIDs
echo "$SELECTED_RESTAURANT $DINER_UUIDS"
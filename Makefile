# Define base path to tooling directory
SCRIPT_PATH=tooling

# Run the script to generate test data to stdout
testdata:
	go run $(SCRIPT_PATH)/generate_data.go --stdout

# Run the script to initialize the database with data
initdb:
	go run $(SCRIPT_PATH)/generate_data.go --initdb
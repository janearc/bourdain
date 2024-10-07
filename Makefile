# Define paths and flags
SCRIPT_PATH=tooling

# Run the script to generate test data to stdout
testdata:
	go run $(SCRIPT_PATH)/generate_data.go --stdout

# Build the Docker images for the app and Postgres
build:
	docker-compose build

# Spin up Postgres and initialize the database with data
initdb: build
	docker-compose up app

# Clean up containers and volumes
clean:
	docker-compose down -v
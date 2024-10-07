# Define paths and flags
SCRIPT_PATH=tooling

# Create the .env file from config.json using jq
generate-env:
	echo "POSTGRES_USER=$(shell jq -r '.database.user' config.json)" > .env
	echo "POSTGRES_PASSWORD=$(shell jq -r '.database.password' config.json)" >> .env
	echo "POSTGRES_DB=$(shell jq -r '.database.dbname' config.json)" >> .env

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

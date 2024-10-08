# Define paths and flags
SCRIPT_PATH=tooling

generate-env:
	echo "POSTGRES_USER=$(shell jq -r '.database.user' config.json)" > .env
	echo "POSTGRES_PASSWORD=$(shell jq -r '.database.password' config.json)" >> .env
	echo "POSTGRES_DB=$(shell jq -r '.database.dbname' config.json)" >> .env

# Build the Docker images for the app and Postgres
build:
	docker-compose build

fresh: clean build initdb verify-db

# Spin up the app service and insert data into the database
initdb: build
	# Start the db service
	docker-compose up -d db
	# Wait for the database to be healthy
	@echo "Waiting for database to be healthy..."
	@until [ "$$(docker inspect --format='{{.State.Health.Status}}' bourdain-db-1)" = "healthy" ]; do \
		sleep 2; \
		echo "Waiting..."; \
	done
	# Run the app container to insert data into the running db container
	docker-compose run --rm app /usr/local/bin/generate_data --initdb --config=/config/config.json

verify-db:
	@echo "Verifying database schema..."
	docker-compose exec db psql -U $(shell jq -r '.database.user' config.json) -d $(shell jq -r '.database.dbname' config.json) -c "\dt" | grep -q "restaurants" && echo "Database is healthy and schema is present." || (echo "Database verification failed." && exit 1)

# Run the app to print SQL statements to stdout (test mode)
testinit: build
	docker-compose run --rm app /usr/local/bin/generate_data --stdout --config=/config/config.json

# Clean up containers and volumes
clean:
	docker-compose down -v
	docker-compose down --remove-orphans
	rm checkAvailability

# Generate random diner and restaurant names and print them
names: build
	@diner=$$(docker-compose run --rm app /usr/local/bin/generate_data --proper-name | tail -n 1) && \
	restaurant=$$(docker-compose run --rm app /usr/local/bin/generate_data --restaurant-name | tail -n 1) && \
	echo "$$diner requests a reservation at $$restaurant"

# Build the checkAvailability binary
checkAvailability:
	go build -o checkAvailability ./tooling/check_availability

# Run the checkAvailability binary and start the app if not running
runCheckAvailability: checkAvailability
	@docker-compose ps | grep app | grep "Up" > /dev/null || (echo "Starting app service..." && docker-compose up -d app)
	@echo "Checking availability..."
	./checkAvailability

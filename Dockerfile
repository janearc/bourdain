# Stage 1: Build the Go applications
FROM golang:1.23-alpine AS build

WORKDIR /app

# Install jq for JSON parsing
RUN apk add --no-cache jq

# Copy go.mod and go.sum files to install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code into the container
COPY . .

# Build the Go application for generate_data
RUN go build -o /usr/local/bin/generate_data ./tooling/generate_data

# Build the Go application for check_availability
RUN go build -o /usr/local/bin/check_availability ./tooling/check_availability/check_availability.go

# Stage 2: Final image with PostgreSQL and the Go applications
FROM postgres:15-alpine

# Copy the Go binaries from the previous stage
COPY --from=build /usr/local/bin/generate_data /usr/local/bin/generate_data
COPY --from=build /usr/local/bin/check_availability /usr/local/bin/check_availability

# Copy the config file to /config in the container
COPY config.json /config/config.json

# Expose the PostgreSQL port
EXPOSE 5432

# Default command to start PostgreSQL
CMD ["docker-entrypoint.sh", "postgres"]

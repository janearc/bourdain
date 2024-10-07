# Stage 1: Build the Go application
FROM golang:1.20-alpine AS build

WORKDIR /app

# Install jq for JSON parsing
RUN apk add --no-cache jq

# Copy the go.mod and go.sum files to install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code into the container
COPY . .

# Build the Go application
RUN go build -o generate_data tooling/generate_data.go

# Stage 2: Final image with PostgreSQL and the Go application
FROM postgres:15-alpine

# Copy the Go binary from the previous stage
COPY --from=build /app/generate_data /usr/local/bin/generate_data

# Copy the config file
COPY config.json /usr/local/bin/config.json

# Expose the PostgreSQL port
EXPOSE 5432

# Default command to start PostgreSQL
CMD ["docker-entrypoint.sh", "postgres"]
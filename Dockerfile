# Stage 1: Build the Go application
FROM golang:1.23-alpine AS build

WORKDIR /app

# Install jq for JSON parsing
RUN apk add --no-cache jq

# Copy go.mod and go.sum files to install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code into the container
COPY . .

# Build the Go application (web_service and generate_data)
RUN go build -o /usr/local/bin/web_service ./service
RUN go build -o /usr/local/bin/check_availability ./tooling/check_availability
RUN go build -o /usr/local/bin/generate_data ./tooling/generate_data

# Stage 2: Run the Go application
FROM alpine:latest

# Copy the Go binaries from the previous stage
COPY --from=build /usr/local/bin/generate_data /usr/local/bin/generate_data
COPY --from=build /usr/local/bin/check_availability /usr/local/bin/check_availability
COPY --from=build /usr/local/bin/web_service /usr/local/bin/web_service

# Copy the config file to /config in the container
COPY config.json /config/config.json
COPY tooling/queries /config/queries

# Expose the app port
EXPOSE 8080

# Set the entry point for the web service
CMD ["/usr/local/bin/web_service"]

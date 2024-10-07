# Stage 1: Build the Go application
FROM golang:1.20-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o generate_data tooling/generate_data.go

# Stage 2: PostgreSQL with the Go application
FROM postgres:15-alpine

ENV POSTGRES_USER=myuser
ENV POSTGRES_PASSWORD=mypassword
ENV POSTGRES_DB=mydatabase

# Copy the Go binary to /usr/local/bin
COPY --from=build /app/generate_data /usr/local/bin/generate_data

# Set the working directory for the application (optional, if needed by the app)
WORKDIR /usr/local/bin

# Expose the PostgreSQL port
EXPOSE 5432

# Default command to run PostgreSQL
CMD ["docker-entrypoint.sh", "postgres"]
# Stage 1: Build the Go application
FROM golang:1.20-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to install dependencies
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the entire source code into the container
COPY . .

# Build the Go application
RUN go build -o generate_data tooling/generate_data.go

# Stage 2: PostgreSQL with the Go application
FROM postgres:15-alpine

# Set environment variables for PostgreSQL
ENV POSTGRES_USER=myuser
ENV POSTGRES_PASSWORD=mypassword
ENV POSTGRES_DB=mydatabase

# Copy the compiled Go application from the previous stage
COPY --from=build /app/generate_data /usr/local/bin/generate_data

# Copy an SQL initialization script into the Postgres container (optional)
# If you have an init.sql script, uncomment the next line
# COPY tooling/init.sql /docker-entrypoint-initdb.d/

# Expose PostgreSQL and application ports
EXPOSE 5432 8080

# Command to start PostgreSQL and run the Go application
CMD ["sh", "-c", "docker-entrypoint.sh postgres & generate_data --initdb"]
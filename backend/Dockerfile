# Use Go 1.23 as the base image
FROM golang:1.23 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to install dependencies
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy the entire backend project into the container
COPY . .  

# Build the Go application and output the binary to /app/tmp/main
RUN go build -o tmp/main cmd/server/main.go

# Install Air for live reloading
RUN go install github.com/air-verse/air@latest

# Command to start Air for hot reloading (development)
CMD ["air"]

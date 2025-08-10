# Use Go 1.23 as the base image
FROM golang:1.24

# Set the working directory inside the container
WORKDIR /app

# Install Air for live reloading
RUN go install github.com/air-verse/air@latest

# Copy only go.mod and go.sum first (for caching dependencies)
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Expose backend port
EXPOSE 3001

# Run Air for hot reloading
CMD ["air", "-c", "air.toml"]

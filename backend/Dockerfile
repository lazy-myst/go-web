# Use Go 1.23 as the base image
FROM golang:1.24

# Set the working directory inside the container
WORKDIR /app

# Install Air for live reloading (pinned: newer releases require Go >= 1.25/1.26)
RUN go install github.com/air-verse/air@v1.62.0

# Install protoc + Go plugins for gRPC code generation
RUN apt-get update && apt-get install -y --no-install-recommends protobuf-compiler && rm -rf /var/lib/apt/lists/*
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

# Copy only go.mod and go.sum first (for caching dependencies)
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Expose backend port
EXPOSE 3001

# Run Air for hot reloading
CMD ["air", "-c", "air.toml"]

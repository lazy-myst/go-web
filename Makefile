# Define directories
BACKEND_DIR=backend
FRONTEND_DIR=frontend

# Default target
.PHONY: all build docker-build docker-run docker-stop docker-restart docker-clean help

## Build and run everything with Docker Compose
all: docker-run  ## Default: Run all services with Docker Compose

## Build backend and frontend (non-Docker)
build:  ## Build backend and frontend
	@echo "Building Backend..."
	cd $(BACKEND_DIR) && go build -o bin/server cmd/server/main.go
	@echo "Building Frontend..."
	cd $(FRONTEND_DIR) && npm install && npm run build

## Build Docker images for backend and frontend
docker-build:  ## Build Docker images
	@echo "Building Docker Images..."
	docker-compose build

## Run backend, frontend, and MongoDB with Docker Compose
docker-run:  ## Run all services with Docker Compose
	@echo "Starting Docker Compose services..."
	docker-compose up --build -d

## Stop all Docker containers
docker-stop:  ## Stop Docker containers
	@echo "Stopping Docker containers..."
	docker-compose down

## Restart Docker containers
docker-restart:  ## Restart Docker containers
	@echo "Restarting Docker containers..."
	docker-compose down && docker-compose up --build -d

## Remove all Docker images, containers, and volumes
docker-clean:  ## Remove Docker containers, images, and volumes
	@echo "Removing Docker containers, images, and volumes..."
	docker-compose down --rmi all --volumes --remove-orphans

## Show available commands
help:  ## Show available commands
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
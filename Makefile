# Define directories
BACKEND_DIR=backend
FRONTEND_DIR=frontend

# Commands
.PHONY: all backend frontend run stop clean build docker-build docker-run docker-stop docker-restart docker-remove help

## Run both frontend and backend
all: backend frontend

## Run the backend (Go) with Air for hot-reloading
backend:
	@echo "Starting Backend with Air (Hot Reloading)..."
	cd $(BACKEND_DIR) && air  

## Run the frontend (Vite)
frontend:
	@echo "Starting Frontend with Vite..."
	cd $(FRONTEND_DIR) && npm install && npm run dev  

## Run both backend and frontend in parallel
run:
	@echo "Starting Backend and Frontend..."
	make -j2 backend frontend  

## Stop all running services (if needed)
stop:
	@echo "Stopping Backend and Frontend..."
	pkill -f "air" || true  
	pkill -f "npm run dev" || true  

## Clean up build artifacts
clean:
	@echo "Cleaning up build files..."
	cd $(BACKEND_DIR) && go clean  
	cd $(FRONTEND_DIR) && rm -rf node_modules dist  
	rm -rf $(BACKEND_DIR)/tmp $(BACKEND_DIR)/bin 

## Build the backend
build-backend:
	@echo "Building Backend..."
	cd $(BACKEND_DIR) && go build -o bin/server cmd/server/main.go  

## Build the frontend
build-frontend:
	@echo "Building Frontend..."
	cd $(FRONTEND_DIR) && npm install && npm run build  

## Build both frontend and backend
build: build-backend build-frontend  # Build both backend and frontend

## 🐳 Build the backend Docker image
docker-build-backend:
	@echo "Building Backend Docker Image..."
	docker build -t go-backend -f $(BACKEND_DIR)/Dockerfile $(BACKEND_DIR)  

## 🐳 Build the frontend Docker image
docker-build-frontend:
	@echo "Building Frontend Docker Image..."
	docker build -t react-frontend -f $(FRONTEND_DIR)/Dockerfile $(FRONTEND_DIR)  

## 🐳 Build both frontend and backend Docker images
docker-build: docker-build-backend docker-build-frontend  

## 🐳 Run the backend in a Docker container
docker-run-backend:
	@echo "Running Backend Docker Container..."
	docker run -d -p 3001:3001 --env-file=$(BACKEND_DIR)/.env go-backend  

## 🐳 Run the frontend in a Docker container
docker-run-frontend:
	@echo "Running Frontend Docker Container..."
	docker run -d -p 5173:5173 react-frontend  

## 🐳 Run backend, frontend, and MongoDB using Docker Compose
docker-run:
	@echo "Starting Docker Compose services..."
	docker-compose up --build -d  

## 🛑 Stop all Docker containers (backend, frontend, MongoDB)
docker-stop:
	@echo "Stopping all running Docker containers..."
	docker-compose down 

## 🔄 Restart Docker containers for frontend and backend
docker-restart:
	@echo "Stopping and restarting Backend Docker Container..."
	docker-compose down && docker-compose up --build -d

## ❌ Remove all Docker images and containers
docker-remove:
	@echo "Stopping and removing all Docker containers and images..."
	docker-compose down --rmi all --volumes --remove-orphans 

## 📄 Show available Makefile commands
help:
	@echo "Available Makefile commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'  

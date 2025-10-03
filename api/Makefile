# Waiting Room System Makefile
# =============================

# Variables
OPENAPI=open-api.yaml
PKG=api
OUT=internal/api
DOCKER_COMPOSE=docker-compose.yml
DOCKER_COMPOSE_DEV=docker-compose.dev.yml
CONFIG_PATH=config.yaml

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# =============================
# API Development Commands
# =============================

.PHONY: gen
gen:
	@echo "$(BLUE)Generating API code from OpenAPI specification...$(NC)"
	@service-generator -wd ./ -open-api ./$(OPENAPI)

.PHONY: build
build:
	@echo "$(BLUE)Building API server...$(NC)"
	@go build -o api ./cmd/api

.PHONY: run
run:
	@echo "$(BLUE)Starting API server...$(NC)"
	@go run ./cmd/api

.PHONY: run-dev
run-dev:
	@echo "$(BLUE)Starting API server in development mode...$(NC)"
	@CONFIG_PATH=$(CONFIG_PATH) go run ./cmd/api

.PHONY: tidy
tidy:
	@echo "$(BLUE)Cleaning up Go modules...$(NC)"
	@go mod tidy

.PHONY: test
test:
	@echo "$(BLUE)Running API tests...$(NC)"
	@go test ./...

.PHONY: clean-api
clean-api:
	@echo "$(YELLOW)Cleaning API build artifacts...$(NC)"
	@rm -f api
	@go clean

# =============================
# Frontend Development Commands
# =============================

.PHONY: install-ui
install-ui:
	@echo "$(BLUE)Installing UI dependencies...$(NC)"
	@cd ../ui && npm install

.PHONY: build-ui
build-ui:
	@echo "$(BLUE)Building all UI applications...$(NC)"
	@cd ../ui && ng build kiosk
	@cd ../ui && ng build mobile
	@cd ../ui && ng build tv
	@cd ../ui && ng build backoffice
	@cd ../ui && ng build api-client

.PHONY: build-ui-kiosk
build-ui-kiosk:
	@echo "$(BLUE)Building Kiosk application...$(NC)"
	@cd ../ui && ng build kiosk

.PHONY: build-ui-mobile
build-ui-mobile:
	@echo "$(BLUE)Building Mobile application...$(NC)"
	@cd ../ui && ng build mobile

.PHONY: build-ui-tv
build-ui-tv:
	@echo "$(BLUE)Building TV application...$(NC)"
	@cd ../ui && ng build tv

.PHONY: build-ui-backoffice
build-ui-backoffice:
	@echo "$(BLUE)Building Backoffice application...$(NC)"
	@cd ../ui && ng build backoffice

.PHONY: serve-ui
serve-ui:
	@echo "$(BLUE)Starting all UI applications in development mode...$(NC)"
	@cd ../ui && ng serve kiosk --port 4201 &
	@cd ../ui && ng serve mobile --port 4204 &
	@cd ../ui && ng serve tv --port 4203 &
	@cd ../ui && ng serve backoffice --port 4200 &
	@echo "$(GREEN)All UI applications started!$(NC)"
	@echo "$(YELLOW)Kiosk: http://localhost:4201$(NC)"
	@echo "$(YELLOW)Mobile: http://localhost:4204$(NC)"
	@echo "$(YELLOW)TV: http://localhost:4203$(NC)"
	@echo "$(YELLOW)Backoffice: http://localhost:4200$(NC)"

.PHONY: serve-kiosk
serve-kiosk:
	@echo "$(BLUE)Starting Kiosk application...$(NC)"
	@cd ../ui && ng serve kiosk --port 4201

.PHONY: serve-mobile
serve-mobile:
	@echo "$(BLUE)Starting Mobile application...$(NC)"
	@cd ../ui && ng serve mobile --port 4204

.PHONY: serve-tv
serve-tv:
	@echo "$(BLUE)Starting TV application...$(NC)"
	@cd ../ui && ng serve tv --port 4203

.PHONY: serve-backoffice
serve-backoffice:
	@echo "$(BLUE)Starting Backoffice application...$(NC)"
	@cd ../ui && ng serve backoffice --port 4200

# =============================
# System Startup Commands
# =============================

.PHONY: start-system
start-system:
	@echo "$(BLUE)Starting the complete waiting room system...$(NC)"
	@cd .. && ./start-system.sh

.PHONY: start-docker
start-docker:
	@echo "$(BLUE)Starting system with Docker...$(NC)"
	@cd .. && ./start-docker.sh

.PHONY: start-docker-dev
start-docker-dev:
	@echo "$(BLUE)Starting system with Docker (development mode)...$(NC)"
	@cd .. && ./start-docker-dev.sh

.PHONY: stop-docker
stop-docker:
	@echo "$(YELLOW)Stopping Docker containers...$(NC)"
	@docker-compose -f ../$(DOCKER_COMPOSE) down
	@docker-compose -f ../$(DOCKER_COMPOSE_DEV) down

.PHONY: restart-docker
restart-docker: stop-docker start-docker

# =============================
# Docker Commands
# =============================

.PHONY: docker-build
docker-build:
	@echo "$(BLUE)Building Docker images...$(NC)"
	@docker-compose -f ../$(DOCKER_COMPOSE) build

.PHONY: docker-build-dev
docker-build-dev:
	@echo "$(BLUE)Building Docker images (development)...$(NC)"
	@docker-compose -f ../$(DOCKER_COMPOSE_DEV) build

.PHONY: docker-logs
docker-logs:
	@echo "$(BLUE)Showing Docker logs...$(NC)"
	@docker-compose -f ../$(DOCKER_COMPOSE) logs -f

.PHONY: docker-logs-dev
docker-logs-dev:
	@echo "$(BLUE)Showing Docker logs (development)...$(NC)"
	@docker-compose -f ../$(DOCKER_COMPOSE_DEV) logs -f

.PHONY: docker-clean
docker-clean:
	@echo "$(YELLOW)Cleaning Docker resources...$(NC)"
	@docker-compose -f ../$(DOCKER_COMPOSE) down -v --remove-orphans
	@docker-compose -f ../$(DOCKER_COMPOSE_DEV) down -v --remove-orphans
	@docker system prune -f

# =============================
# Database Commands
# =============================

.PHONY: mongo-start
mongo-start:
	@echo "$(BLUE)Starting MongoDB...$(NC)"
	@docker run -d --name waiting-room-mongo -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=admin mongo:latest

.PHONY: mongo-stop
mongo-stop:
	@echo "$(YELLOW)Stopping MongoDB...$(NC)"
	@docker stop waiting-room-mongo || true
	@docker rm waiting-room-mongo || true

.PHONY: mongo-shell
mongo-shell:
	@echo "$(BLUE)Connecting to MongoDB shell...$(NC)"
	@docker exec -it waiting-room-mongo mongosh -u admin -p admin

# =============================
# Testing Commands
# =============================

.PHONY: test-system
test-system:
	@echo "$(BLUE)Testing system components...$(NC)"
	@cd .. && ./test-system.sh

.PHONY: test-api
test-api:
	@echo "$(BLUE)Testing API endpoints...$(NC)"
	@curl -s http://localhost:8080/health || echo "$(RED)API not responding$(NC)"
	@curl -s http://localhost:8080/api/waiting-rooms/triage-1/queue | jq . || echo "$(RED)Queue endpoint not responding$(NC)"

.PHONY: test-ui
test-ui:
	@echo "$(BLUE)Testing UI applications...$(NC)"
	@curl -s http://localhost:4200 > /dev/null && echo "$(GREEN)Backoffice: OK$(NC)" || echo "$(RED)Backoffice: FAILED$(NC)"
	@curl -s http://localhost:4201 > /dev/null && echo "$(GREEN)Kiosk: OK$(NC)" || echo "$(RED)Kiosk: FAILED$(NC)"
	@curl -s http://localhost:4203 > /dev/null && echo "$(GREEN)TV: OK$(NC)" || echo "$(RED)TV: FAILED$(NC)"
	@curl -s http://localhost:4204 > /dev/null && echo "$(GREEN)Mobile: OK$(NC)" || echo "$(RED)Mobile: FAILED$(NC)"

# =============================
# Development Workflow
# =============================

.PHONY: dev-setup
dev-setup:
	@echo "$(BLUE)Setting up development environment...$(NC)"
	@make install-ui
	@make tidy
	@make gen
	@make build-ui
	@echo "$(GREEN)Development environment ready!$(NC)"

.PHONY: dev-start
dev-start:
	@echo "$(BLUE)Starting development environment...$(NC)"
	@make mongo-start
	@sleep 3
	@make run-dev &
	@sleep 2
	@make serve-ui

.PHONY: dev-stop
dev-stop:
	@echo "$(YELLOW)Stopping development environment...$(NC)"
	@pkill -f "ng serve" || true
	@pkill -f "go run" || true
	@make mongo-stop

# =============================
# Production Deployment
# =============================

.PHONY: prod-build
prod-build:
	@echo "$(BLUE)Building production images...$(NC)"
	@make build
	@make build-ui
	@make docker-build

.PHONY: prod-deploy
prod-deploy:
	@echo "$(BLUE)Deploying to production...$(NC)"
	@make prod-build
	@make start-docker

# =============================
# Utility Commands
# =============================

.PHONY: status
status:
	@echo "$(BLUE)System Status:$(NC)"
	@echo "$(YELLOW)API:$(NC) $$(curl -s http://localhost:8080/health > /dev/null && echo '$(GREEN)Running$(NC)' || echo '$(RED)Stopped$(NC)')"
	@echo "$(YELLOW)MongoDB:$(NC) $$(docker ps | grep mongo > /dev/null && echo '$(GREEN)Running$(NC)' || echo '$(RED)Stopped$(NC)')"
	@echo "$(YELLOW)UI Apps:$(NC)"
	@curl -s http://localhost:4200 > /dev/null && echo "  Backoffice: $(GREEN)Running$(NC)" || echo "  Backoffice: $(RED)Stopped$(NC)"
	@curl -s http://localhost:4201 > /dev/null && echo "  Kiosk: $(GREEN)Running$(NC)" || echo "  Kiosk: $(RED)Stopped$(NC)"
	@curl -s http://localhost:4203 > /dev/null && echo "  TV: $(GREEN)Running$(NC)" || echo "  TV: $(RED)Stopped$(NC)"
	@curl -s http://localhost:4204 > /dev/null && echo "  Mobile: $(GREEN)Running$(NC)" || echo "  Mobile: $(RED)Stopped$(NC)"

.PHONY: logs
logs:
	@echo "$(BLUE)Showing system logs...$(NC)"
	@echo "$(YELLOW)API Logs:$(NC)"
	@tail -f /tmp/waiting-room-api.log 2>/dev/null || echo "No API logs found"

.PHONY: clean
clean:
	@echo "$(YELLOW)Cleaning all build artifacts...$(NC)"
	@make clean-api
	@cd ../ui && rm -rf dist/ node_modules/.cache/
	@make docker-clean

.PHONY: help
help:
	@echo "$(BLUE)Waiting Room System - Available Commands:$(NC)"
	@echo ""
	@echo "$(YELLOW)Development:$(NC)"
	@echo "  dev-setup     - Set up development environment"
	@echo "  dev-start     - Start development environment"
	@echo "  dev-stop      - Stop development environment"
	@echo ""
	@echo "$(YELLOW)API Commands:$(NC)"
	@echo "  gen           - Generate API code from OpenAPI spec"
	@echo "  build         - Build API server"
	@echo "  run           - Run API server"
	@echo "  test          - Run API tests"
	@echo ""
	@echo "$(YELLOW)UI Commands:$(NC)"
	@echo "  install-ui    - Install UI dependencies"
	@echo "  build-ui      - Build all UI applications"
	@echo "  serve-ui      - Start all UI applications"
	@echo "  serve-kiosk   - Start Kiosk application"
	@echo "  serve-mobile  - Start Mobile application"
	@echo "  serve-tv      - Start TV application"
	@echo "  serve-backoffice - Start Backoffice application"
	@echo ""
	@echo "$(YELLOW)System Commands:$(NC)"
	@echo "  start-system  - Start complete system"
	@echo "  start-docker  - Start with Docker"
	@echo "  stop-docker   - Stop Docker containers"
	@echo "  status        - Show system status"
	@echo ""
	@echo "$(YELLOW)Testing:$(NC)"
	@echo "  test-system   - Test system components"
	@echo "  test-api      - Test API endpoints"
	@echo "  test-ui       - Test UI applications"
	@echo ""
	@echo "$(YELLOW)Utilities:$(NC)"
	@echo "  clean         - Clean all build artifacts"
	@echo "  logs          - Show system logs"
	@echo "  help          - Show this help message"

# Default target
.DEFAULT_GOAL := help


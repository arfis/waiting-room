# Waiting Room System - Root Makefile
# ====================================

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# =============================
# Quick Start Commands
# =============================

.PHONY: help
help:
	@echo "$(BLUE)Waiting Room System - Quick Start Commands:$(NC)"
	@echo ""
	@echo "$(YELLOW)Quick Start:$(NC)"
	@echo "  setup         - Complete system setup"
	@echo "  start         - Start the entire system"
	@echo "  stop          - Stop the entire system"
	@echo "  restart       - Restart the entire system"
	@echo ""
	@echo "$(YELLOW)Docker Commands:$(NC)"
	@echo "  docker-start  - Start with Docker"
	@echo "  docker-stop   - Stop Docker containers"
	@echo "  docker-build  - Build Docker images"
	@echo ""
	@echo "$(YELLOW)Development:$(NC)"
	@echo "  dev           - Start development environment"
	@echo "  dev-stop      - Stop development environment"
	@echo ""
	@echo "$(YELLOW)Testing:$(NC)"
	@echo "  test          - Test entire system"
	@echo "  status        - Show system status"
	@echo ""
	@echo "$(YELLOW)Utilities:$(NC)"
	@echo "  clean         - Clean all build artifacts"
	@echo "  logs          - Show system logs"
	@echo "  help          - Show this help message"

.PHONY: setup
setup:
	@echo "$(BLUE)Setting up the complete waiting room system...$(NC)"
	@echo "$(YELLOW)1. Installing UI dependencies...$(NC)"
	@cd ui && npm install
	@echo "$(YELLOW)2. Building UI applications...$(NC)"
	@cd ui && ng build kiosk
	@cd ui && ng build mobile
	@cd ui && ng build tv
	@cd ui && ng build backoffice
	@cd ui && ng build api-client
	@echo "$(YELLOW)3. Building API server...$(NC)"
	@cd api && make build
	@echo "$(GREEN)Setup complete!$(NC)"

.PHONY: start
start:
	@echo "$(BLUE)Starting the complete waiting room system...$(NC)"
	@./start-system.sh

.PHONY: stop
stop:
	@echo "$(YELLOW)Stopping the waiting room system...$(NC)"
	@pkill -f "ng serve" || true
	@pkill -f "go run" || true
	@pkill -f "node server.js" || true
	@docker stop waiting-room-mongo || true
	@docker rm waiting-room-mongo || true
	@echo "$(GREEN)System stopped!$(NC)"

.PHONY: restart
restart: stop start

# =============================
# Docker Commands
# =============================

.PHONY: docker-start
docker-start:
	@echo "$(BLUE)Starting system with Docker...$(NC)"
	@./start-docker.sh

.PHONY: docker-stop
docker-stop:
	@echo "$(YELLOW)Stopping Docker containers...$(NC)"
	@docker-compose down
	@docker-compose -f docker-compose.dev.yml down

.PHONY: docker-build
docker-build:
	@echo "$(BLUE)Building Docker images...$(NC)"
	@docker-compose build

.PHONY: docker-clean
docker-clean:
	@echo "$(YELLOW)Cleaning Docker resources...$(NC)"
	@docker-compose down -v --remove-orphans
	@docker-compose -f docker-compose.dev.yml down -v --remove-orphans
	@docker system prune -f

# =============================
# Development Commands
# =============================

.PHONY: dev
dev:
	@echo "$(BLUE)Starting development environment...$(NC)"
	@echo "$(YELLOW)Starting MongoDB...$(NC)"
	@docker run -d --name waiting-room-mongo -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=admin mongo:latest || true
	@sleep 3
	@echo "$(YELLOW)Starting API server...$(NC)"
	@cd api && make run-dev &
	@sleep 2
	@echo "$(YELLOW)Starting UI applications...$(NC)"
	@cd ui && ng serve kiosk --port 4201 &
	@cd ui && ng serve mobile --port 4204 &
	@cd ui && ng serve tv --port 4203 &
	@cd ui && ng serve backoffice --port 4200 &
	@echo "$(GREEN)Development environment started!$(NC)"
	@echo "$(YELLOW)Access points:$(NC)"
	@echo "  API: http://localhost:8080"
	@echo "  Kiosk: http://localhost:4201"
	@echo "  Mobile: http://localhost:4204"
	@echo "  TV: http://localhost:4203"
	@echo "  Backoffice: http://localhost:4200"

.PHONY: dev-stop
dev-stop:
	@echo "$(YELLOW)Stopping development environment...$(NC)"
	@pkill -f "ng serve" || true
	@pkill -f "go run" || true
	@docker stop waiting-room-mongo || true
	@docker rm waiting-room-mongo || true
	@echo "$(GREEN)Development environment stopped!$(NC)"

# =============================
# Testing Commands
# =============================

.PHONY: test
test:
	@echo "$(BLUE)Testing the complete system...$(NC)"
	@./test-system.sh

.PHONY: test-api
test-api:
	@echo "$(BLUE)Testing API endpoints...$(NC)"
	@curl -s http://localhost:8080/health > /dev/null && echo "$(GREEN)API: OK$(NC)" || echo "$(RED)API: FAILED$(NC)"
	@curl -s http://localhost:8080/api/waiting-rooms/triage-1/queue > /dev/null && echo "$(GREEN)Queue endpoint: OK$(NC)" || echo "$(RED)Queue endpoint: FAILED$(NC)"

.PHONY: test-ui
test-ui:
	@echo "$(BLUE)Testing UI applications...$(NC)"
	@curl -s http://localhost:4200 > /dev/null && echo "$(GREEN)Backoffice: OK$(NC)" || echo "$(RED)Backoffice: FAILED$(NC)"
	@curl -s http://localhost:4201 > /dev/null && echo "$(GREEN)Kiosk: OK$(NC)" || echo "$(RED)Kiosk: FAILED$(NC)"
	@curl -s http://localhost:4203 > /dev/null && echo "$(GREEN)TV: OK$(NC)" || echo "$(RED)TV: FAILED$(NC)"
	@curl -s http://localhost:4204 > /dev/null && echo "$(GREEN)Mobile: OK$(NC)" || echo "$(RED)Mobile: FAILED$(NC)"

# =============================
# Status and Monitoring
# =============================

.PHONY: status
status:
	@echo "$(BLUE)System Status:$(NC)"
	@echo "$(YELLOW)API:$(NC) $$(curl -s http://localhost:8080/health > /dev/null && echo '$(GREEN)Running$(NC)' || echo '$(RED)Stopped$(NC)')"
	@echo "$(YELLOW)MongoDB:$(NC) $$(docker ps | grep mongo > /dev/null && echo '$(GREEN)Running$(NC)' || echo '$(RED)Stopped$(NC)')"
	@echo "$(YELLOW)UI Applications:$(NC)"
	@curl -s http://localhost:4200 > /dev/null && echo "  Backoffice: $(GREEN)Running$(NC)" || echo "  Backoffice: $(RED)Stopped$(NC)"
	@curl -s http://localhost:4201 > /dev/null && echo "  Kiosk: $(GREEN)Running$(NC)" || echo "  Kiosk: $(RED)Stopped$(NC)"
	@curl -s http://localhost:4203 > /dev/null && echo "  TV: $(GREEN)Running$(NC)" || echo "  TV: $(RED)Stopped$(NC)"
	@curl -s http://localhost:4204 > /dev/null && echo "  Mobile: $(GREEN)Running$(NC)" || echo "  Mobile: $(RED)Stopped$(NC)"

.PHONY: logs
logs:
	@echo "$(BLUE)Showing system logs...$(NC)"
	@echo "$(YELLOW)API Logs:$(NC)"
	@tail -f /tmp/waiting-room-api.log 2>/dev/null || echo "No API logs found"

# =============================
# Cleanup Commands
# =============================

.PHONY: clean
clean:
	@echo "$(YELLOW)Cleaning all build artifacts...$(NC)"
	@cd api && make clean
	@cd ui && rm -rf dist/ node_modules/.cache/
	@make docker-clean
	@echo "$(GREEN)Cleanup complete!$(NC)"

.PHONY: clean-all
clean-all: clean
	@echo "$(YELLOW)Deep cleaning...$(NC)"
	@cd ui && rm -rf node_modules/
	@cd api && go clean -cache
	@docker system prune -a -f
	@echo "$(GREEN)Deep cleanup complete!$(NC)"

# =============================
# Production Commands
# =============================

.PHONY: prod-build
prod-build:
	@echo "$(BLUE)Building production images...$(NC)"
	@make setup
	@make docker-build

.PHONY: prod-deploy
prod-deploy:
	@echo "$(BLUE)Deploying to production...$(NC)"
	@make prod-build
	@make docker-start

# =============================
# Service Point Management
# =============================

.PHONY: manager-login
manager-login:
	@echo "$(BLUE)Manager login examples:$(NC)"
	@echo "$(YELLOW)Window 1 Manager:$(NC)"
	@echo "curl -X POST http://localhost:8080/api/managers/manager-1/login -H 'Content-Type: application/json' -d '{\"roomId\": \"triage-1\", \"servicePointId\": \"window-1\"}'"
	@echo "$(YELLOW)Window 2 Manager:$(NC)"
	@echo "curl -X POST http://localhost:8080/api/managers/manager-2/login -H 'Content-Type: application/json' -d '{\"roomId\": \"triage-1\", \"servicePointId\": \"window-2\"}'"

.PHONY: test-service-points
test-service-points:
	@echo "$(BLUE)Testing service point functionality...$(NC)"
	@echo "$(YELLOW)1. Creating test entry...$(NC)"
	@curl -X POST http://localhost:8080/api/waiting-rooms/triage-1/swipe -H "Content-Type: application/json" -d '{"idCardRaw": "test-card-123"}' | jq .
	@echo "$(YELLOW)2. Calling next for window-1...$(NC)"
	@curl -X POST http://localhost:8080/api/waiting-rooms/triage-1/service-points/window-1/call-next | jq .
	@echo "$(YELLOW)3. Marking as in room...$(NC)"
	@curl -X POST http://localhost:8080/api/waiting-rooms/triage-1/service-points/window-1/mark-in-room -H "Content-Type: application/json" -d '{"entryId": "test-id"}' | jq .


.PHONY: create-component
create-component:
	npx nx g @nx/angular:component $(name)
# Default target
.DEFAULT_GOAL := help

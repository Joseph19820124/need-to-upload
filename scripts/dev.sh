#!/bin/bash

set -e

# Development script for GitHub MCP HTTP server

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DEPLOY_DIR="$PROJECT_ROOT/deployments/local"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}GitHub MCP HTTP Server - Development Environment${NC}"

# Navigate to deployment directory
cd "$DEPLOY_DIR"

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}Creating .env file from template...${NC}"
    cp .env.example .env
    echo -e "${RED}Please edit .env file and add your GitHub token!${NC}"
    echo -e "${BLUE}Required: Set GITHUB_TOKEN=ghp_your_token_here${NC}"
    exit 1
fi

# Check if GitHub token is set
if ! grep -q "GITHUB_TOKEN=ghp_" .env; then
    echo -e "${RED}Error: GitHub token not set in .env file${NC}"
    echo -e "${BLUE}Please set GITHUB_TOKEN=ghp_your_token_here in .env${NC}"
    exit 1
fi

# Build and start the services
echo -e "${YELLOW}Building and starting development environment...${NC}"
docker-compose up -d --build

echo -e "${GREEN}Development environment started!${NC}"
echo -e "${BLUE}Server running at: http://localhost:8080${NC}"
echo -e "${BLUE}Health check: http://localhost:8080/api/v1/health${NC}"

echo -e "${YELLOW}Useful commands:${NC}"
echo "  View logs: docker-compose logs -f"
echo "  Stop services: docker-compose down"
echo "  Rebuild: docker-compose up -d --build"
echo ""
echo -e "${YELLOW}API Endpoints:${NC}"
echo "  POST /api/v1/connect - Establish MCP connection"
echo "  GET  /api/v1/events - SSE event stream"
echo "  POST /api/v1/rpc - MCP RPC calls"
echo "  POST /api/v1/disconnect - Close connection"
echo "  GET  /api/v1/health - Health check"
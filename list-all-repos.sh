#!/bin/bash

# List all repositories from GitHub MCP server
# Usage: ./list-all-repos.sh

SERVER_URL="http://localhost:8080"

echo "=== GitHub Repositories via MCP Server ==="

# Connect and get session
SESSION_ID=$(curl -s -X POST $SERVER_URL/api/v1/connect \
  -H "Content-Type: application/json" \
  -d '{"clientInfo": {"name": "repo-lister", "version": "1.0.0"}}' | \
  jq -r '.sessionId')

if [ "$SESSION_ID" = "null" ] || [ -z "$SESSION_ID" ]; then
    echo "Failed to connect to MCP server"
    exit 1
fi

echo "Connected with session: $SESSION_ID"
echo ""

# Get repositories
REPOS_RESPONSE=$(curl -s -X POST $SERVER_URL/api/v1/rpc \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "list_repositories",
      "arguments": {
        "type": "all",
        "sort": "updated"
      }
    },
    "id": 1
  }')

# Parse and display results
TOTAL=$(echo "$REPOS_RESPONSE" | jq -r '.result.content[0].text' | jq '. | length')
echo "Total repositories returned: $TOTAL"
echo ""

echo "All your repositories (sorted by last updated):"
echo "================================================"

# Display repositories with details
echo "$REPOS_RESPONSE" | jq -r '.result.content[0].text' | \
  jq -r '.[] | "\(.name) | \(.language // "N/A") | \(.description // "No description") | Updated: \(.updated_at)"' | \
  nl -w3 -s'. '

echo ""
echo "Repository names only:"
echo "====================="
echo "$REPOS_RESPONSE" | jq -r '.result.content[0].text' | jq -r '.[].name'

# Disconnect
curl -s -X POST $SERVER_URL/api/v1/disconnect \
  -H "X-Session-ID: $SESSION_ID" > /dev/null

echo ""
echo "Disconnected from MCP server"
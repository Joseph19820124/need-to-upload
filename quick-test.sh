#!/bin/bash

# Quick GitHub MCP HTTP Server Test
# Usage: ./quick-test.sh

SERVER_URL="http://localhost:8080"

echo "=== GitHub MCP HTTP Server Quick Test ==="
echo "Server: $SERVER_URL"
echo ""

# Test 1: Health Check
echo "1. Health Check:"
curl -s $SERVER_URL/api/v1/health | jq '.'
echo ""

# Test 2: Connect and get session
echo "2. Connect to MCP:"
CONNECT_RESPONSE=$(curl -s -X POST $SERVER_URL/api/v1/connect \
  -H "Content-Type: application/json" \
  -d '{"clientInfo": {"name": "quick-test", "version": "1.0.0"}}')

echo "$CONNECT_RESPONSE" | jq '.'
SESSION_ID=$(echo "$CONNECT_RESPONSE" | jq -r '.sessionId')
echo "Session ID: $SESSION_ID"
echo ""

# Test 3: List tools
echo "3. List Available Tools:"
curl -s -X POST $SERVER_URL/api/v1/rpc \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION_ID" \
  -d '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' | jq '.result.tools[].name'
echo ""

# Test 4: List repositories
echo "4. List Your Repositories:"
REPOS_RESPONSE=$(curl -s -X POST $SERVER_URL/api/v1/rpc \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION_ID" \
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "list_repositories", "arguments": {"type": "all", "sort": "updated"}}, "id": 2}')

TOTAL_REPOS=$(echo "$REPOS_RESPONSE" | jq -r '.result.content[0].text' | jq '. | length')
echo "Total repositories returned: $TOTAL_REPOS"
echo "First 10 repositories:"
echo "$REPOS_RESPONSE" | jq -r '.result.content[0].text' | jq '.[].name' | head -10
echo ""

# Test 5: Disconnect
echo "5. Disconnect:"
curl -s -X POST $SERVER_URL/api/v1/disconnect \
  -H "X-Session-ID: $SESSION_ID"
echo "✓ Disconnected"
echo ""

echo "=== Quick Test Complete ==="
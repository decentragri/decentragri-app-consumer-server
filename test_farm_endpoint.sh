#!/bin/bash

# Test script for the new farm list endpoint
echo "Testing Farm List Endpoint..."

DEV_TOKEN="dev_bypass_authorized"
SERVER_URL="http://127.0.0.1:9085"

echo ""
echo "1. Testing farm list with dev bypass token..."
response=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $DEV_TOKEN" "$SERVER_URL/api/farm/list")
echo "Response: $response"

echo ""
echo "2. Testing with X-Dev-Bypass-Token header..."
response2=$(curl -s -w "\n%{http_code}" -H "X-Dev-Bypass-Token: $DEV_TOKEN" "$SERVER_URL/api/farm/list")
echo "Response: $response2"

echo ""
echo "3. Testing without token (should fail)..."
response3=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/farm/list")
echo "Response: $response3"

echo ""
echo "Test complete!"

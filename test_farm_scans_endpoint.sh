#!/bin/bash

# Test script for the new farm scans endpoint
echo "Testing Farm Scans Endpoint..."

SERVER_URL="http://127.0.0.1:9085"
FARM_NAME="papaya_farm"

echo ""
echo "1. Testing farm scans with default pagination (page 1, limit 10)..."
response=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/farm/scans/$FARM_NAME")
echo "Response: $response"

echo ""
echo "2. Testing farm scans with custom pagination (page 1, limit 5)..."
response2=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/farm/scans/$FARM_NAME?page=1&limit=5")
echo "Response: $response2"

echo ""
echo "3. Testing farm scans with page 2..."
response3=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/farm/scans/$FARM_NAME?page=2&limit=5")
echo "Response: $response3"

echo ""
echo "4. Testing with another farm name..."
response4=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/farm/scans/test_farm")
echo "Response: $response4"

echo ""
echo "5. Testing JSON structure..."
curl -s "$SERVER_URL/api/farm/scans/$FARM_NAME?limit=2" | jq '.'

echo ""
echo "Test complete!"

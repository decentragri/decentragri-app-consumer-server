#!/bin/bash

# Quick test script for dev bypass token
echo "Testing Dev Bypass Token..."

DEV_TOKEN="dev_bypass_7k9m2x8p4q1w6e3r5t7y9u0i2o4p6a8s1d3f5g7h9j2k4l6n8m0q2w4e6r8t0y2u4i6o8p0"
SERVER_URL="http://127.0.0.1:9085"

echo ""
echo "1. Testing with header method..."
response1=$(curl -s -w "\n%{http_code}" -H "X-Dev-Bypass-Token: $DEV_TOKEN" "$SERVER_URL/api/marketplace/valid-farmplots")
echo "Response: $response1"

echo ""
echo "2. Testing with query parameter method..."
response2=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/marketplace/valid-farmplots?dev_bypass_token=$DEV_TOKEN")
echo "Response: $response2"

echo ""
echo "3. Testing without token (should fail)..."
response3=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/marketplace/valid-farmplots")
echo "Response: $response3"

echo ""
echo "Test complete!"
echo "Use this token in your Godot client: $DEV_TOKEN"

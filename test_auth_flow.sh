#!/bin/bash

echo "🔧 Testing Dev Bypass Authentication"

DEV_TOKEN="dev_bypass_7k9m2x8p4q1w6e3r5t7y9u0i2o4p6a8s1d3f5g7h9j2k4l6n8m0q2w4e6r8t0y2u4i6o8p0"
SERVER_URL="http://127.0.0.1:9085"

echo ""
echo "1️⃣ Testing dev bypass authentication..."
RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "X-Dev-Bypass-Token: $DEV_TOKEN" \
  -d "{}" \
  "$SERVER_URL/api/auth/dev-bypass")

echo "Auth Response: $RESPONSE"

# Extract access token from response
ACCESS_TOKEN=$(echo "$RESPONSE" | grep -o '"accessToken":"[^"]*"' | sed 's/"accessToken":"\(.*\)"/\1/')

if [ -z "$ACCESS_TOKEN" ]; then
    echo "❌ Failed to get access token from dev bypass"
    exit 1
fi

echo "🔑 Access Token: $ACCESS_TOKEN"

echo ""
echo "2️⃣ Testing portfolio API with JWT token only..."
curl -s -X GET \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  "$SERVER_URL/api/portfolio/summary" \
  -w "\nHTTP Status: %{http_code}\n"

echo ""
echo "3️⃣ Testing portfolio API with dev bypass header..."
curl -s -X GET \
  -H "Content-Type: application/json" \
  -H "X-Dev-Bypass-Token: $DEV_TOKEN" \
  "$SERVER_URL/api/portfolio/summary" \
  -w "\nHTTP Status: %{http_code}\n"

echo ""
echo "4️⃣ Testing portfolio API with both headers (like your Godot client)..."
curl -s -X GET \
  -H "Content-Type: application/json" \
  -H "X-Dev-Bypass-Token: $DEV_TOKEN" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  "$SERVER_URL/api/portfolio/summary" \
  -w "\nHTTP Status: %{http_code}\n"

echo ""
echo "✅ Test completed!"

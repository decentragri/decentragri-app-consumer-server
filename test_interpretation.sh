#!/bin/bash

# Test the farm scans endpoint after fixing interpretation parsing
echo "Testing plant scan interpretation parsing..."

# Start the server in background
./decentragri-app-cx-server &
SERVER_PID=$!

# Wait for server to start
sleep 5

# Make a request and capture the response
echo "Making request to farm scans endpoint..."
RESPONSE=$(curl -s "http://localhost:9085/api/farm/scans/strawberry?limit=1")

# Check if we got a response
if [ $? -eq 0 ]; then
    echo "Request successful!"
    echo "Response (first 1000 chars):"
    echo "$RESPONSE" | head -c 1000
    echo ""
    echo ""
    echo "Checking interpretation field:"
    echo "$RESPONSE" | jq '.plantScans[0].interpretation' 2>/dev/null || echo "Could not parse JSON or interpretation field not found"
else
    echo "Request failed!"
fi

# Kill the server
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo "Test completed."

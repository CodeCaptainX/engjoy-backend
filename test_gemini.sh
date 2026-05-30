#!/bin/bash

# Load API Key from .env file
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

KEY=$GEMINI_API_KEY

echo "Testing v1beta..."
curl -s -X POST "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=$KEY" \
     -H "Content-Type: application/json" \
     -d '{"contents": [{"parts":[{"text": "Say hello"}]}]}' | head -n 20

echo -e "\n\nTesting v1..."
curl -s -X POST "https://generativelanguage.googleapis.com/v1/models/gemini-1.5-flash:generateContent?key=$KEY" \
     -H "Content-Type: application/json" \
     -d '{"contents": [{"parts":[{"text": "Say hello"}]}]}' | head -n 20

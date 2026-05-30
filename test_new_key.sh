#!/bin/bash

# Load API Key from .env file
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

KEY=$GEMINI_API_KEY

echo "Testing new key with v1beta..."
curl -v -X POST "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=$KEY" \
     -H "Content-Type: application/json" \
     -d '{"contents": [{"parts":[{"text": "Say hello"}]}]}' 2>&1 | head -n 50

#!/bin/bash

# Load API Key from .env file
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

KEY=$GEMINI_API_KEY

echo "Listing models..."
curl -s -X GET "https://generativelanguage.googleapis.com/v1beta/models?key=$KEY" | head -n 50

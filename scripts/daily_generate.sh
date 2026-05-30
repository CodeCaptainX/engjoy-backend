#!/bin/bash

# Configuration
# Note: You need a valid Admin JWT token. 
# If you don't have one, you can generate it by logging in as an admin.
TOKEN="REPLACE_WITH_YOUR_ADMIN_JWT_TOKEN"
BASE_URL="http://localhost:8080/api/admin/generate-sentences"

# List of categories to pre-generate
CATEGORIES=(
    "daily-life"
    "travel"
    "airport"
    "restaurant"
    "hospital"
    "banking"
    "job-interview"
    "office"
    "shopping"
    "tech-support"
    "school"
    "sports"
    "phone-call"
    "emergency"
    "renting"
    "general"
    "deep-sea-exploration"
    "space-travel"
)

echo "Starting daily sentence generation at $(date)"

for CAT in "${CATEGORIES[@]}"; do
    echo "Generating for category: $CAT..."
    curl -s -X POST "$BASE_URL" \
         -H "Authorization: Bearer $TOKEN" \
         -H "Content-Type: application/json" \
         -d "{\"category\": \"$CAT\", \"count\": 10, \"focus\": \"useful expressions\"}"
    echo -e "\nDone with $CAT."
    # Small sleep to avoid hitting rate limits too fast
    sleep 2
done

echo "Daily generation complete at $(date)"

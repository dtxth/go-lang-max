#!/bin/bash

# Тест concurrent refresh requests
REFRESH_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjYyNTI1NTcsImlhdCI6MTc2NTY0Nzc1NywianRpIjoiZWI2MDA5MmUtNmE0Yi00MzAzLTk3MjQtMDA3YWI3NmE0ODFkIiwicGhvbmUiOiIrNzkwMDEyMzQ1NjciLCJyb2xlIjoib3BlcmF0b3IiLCJzdWIiOiIxIn0.iV21iFMb1MBV0jALewRnfvcixINf8A4VQhmoBoEBdnI"

echo "Testing concurrent refresh requests..."

# Запускаем 3 одновременных запроса
for i in {1..3}; do
    curl -X POST http://localhost:8080/refresh \
      -H "Content-Type: application/json" \
      -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}" \
      -w "\nRequest $i: HTTP %{http_code}\n" &
done

wait
echo "All requests completed"
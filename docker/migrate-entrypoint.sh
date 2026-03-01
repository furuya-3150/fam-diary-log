#!/bin/sh
set -e

echo "Running diary migrations..."
migrate -path setup/db/migrations/diary -database "$DIARY_DATABASE_URL" up

echo "Running diary-analyzer migrations..."
migrate -path setup/db/migrations/diary-analyzer -database "$DIARY_ANALYZER_DATABASE_URL" up

echo "Running user-context migrations..."
migrate -path setup/db/migrations/user-context -database "$USER_CONTEXT_DATABASE_URL" up

echo "All migrations completed."

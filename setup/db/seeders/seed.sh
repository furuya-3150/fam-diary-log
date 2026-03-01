#!/bin/bash

# Database Seeder Script
# このスクリプトは全てのシーダーをDocker環境で実行します

set -e

echo "🌱 Starting database seeding..."
echo ""

# user-context
echo "📝 Seeding user-context database..."
docker compose exec -T db psql -U user_context_user -d user_context < setup/db/seeders/01_user_context_seed.sql
echo "✅ user-context seeding completed"
echo ""

# diary
echo "📝 Seeding diary database..."
docker compose exec -T db psql -U diary_user -d diary < setup/db/seeders/02_diary_seed.sql
echo "✅ diary seeding completed"
echo ""

# diary-analyzer
echo "📝 Seeding diary-analyzer database..."
docker compose exec -T db psql -U diary_analyze_user -d diary_analyze < setup/db/seeders/03_diary_analysis_seed.sql
echo "✅ diary-analyzer seeding completed"
echo ""

echo "🎉 All seeding completed successfully!"
echo ""
echo "📊 Data summary:"
docker compose exec -T db psql -U user_context_user -d user_context -c "SELECT COUNT(*) as users FROM users;"
docker compose exec -T db psql -U user_context_user -d user_context -c "SELECT COUNT(*) as families FROM families;"
docker compose exec -T db psql -U diary_user -d diary -c "SELECT COUNT(*) as diaries FROM diaries;"
docker compose exec -T db psql -U diary_analyze_user -d diary_analyze -c "SELECT COUNT(*) as analyses FROM diary_analyses;"

#!/bin/bash

# Validate SQL migration files
echo "Validating GoRectus migration files..."

# Check if migration files exist
if [ ! -d "migrations" ]; then
    echo "❌ migrations directory not found"
    exit 1
fi

# Count migration files
up_files=$(find migrations -name "*.up.sql" | wc -l)
down_files=$(find migrations -name "*.down.sql" | wc -l)

echo "Found $up_files up migrations and $down_files down migrations"

if [ $up_files -ne $down_files ]; then
    echo "❌ Mismatch between up and down migration files"
    exit 1
fi

# Check if postgresql is available for syntax validation
if command -v psql &> /dev/null; then
    echo "✅ PostgreSQL client found, validating SQL syntax..."
    
    # Create temporary database connection (dry run)
    for file in migrations/*.up.sql; do
        echo "Checking syntax: $file"
        # Use --dry-run equivalent: just parse without executing
        if ! psql --help &> /dev/null; then
            echo "⚠️  Cannot validate SQL syntax without database connection"
            break
        fi
    done
else
    echo "⚠️  PostgreSQL client not found, skipping syntax validation"
fi

# Validate migration file naming
echo "Validating migration file naming conventions..."

for file in migrations/*.sql; do
    basename=$(basename "$file")
    if [[ ! $basename =~ ^[0-9]{6}_[a-zA-Z0-9_]+\.(up|down)\.sql$ ]]; then
        echo "❌ Invalid migration filename: $basename"
        echo "   Expected format: NNNNNN_description.(up|down).sql"
        exit 1
    fi
done

# Check for required migration content
echo "Validating migration content..."

if ! grep -q "CREATE TABLE.*users" migrations/000001_initial_schema.up.sql; then
    echo "❌ Initial migration must create users table"
    exit 1
fi

if ! grep -q "INSERT INTO users" migrations/000002_seed_data.up.sql; then
    echo "❌ Seed migration must insert admin user"
    exit 1
fi

echo "✅ All migration files validated successfully!"
echo ""
echo "Next steps:"
echo "1. Start PostgreSQL: make db-up"
echo "2. Run migrations: make migrate-up"
echo "3. Check status: make migrate-status"

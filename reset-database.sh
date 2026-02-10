#!/bin/bash
# Script to reset the urlshortener_db database

# Database connection details (from docker-compose.oracle.yml)
DB_HOST="postgresql-central"
DB_USER="urlshortener_user"
DB_PASSWORD="${DB_PASSWORD:-URLShortenerSecurePassword123!}"
DB_NAME="urlshortener_db"

echo "Connecting to PostgreSQL to drop and recreate database..."

# Execute commands in the postgresql-central container
docker exec -i postgresql-central psql -U postgres <<EOF
-- Terminate all connections to the database
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = '$DB_NAME'
  AND pid <> pg_backend_pid();

-- Drop the database
DROP DATABASE IF EXISTS $DB_NAME;

-- Recreate the database
CREATE DATABASE $DB_NAME;

-- Grant privileges to the user
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;

\c $DB_NAME

-- Grant schema privileges (PostgreSQL 15+ requires this)
GRANT ALL ON SCHEMA public TO $DB_USER;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DB_USER;

EOF

echo "Database reset complete! Now restart the app container to initialize tables."
echo "docker restart urlshortener-app"

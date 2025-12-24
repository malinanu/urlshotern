#!/bin/bash
# Development script to run Go backend server
# Usage: ./scripts/dev-backend.sh

set -e

echo "🚀 Starting Go backend server..."
go run cmd/server/main.go

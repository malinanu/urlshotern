#!/bin/bash
# Development script to run Next.js frontend dev server
# Usage: ./scripts/dev-frontend.sh

set -e

echo "🎨 Starting Next.js development server..."
cd frontend && npm run dev

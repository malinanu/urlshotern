# GitHub Issues Creation Guide

This guide helps you create all 34 identified issues for the URL Shortener project.

## Prerequisites

### Install GitHub CLI
```bash
# Windows (using winget)
winget install --id GitHub.cli

# Or download from: https://cli.github.com/

# Verify installation
gh --version

# Authenticate
gh auth login
```

## Step 1: Create Labels

Run this script to create all necessary labels:

```bash
# Severity labels
gh label create "critical" --color "d73a4a" --description "Critical priority - fix immediately"
gh label create "high" --color "ff9500" --description "High priority - fix soon"
gh label create "medium" --color "fbca04" --description "Medium priority - schedule for sprint"
gh label create "low" --color "0e8a16" --description "Low priority - backlog"

# Component labels
gh label create "backend" --color "0052cc" --description "Backend Go code"
gh label create "frontend" --color "5319e7" --description "Frontend React/Next.js"
gh label create "infra" --color "1d76db" --description "Infrastructure & deployment"
gh label create "docs" --color "d4c5f9" --description "Documentation"

# Type labels
gh label create "bug" --color "d73a4a" --description "Something isn't working"
gh label create "enhancement" --color "a2eeef" --description "New feature or request"
gh label create "security" --color "ff0000" --description "Security vulnerability"
gh label create "technical-debt" --color "f9d0c4" --description "Code quality improvement"
gh label create "a11y" --color "7057ff" --description "Accessibility"

# Effort labels
gh label create "quick-win" --color "bfd4f2" --description "Can be done quickly (< 4 hours)"
gh label create "medium-effort" --color "5319e7" --description "Medium effort (1-2 days)"
gh label create "large-effort" --color "0e8a16" --description "Large effort (> 2 days)"
```

## Step 2: Create Issues

### Critical Issues (4)

```bash
# Issue #1
gh issue create \
  --title "[CRITICAL] Remove Hardcoded User ID Fallback" \
  --body-file .github/issues/001-critical-hardcoded-user-id.md \
  --label "critical,security,backend,bug,quick-win"

# Issue #2
gh issue create \
  --title "[CRITICAL] Implement Usage Limit Enforcement" \
  --body-file .github/issues/002-critical-usage-limit-enforcement.md \
  --label "critical,backend,enhancement,medium-effort"

# Issue #3
gh issue create \
  --title "[CRITICAL] Enable PostgreSQL SSL/TLS in Production" \
  --body-file .github/issues/024-critical-postgres-ssl.md \
  --label "critical,security,infra,quick-win"

# Issue #4
gh issue create \
  --title "[CRITICAL] Remove Hardcoded Database Credentials" \
  --body "## Description
Production and development docker-compose files contain hardcoded database passwords.

## Affected Files
- docker/docker-compose.yml
- docker/docker-compose.local.yml

## Proposed Solution
1. Use environment variables
2. Create .env.example template
3. Use Kubernetes secrets in production
4. Update documentation" \
  --label "critical,security,infra,quick-win"
```

### High Priority Issues (10)

```bash
# Issue #5
gh issue create \
  --title "[HIGH] Implement Webhook Retry Mechanism" \
  --body-file .github/issues/003-high-webhook-retry.md \
  --label "high,backend,enhancement,medium-effort"

# Issue #6
gh issue create \
  --title "[HIGH] Replace Mock Invoice Data with Real Polar Integration" \
  --body "## Description
Billing handler returns hardcoded mock invoice data instead of fetching from Polar API.

## Affected Files
- internal/handlers/billing.go (line 314-327)

## Proposed Solution
1. Implement GetInvoices() method in Polar service
2. Update billing handler to call real API
3. Add invoice list to frontend billing page" \
  --label "high,backend,enhancement,medium-effort"

# Issue #7
gh issue create \
  --title "[HIGH] Fix SQL Query Parameter Handling in Analytics" \
  --body "## Description
Advanced analytics service uses fmt.Sprintf to inject days parameter into SQL queries instead of parameterized queries.

## Affected Files
- internal/services/advanced_analytics.go (lines 334, 504, 554, 589, 628, 658, 702, 723)

## Proposed Solution
Replace fmt.Sprintf with proper query parameters:
\`\`\`go
// Instead of:
rows, err := a.storage.Query(fmt.Sprintf(query, days), shortCode)

// Use:
rows, err := a.storage.Query(query, days, shortCode)
\`\`\`" \
  --label "high,security,backend,bug,quick-win"

# Issue #8
gh issue create \
  --title "[HIGH] Fix Docker Port Mismatch" \
  --body "## Description
Dockerfile exposes port 8081 but docker-compose.yml uses port 8080.

## Affected Files
- docker/Dockerfile (line 44)
- docker/docker-compose.yml (line 9)

## Proposed Solution
Standardize on port 8080 everywhere." \
  --label "high,infra,bug,quick-win"

# Issue #9
gh issue create \
  --title "[HIGH] Implement Database Migration Rollback Scripts" \
  --body "## Description
Only forward migrations exist. No down/rollback migrations available.

## Affected Files
- configs/migrations/ directory

## Proposed Solution
1. Create down migrations for all existing migrations
2. Add rollback targets to Makefile
3. Document rollback procedures" \
  --label "high,infra,enhancement,medium-effort"

# Issue #10
gh issue create \
  --title "[HIGH] Configure Database Connection Pooling" \
  --body "## Description
No connection pooling configured (PgBouncer or similar).

## Affected Files
- configs/config.go (line 23)

## Proposed Solution
1. Add PgBouncer to infrastructure
2. Configure pool size limits
3. Add monitoring for connection pool" \
  --label "high,infra,enhancement,medium-effort"

# Issue #11
gh issue create \
  --title "[HIGH] Remove Console Logging from Production Code" \
  --body "## Description
73+ console.log/console.error statements in production frontend code.

## Affected Files
- frontend/src/contexts/AuthContext.tsx (30+ console statements)
- frontend/src/utils/performanceMonitoring.ts (15+ statements)
- 20+ other files

## Proposed Solution
1. Replace console.log with proper logging service
2. Use environment-based debug logging
3. Add eslint rule to prevent console statements" \
  --label "high,frontend,technical-debt,quick-win"

# Issue #12
gh issue create \
  --title "[HIGH] Replace Hardcoded API URLs with Environment Variables" \
  --body "## Description
20+ files use hardcoded 'http://localhost:8080' as API fallback.

## Affected Files
- Multiple files across frontend/src

## Proposed Solution
1. Create centralized API config
2. Remove localhost fallback
3. Require NEXT_PUBLIC_API_URL environment variable
4. Update deployment documentation" \
  --label "high,frontend,bug,quick-win"
```

### Continue with remaining 24 issues...

See `COMPLETE_ISSUES_LIST.md` for full list of all 34 issues.

## Step 3: Verify Issues Created

```bash
# List all issues
gh issue list --limit 100

# Check by label
gh issue list --label "critical"
gh issue list --label "backend"
```

## Step 4: Create Project Board (Optional)

```bash
# Create project
gh project create --title "URL Shortener Improvements" --body "Tracking all issues from codebase audit"

# Add issues to project (done via web UI or gh project commands)
```

## Quick Batch Creation Script

Save this as `create-all-issues.sh`:

```bash
#!/bin/bash

# Create labels first (run once)
./create-labels.sh

# Create all 34 issues
for issue_file in .github/issues/*.md; do
    # Extract title from first line of file
    title=$(head -n 1 "$issue_file" | sed 's/^# //')

    # Extract labels from second line
    labels=$(sed -n '3p' "$issue_file" | sed 's/**Labels:** `//; s/`//')

    echo "Creating issue: $title"

    gh issue create \
      --title "$title" \
      --body-file "$issue_file" \
      --label "$labels"

    sleep 1  # Rate limiting
done

echo "All issues created!"
```

## Manual Creation (if CLI not available)

1. Go to https://github.com/YOUR_USERNAME/URLshorter/issues/new
2. Copy title from issue file first line
3. Copy body from issue file
4. Add labels manually
5. Repeat for all 34 issues

## Summary

- **Total Issues**: 34
- **Critical**: 4 issues
- **High**: 10 issues
- **Medium**: 14 issues
- **Low**: 6 issues

All issue templates are in `.github/issues/` directory.

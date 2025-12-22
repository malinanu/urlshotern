@echo off
echo Creating Critical Priority GitHub Issues...
echo.

REM Issue #1
echo Creating Issue #1: Remove Hardcoded User ID Fallback...
gh issue create ^
  --title "[CRITICAL] Remove Hardcoded User ID Fallback" ^
  --body-file .github/issues/001-critical-hardcoded-user-id.md ^
  --label "critical,security,backend,bug,quick-win"

REM Issue #2
echo Creating Issue #2: Implement Usage Limit Enforcement...
gh issue create ^
  --title "[CRITICAL] Implement Usage Limit Enforcement" ^
  --body-file .github/issues/002-critical-usage-limit-enforcement.md ^
  --label "critical,backend,enhancement,medium-effort"

REM Issue #3
echo Creating Issue #3: Enable PostgreSQL SSL/TLS...
gh issue create ^
  --title "[CRITICAL] Enable PostgreSQL SSL/TLS in Production" ^
  --body-file .github/issues/024-critical-postgres-ssl.md ^
  --label "critical,security,infra,quick-win"

REM Issue #4
echo Creating Issue #4: Remove Hardcoded Database Credentials...
gh issue create ^
  --title "[CRITICAL] Remove Hardcoded Database Credentials" ^
  --body "## Description%nProduction and development docker-compose files contain hardcoded database passwords.%n%n## Affected Files%n- docker/docker-compose.yml%n- docker/docker-compose.local.yml%n%n## Proposed Solution%n1. Use environment variables%n2. Create .env.example template%n3. Use Kubernetes secrets in production%n4. Update documentation" ^
  --label "critical,security,infra,quick-win"

echo.
echo All critical issues created successfully!
echo.
echo To view issues:
echo   gh issue list --label critical
echo.
pause

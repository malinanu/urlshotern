@echo off
echo Creating GitHub labels for URL Shortener project...
echo.

REM Severity labels
echo Creating severity labels...
gh label create "critical" --color "d73a4a" --description "Critical priority - fix immediately" 2>nul
gh label create "high" --color "ff9500" --description "High priority - fix soon" 2>nul
gh label create "medium" --color "fbca04" --description "Medium priority - schedule for sprint" 2>nul
gh label create "low" --color "0e8a16" --description "Low priority - backlog" 2>nul

REM Component labels
echo Creating component labels...
gh label create "backend" --color "0052cc" --description "Backend Go code" 2>nul
gh label create "frontend" --color "5319e7" --description "Frontend React/Next.js" 2>nul
gh label create "infra" --color "1d76db" --description "Infrastructure & deployment" 2>nul
gh label create "docs" --color "d4c5f9" --description "Documentation" 2>nul

REM Type labels
echo Creating type labels...
gh label create "bug" --color "d73a4a" --description "Something isn't working" 2>nul
gh label create "enhancement" --color "a2eeef" --description "New feature or request" 2>nul
gh label create "security" --color "ff0000" --description "Security vulnerability" 2>nul
gh label create "technical-debt" --color "f9d0c4" --description "Code quality improvement" 2>nul
gh label create "a11y" --color "7057ff" --description "Accessibility" 2>nul

REM Effort labels
echo Creating effort labels...
gh label create "quick-win" --color "bfd4f2" --description "Can be done quickly (< 4 hours)" 2>nul
gh label create "medium-effort" --color "5319e7" --description "Medium effort (1-2 days)" 2>nul
gh label create "large-effort" --color "0e8a16" --description "Large effort (> 2 days)" 2>nul

echo.
echo All labels created successfully!
echo.
echo To view labels:
echo   gh label list
echo.
pause

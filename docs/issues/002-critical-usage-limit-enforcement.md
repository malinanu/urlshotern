# [CRITICAL] Implement Usage Limit Enforcement

**Labels:** `critical`, `backend`, `enhancement`, `medium`

## Description
Usage limits are defined in the billing service but not enforced when creating URLs or processing clicks, allowing users to exceed their plan limits without restriction.

## Severity: CRITICAL
This is critical because:
- Free users can create unlimited URLs (should be limited by plan)
- Click tracking limits not enforced
- Revenue loss from users not upgrading
- Database/resource abuse potential
- Core business model not enforced

## Affected Files
- [ ] `internal/handlers/handlers.go` (ShortenURL method, lines 75-123)
- [ ] `internal/services/shortener.go` (RecordClick method, lines 237-293)
- [ ] `internal/services/billing.go` (CanCreateURL, CanProcessClick methods exist but unused)

## Current Behavior
```go
// ShortenURL handler
func (h *Handlers) ShortenURL(c *gin.Context) {
    // ... validation ...

    // NO CHECK FOR URL LIMITS
    shortCode, err := h.shortenerService.CreateShortURL(...)
    // ...
}
```

The billing service has `CanCreateURL()` and `CanProcessClick()` methods, but they are never called before allowing the operations.

## Expected Behavior
- Check user's plan limits before creating URLs
- Check click processing limits
- Return clear error message when limit exceeded
- Encourage upgrade to higher plan
- Properly enforce business model

## Proposed Solution

### 1. Modify ShortenURL Handler
```go
func (h *Handlers) ShortenURL(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        c.JSON(401, gin.H{"error": "Authentication required"})
        return
    }

    // CHECK USAGE LIMITS
    canCreate, err := h.billingService.CanCreateURL(c.Request.Context(), userID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to check usage limits"})
        return
    }

    if !canCreate {
        subscription, _ := h.billingService.GetSubscription(c.Request.Context(), userID)
        c.JSON(403, gin.H{
            "error": "URL creation limit reached for your plan",
            "current_plan": subscription.PlanType,
            "upgrade_url": "/billing",
        })
        return
    }

    // Proceed with URL creation
    // ...
}
```

### 2. Modify RecordClick Method
```go
func (s *ShortenerService) RecordClick(...) error {
    // CHECK CLICK LIMITS
    canProcess, err := s.billingService.CanProcessClick(ctx, userID)
    if err != nil {
        return fmt.Errorf("failed to check click limits: %w", err)
    }

    if !canProcess {
        // Log but don't fail the redirect
        log.Warn("User exceeded click processing limit", "user_id", userID)
        return nil // Don't record, but allow redirect
    }

    // Proceed with click recording
    // ...
}
```

### 3. Add Usage Stats to Response
Include current usage in API responses:
```go
{
    "short_code": "abc123",
    "usage": {
        "urls_created": 45,
        "url_limit": 50,
        "clicks_this_month": 890,
        "click_limit": 1000
    }
}
```

## Acceptance Criteria
- [ ] URL creation checks limits before allowing creation
- [ ] 403 error returned when limit exceeded
- [ ] Click processing respects limits
- [ ] Clear error messages guide users to upgrade
- [ ] Usage stats included in responses
- [ ] Existing functionality preserved
- [ ] Unit tests for limit enforcement
- [ ] Integration tests for limit scenarios
- [ ] Documentation updated

## Implementation Steps
1. Add billing service dependency to handlers
2. Call `CanCreateURL()` in ShortenURL handler
3. Call `CanProcessClick()` in RecordClick service
4. Add error handling and user-friendly messages
5. Add current usage to API responses
6. Write comprehensive tests
7. Update API documentation

## Business Impact
- **Revenue Protection**: Forces users to upgrade when hitting limits
- **Resource Management**: Prevents database bloat from unlimited URLs
- **Fair Usage**: Ensures free tier users don't abuse the system

## Related Issues
- #1 (Remove Hardcoded User ID) - must be fixed first for proper user identification
- #4 (Mock Invoice Data) - related to billing integration

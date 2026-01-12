# Admin Authentication Separation

## Overview
The admin panel now has a completely separate authentication system from the regular user authentication. This ensures proper security isolation and prevents regular users from accessing admin functionality.

## Changes Made

### 1. Frontend Changes

#### New AdminAuthContext (`frontend/src/contexts/AdminAuthContext.tsx`)
- Created a dedicated authentication context specifically for admin users
- Completely separate from the normal `AuthContext` used by regular users
- Stores admin tokens in separate localStorage keys:
  - `admin_access_token`
  - `admin_refresh_token`
  - `admin_user`
- Validates admin sessions against the backend
- Handles admin login, logout, and token refresh

#### Updated Admin Layout (`frontend/src/app/admin/layout.tsx`)
- Wraps all admin routes with `AdminAuthProvider`
- Checks authentication using `AdminAuthContext` (not regular `AuthContext`)
- Redirects unauthenticated users to `/admin/login`
- Shows loading state while verifying admin access

#### New Admin Components
- `AdminHeader.tsx` - Dedicated header for admin panel with logout button
- `AdminLayout.tsx` - Layout wrapper specifically for admin pages
- Updated all admin pages to use `AdminLayout` instead of regular `Layout`

#### Updated Admin Login (`frontend/src/app/admin/login/page.tsx`)
- Uses `AdminAuthContext` for authentication
- Redirects authenticated admins to `/admin` dashboard
- Displays admin-specific error messages

### 2. Backend Changes

#### Backend Security (Already in Place)
The backend was already properly configured with strong admin security:

**Admin Routes Protection** (`internal/routes/admin_routes.go`):
```go
admin := router.Group("/api/v1/admin")
admin.Use(authMiddleware.RequireAuth())           // Must be authenticated
admin.Use(adminMiddleware.RequireAdminAccountType()) // Must have admin account type
```

**Admin Middleware** (`internal/middleware/admin.go`):
- `RequireAdminAccountType()` - Verifies user has admin account type
- Checks for account types: `super_admin`, `system_admin`, `support_admin`, `billing_admin`, `content_admin`
- Also checks legacy `is_admin` flag
- Returns 403 Forbidden for non-admin users

**Admin Login Handler** (`internal/handlers/admin/admin_auth_handlers.go`):
- Validates admin credentials
- Checks if user has admin privileges before generating tokens
- Logs all admin login attempts for audit

### 3. Authentication Flow

#### Regular User Authentication
1. User logs in at `/login`
2. Tokens stored in `access_token`, `refresh_token`, `user_data`
3. Regular `AuthContext` manages authentication state
4. Can access: `/dashboard`, `/analytics`, `/settings`, etc.

#### Admin Authentication
1. Admin logs in at `/admin/login`
2. Tokens stored in `admin_access_token`, `admin_refresh_token`, `admin_user`
3. `AdminAuthContext` manages admin authentication state
4. Can access: `/admin`, `/admin/users`, `/admin/analytics`, etc.

### 4. Security Benefits

✅ **Complete Isolation**: Regular users cannot access admin routes even if they manipulate frontend code

✅ **Separate Token Storage**: Admin and user tokens are stored separately, preventing token confusion

✅ **Backend Validation**: Every admin API call validates:
   - Valid JWT token
   - User has admin account type
   - User account is active

✅ **Audit Logging**: All admin actions are logged for security auditing

✅ **Session Management**: Admin sessions have shorter timeouts and stricter controls

## Testing the Admin Panel

### 1. Create an Admin User
```bash
# Windows
.\scripts\create-admin.ps1 -Name "Admin" -Email "admin@example.com" -Password "secure123" -Type "super_admin"

# Linux/macOS
./scripts/create-admin.sh -n "Admin" -e "admin@example.com" -p "secure123" -t "super_admin"
```

### 2. Test Admin Login
1. Navigate to `http://localhost:3000/admin/login`
2. Enter admin credentials
3. Should redirect to admin dashboard
4. Verify admin header shows admin name and logout button

### 3. Test Regular User Cannot Access Admin
1. Log in as regular user at `/login`
2. Try to navigate to `/admin`
3. Should be redirected to `/admin/login`
4. Regular user token should not work for admin API calls

### 4. Test Logout
1. Click logout button in admin header
2. Should be redirected to `/admin/login`
3. Admin tokens should be cleared from localStorage
4. Attempting to access `/admin` should require login

## Admin Account Types

- **super_admin** - Full system access, can manage all resources
- **system_admin** - System management, can manage users and settings
- **support_admin** - User support, can view user data and assist
- **billing_admin** - Billing management, can manage subscriptions
- **content_admin** - Content management, can manage CMS content

## API Endpoints

### Public Admin Endpoints (No Auth)
- `POST /api/v1/admin/auth/login` - Admin login
- `POST /api/v1/admin/auth/refresh` - Token refresh

### Protected Admin Endpoints (Require Admin Auth)
- `POST /api/v1/admin/auth/logout` - Admin logout
- `GET /api/v1/admin` - Dashboard overview
- `GET /api/v1/admin/users` - User management
- `GET /api/v1/admin/analytics` - Admin analytics
- And many more (see `internal/routes/admin_routes.go`)

## Security Considerations

1. **Never skip authentication checks** - All admin routes must have both `RequireAuth()` and `RequireAdminAccountType()` middleware

2. **Validate on backend** - Frontend checks are for UX only; backend always validates admin status

3. **Audit logging** - All admin actions are logged with IP, user agent, and timestamp

4. **Token refresh** - Admin tokens expire and must be refreshed; use separate refresh token

5. **Rate limiting** - Admin endpoints have stricter rate limits for sensitive operations

## Troubleshooting

### Issue: Regular user can access admin panel
- **Cause**: Frontend check only
- **Solution**: Backend properly rejects non-admin users with 403 Forbidden

### Issue: Admin login shows "Invalid credentials"
- **Check**: User has admin account type in database
- **Check**: User account is active (`is_active = true`)
- **Check**: Backend is running and accessible

### Issue: Admin session expires quickly
- **Expected**: Admin sessions have 30-minute timeout for security
- **Solution**: This is intentional; admin must re-authenticate frequently

### Issue: Token refresh fails
- **Check**: Refresh token is stored in `admin_refresh_token`
- **Check**: Refresh token hasn't expired
- **Check**: Backend refresh endpoint is accessible

## Next Steps

The admin panel foundation is complete and secure. Future enhancements could include:

1. ✅ Two-factor authentication for admin accounts
2. ✅ IP whitelisting for admin access
3. ✅ Admin activity dashboard
4. ✅ Real-time admin notifications
5. ✅ Advanced RBAC with granular permissions

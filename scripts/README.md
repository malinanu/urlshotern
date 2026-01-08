# Admin User Management Scripts

This directory contains scripts to create and manage admin users for the URL Shortener application.

## Creating an Admin User

### Windows (PowerShell)

```powershell
# Interactive mode
.\scripts\create-admin.ps1

# Command line mode
.\scripts\create-admin.ps1 -Name "Admin User" -Email "admin@example.com" -Password "securepass123" -Type "super_admin"
```

### Linux/macOS (Bash)

```bash
# Make script executable (first time only)
chmod +x scripts/create-admin.sh

# Interactive mode
./scripts/create-admin.sh

# Command line mode
./scripts/create-admin.sh -n "Admin User" -e "admin@example.com" -p "securepass123" -t "super_admin"
```

### Direct Go Script

```bash
go run scripts/create-admin.go \
  -name="Admin User" \
  -email="admin@example.com" \
  -password="securepass123" \
  -type="super_admin"
```

## Account Types

The system supports the following admin account types:

| Type | Description | Access Level |
|------|-------------|--------------|
| `super_admin` | Full system access | All features and settings |
| `system_admin` | System management | System configuration and monitoring |
| `support_admin` | User support | User management and support tools |
| `billing_admin` | Billing management | Billing and subscription management |
| `content_admin` | Content management | CMS and content editing |

## Environment Variables

The scripts read database connection details from environment variables. If not set, they use the following defaults:

- `DB_HOST` - Database host (default: `localhost`)
- `DB_PORT` - Database port (default: `5432`)
- `DB_USER` - Database user (default: `postgres`)
- `DB_PASSWORD` - Database password (default: `postgres`)
- `DB_NAME` - Database name (default: `urlshortener`)

## Examples

### Create a Super Admin

```powershell
# Windows
.\scripts\create-admin.ps1 -Name "John Doe" -Email "john@company.com" -Password "SuperSecure123!"

# Linux/macOS
./scripts/create-admin.sh -n "John Doe" -e "john@company.com" -p "SuperSecure123!"
```

### Create a Support Admin

```powershell
# Windows
.\scripts\create-admin.ps1 -Name "Support Team" -Email "support@company.com" -Password "Support123!" -Type "support_admin"

# Linux/macOS
./scripts/create-admin.sh -n "Support Team" -e "support@company.com" -p "Support123!" -t "support_admin"
```

### Update Existing User to Admin

If you want to update an existing user to admin status, the script will automatically update the user if the email already exists:

```powershell
.\scripts\create-admin.ps1 -Name "Existing User" -Email "existing@example.com" -Password "NewPassword123" -Type "super_admin"
```

## Security Notes

- **Password Requirements**: Use strong passwords with at least 8 characters
- **Email Validation**: Ensure the email is valid and accessible
- **Account Types**: Choose the minimal account type needed for the user's role
- **Password Storage**: Passwords are automatically hashed using bcrypt before storage
- **Audit Logging**: All admin authentication attempts are logged

## Login

After creating an admin user, you can login at:
- **Development**: `http://localhost:3000/admin/login`
- **Production**: `https://yourdomain.com/admin/login`

## Troubleshooting

### Database Connection Failed

Check that:
1. PostgreSQL is running
2. Database credentials are correct
3. Database name exists
4. Network access is allowed

### Email Already Exists

If the email already exists, the script will update the existing user to admin status with the new password.

### Permission Denied

Make sure the database user has INSERT and UPDATE permissions on the `users` table.

## Other Scripts

- `setup-local.ps1` - Local development environment setup (Windows)
- `migrate.ps1` / `migrate.sh` - Database migration scripts

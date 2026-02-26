# Security Policy

## Credential Management

### Critical Security Requirements

**NEVER commit the following to version control:**
- Database passwords
- API keys and secrets
- JWT signing keys
- OAuth client secrets
- SSL/TLS private keys
- Production environment files (`.env`, `.env.prod`)

### Environment Variables

All sensitive credentials MUST be stored in environment variables or secure secret management systems.

#### Required Environment Variables

**Database Credentials:**
```bash
DB_PASSWORD=<strong-unique-password>
```

**API Keys (Production):**
```bash
POLAR_ACCESS_TOKEN=<your-polar-access-token>
POLAR_WEBHOOK_SECRET=<your-webhook-secret>
JWT_SECRET=<minimum-32-character-random-string>
```

#### Generating Secure Passwords

**Database Password:**
```bash
# Generate a strong random password
openssl rand -base64 24

# Or use a password manager to generate 20+ character passwords
```

**JWT Secret:**
```bash
# Generate a secure JWT secret (minimum 32 characters)
openssl rand -base64 48
```

### Development vs Production

#### Development (Local)
- Uses weak default passwords for convenience (`devpassword`)
- These defaults are ONLY for local development
- Docker Compose uses environment variable substitution: `${DB_PASSWORD:-devpassword}`

#### Production
- MUST use strong, unique passwords
- Set environment variables via:
  - Kubernetes Secrets
  - Docker Swarm Secrets
  - Cloud provider secret management (AWS Secrets Manager, Azure Key Vault, GCP Secret Manager)
  - Environment variable injection in CI/CD

### Kubernetes Secret Management

**Create database credential secret:**
```bash
kubectl create secret generic urlshortener-secrets \
  --from-literal=DATABASE_USER=urlshortener \
  --from-literal=DATABASE_PASSWORD=<strong-password-here> \
  --from-literal=JWT_SECRET=<jwt-secret-here> \
  -n production
```

**Create SSL certificate secret:**
```bash
# Generate certificates first
./scripts/generate-postgres-certs.sh

# Create secret from certificate files
kubectl create secret generic postgres-ssl-certs \
  --from-file=server.crt=certs/postgres/server.crt \
  --from-file=server.key=certs/postgres/server.key \
  --from-file=root.crt=certs/postgres/root.crt \
  -n production
```

### Docker Compose Production Deployment

**Create `.env.prod` file (NEVER commit this):**
```bash
# Copy from example
cp .env.prod.example .env.prod

# Edit and add your secure credentials
nano .env.prod
```

**Ensure `.env.prod` is in `.gitignore`:**
```bash
# Verify it's ignored
git check-ignore .env.prod
# Should output: .env.prod
```

### Testing and CI/CD

**GitHub Actions / CI Pipelines:**
- Store credentials as repository secrets
- Reference in workflows: `${{ secrets.DB_PASSWORD }}`
- Use test credentials ONLY for test databases
- Never use production credentials in CI/CD logs

**Local Testing:**
```bash
# Override defaults with environment variables
export DATABASE_URL="postgres://urlshortener:devpassword@localhost:5432/urlshortener_db?sslmode=disable"

# Run tests
make test
```

## Database Security

### PostgreSQL SSL/TLS

**Production:** PostgreSQL connections MUST use SSL/TLS encryption.

**Configuration:**
- `sslmode=require` - Production (enforced encryption)
- `sslmode=disable` - Development only

**See:** `docs/DATABASE_SECURITY.md` for detailed setup.

### Connection String Security

**NEVER log or expose connection strings:**
```go
// ❌ BAD - Logs password
log.Printf("Connecting to: %s", connectionString)

// ✅ GOOD - Redact password
log.Printf("Connecting to database: %s@%s:%s/%s", user, host, port, database)
```

## API Security

### Authentication
- JWT-based authentication for protected endpoints
- Tokens expire after configured duration
- Refresh tokens stored securely

### Rate Limiting
- Implemented at Nginx reverse proxy level
- Protects against brute force attacks
- See: `docker/nginx.conf`

## Vulnerability Reporting

If you discover a security vulnerability, please email:
- **Email:** security@trunc.site
- **Response Time:** Within 48 hours

**Please do NOT:**
- Open public GitHub issues for security vulnerabilities
- Disclose vulnerabilities publicly before receiving confirmation

## Security Checklist for Deployment

- [ ] All environment variables set with strong passwords
- [ ] `.env` and `.env.prod` files are in `.gitignore`
- [ ] PostgreSQL SSL/TLS enabled (`sslmode=require`)
- [ ] SSL certificates generated and deployed
- [ ] Kubernetes secrets created for credentials
- [ ] JWT secret is minimum 32 characters
- [ ] Database password is minimum 16 characters
- [ ] API rate limiting configured
- [ ] CORS configured for production domain only
- [ ] Production logs redact sensitive information
- [ ] Database backups are encrypted
- [ ] Security headers configured in Nginx

## Compliance

This application handles shortened URLs which may contain sensitive information. Ensure compliance with:
- GDPR (data protection)
- CCPA (California privacy)
- Industry-specific regulations

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Docker Security Best Practices](https://docs.docker.com/develop/security-best-practices/)
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/overview/)
- [PostgreSQL Security](https://www.postgresql.org/docs/current/security.html)

---

**Last Updated:** 2025-01-18
**Security Policy Version:** 1.0

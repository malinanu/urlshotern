# Secret Rotation Guide

This guide provides comprehensive procedures for rotating all secrets and credentials in the URL Shortener system. Regular secret rotation is a critical security practice that limits the impact of credential exposure.

## Table of Contents

1. [Overview](#overview)
2. [Rotation Schedule](#rotation-schedule)
3. [Database Password Rotation](#database-password-rotation)
4. [JWT Secret Rotation](#jwt-secret-rotation)
5. [API Key Rotation](#api-key-rotation)
6. [SSL/TLS Certificate Rotation](#ssltls-certificate-rotation)
7. [Kubernetes Secret Rotation](#kubernetes-secret-rotation)
8. [Third-Party Service Keys](#third-party-service-keys)
9. [Emergency Rotation](#emergency-rotation)
10. [Audit and Verification](#audit-and-verification)

---

## Overview

### Why Rotate Secrets?

- **Limit Exposure Window**: Even if a secret is compromised, rotation limits how long it remains valid
- **Compliance Requirements**: Many security frameworks (SOC 2, ISO 27001, PCI DSS) mandate regular rotation
- **Defense in Depth**: Regular rotation is an additional security layer beyond access controls
- **Incident Response**: Establishes procedures for rapid response to security events

### General Principles

1. **Zero-Downtime**: All rotation procedures should maintain service availability
2. **Rollback Plan**: Always have a rollback procedure for each rotation
3. **Validation**: Verify new secrets work before decommissioning old ones
4. **Audit Trail**: Log all rotation events for compliance and troubleshooting
5. **Secure Storage**: Use secure secret management (Vault, AWS Secrets Manager, etc.)

---

## Rotation Schedule

| Secret Type | Rotation Frequency | Priority |
|-------------|-------------------|----------|
| Database Passwords | 90 days | Critical |
| JWT Secrets | 180 days | Critical |
| API Keys (Internal) | 90 days | High |
| SSL/TLS Certificates | 90 days before expiry | Critical |
| Third-Party API Keys | Per vendor policy | Medium |
| Service Account Tokens | 30 days | High |
| Redis Password | 90 days | High |
| SMTP Credentials | 180 days | Medium |

---

## Database Password Rotation

### PostgreSQL Password Rotation (Zero-Downtime)

#### Prerequisites

- Access to PostgreSQL with SUPERUSER privileges
- Database connection string in environment variables
- Ability to update application configuration

#### Step 1: Create New Database User

```sql
-- Connect as postgres superuser
psql -U postgres

-- Create new user with same privileges
CREATE USER urlshortener_new WITH PASSWORD 'NEW_STRONG_PASSWORD_HERE';

-- Grant all privileges on existing database
GRANT ALL PRIVILEGES ON DATABASE urlshortener TO urlshortener_new;

-- Grant privileges on all tables
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO urlshortener_new;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO urlshortener_new;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO urlshortener_new;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO urlshortener_new;
```

#### Step 2: Update Application Configuration (Gradual Rollout)

**Option A: Environment Variable Update (Docker/K8s)**

```bash
# Update .env or ConfigMap
DB_USER=urlshortener_new
DB_PASSWORD=NEW_STRONG_PASSWORD_HERE

# For Kubernetes (see Kubernetes section for details)
kubectl create secret generic db-credentials \
  --from-literal=username=urlshortener_new \
  --from-literal=password=NEW_STRONG_PASSWORD_HERE \
  --dry-run=client -o yaml | kubectl apply -f -

# Rolling restart of pods
kubectl rollout restart deployment urlshortener
```

**Option B: Direct Connection String Update**

```bash
# Old connection string
DB_CONNECTION_STRING="postgresql://urlshortener:old_password@localhost:5432/urlshortener"

# New connection string
DB_CONNECTION_STRING="postgresql://urlshortener_new:NEW_STRONG_PASSWORD_HERE@localhost:5432/urlshortener"

# Restart application with new connection string
systemctl restart urlshortener
```

#### Step 3: Verify New User Works

```bash
# Test connection with new credentials
psql -U urlshortener_new -d urlshortener -c "SELECT 1;"

# Check application logs for successful database connections
tail -f /var/log/urlshortener/app.log | grep "database connected"

# Monitor for any authentication errors
tail -f /var/log/postgresql/postgresql.log | grep "authentication failed"
```

#### Step 4: Monitor Application

Wait 24-48 hours while monitoring:

```bash
# Check error logs
grep -i "database.*error" /var/log/urlshortener/*.log

# Monitor connection pool
# Check application metrics dashboard for database connection issues

# Verify no old credentials are still in use
SELECT usename, client_addr, state, query_start
FROM pg_stat_activity
WHERE usename = 'urlshortener'
ORDER BY query_start DESC;
```

#### Step 5: Revoke Old User Access

After confirming the new user works correctly:

```sql
-- Revoke all privileges from old user
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM urlshortener;
REVOKE ALL PRIVILEGES ON DATABASE urlshortener FROM urlshortener;

-- Drop old user
DROP USER urlshortener;
```

#### Rollback Procedure

If issues occur after rotation:

```bash
# Revert environment variables to old credentials
DB_USER=urlshortener
DB_PASSWORD=old_password

# Restart application
kubectl rollout undo deployment urlshortener  # Kubernetes
# or
systemctl restart urlshortener  # Systemd
```

---

## JWT Secret Rotation

JWT secret rotation requires a dual-secret validation period to avoid invalidating existing tokens.

### Strategy: Dual-Secret Validation

During rotation, the application validates tokens signed with either the old or new secret for a grace period.

#### Step 1: Add New JWT Secret

Update your configuration to support multiple secrets:

**configs/config.go**:

```go
type Config struct {
    JWTSecret    string   // Primary secret
    JWTSecretOld string   // Previous secret (for grace period)
    // ... other fields
}
```

**internal/services/jwt.go**:

```go
// ValidateToken now tries both secrets
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
    var claims Claims

    // Try validating with primary secret first
    token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
        return []byte(j.config.JWTSecret), nil
    })

    if err != nil && j.config.JWTSecretOld != "" {
        // If validation fails and old secret exists, try old secret
        token, err = jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
            return []byte(j.config.JWTSecretOld), nil
        })
    }

    if err != nil || !token.Valid {
        return nil, fmt.Errorf("invalid token: %w", err)
    }

    return &claims, nil
}
```

#### Step 2: Generate New Secret

```bash
# Generate a strong random secret (256-bit)
openssl rand -base64 32

# Example output: xK7pQ9mN2wV8yE4hR6tL3sF1dG5jB0aP9cX8zW7vU2n=
```

#### Step 3: Deploy Dual-Secret Configuration

```bash
# Set both secrets in environment
JWT_SECRET=NEW_SECRET_HERE
JWT_SECRET_OLD=OLD_SECRET_HERE

# Deploy update (zero-downtime)
kubectl set env deployment/urlshortener \
  JWT_SECRET=NEW_SECRET_HERE \
  JWT_SECRET_OLD=OLD_SECRET_HERE

kubectl rollout status deployment/urlshortener
```

#### Step 4: Grace Period

Wait for the access token TTL to expire (default: 24 hours):

- All new tokens are signed with the new secret
- Old tokens are still validated using JWT_SECRET_OLD
- Monitor logs for validation using old secret

```bash
# Monitor which secret is being used
grep "JWT validation" /var/log/urlshortener/app.log
```

#### Step 5: Remove Old Secret

After the grace period (typically 2x the token TTL):

```bash
# Remove old secret from configuration
JWT_SECRET=NEW_SECRET_HERE
# Remove JWT_SECRET_OLD

# Deploy
kubectl set env deployment/urlshortener JWT_SECRET_OLD-
kubectl rollout status deployment/urlshortener
```

#### Rollback Procedure

```bash
# If issues occur, swap the secrets back
JWT_SECRET=OLD_SECRET_HERE
JWT_SECRET_OLD=NEW_SECRET_HERE

# Deploy
kubectl rollout restart deployment/urlshortener
```

### Emergency JWT Secret Rotation

If a JWT secret is compromised, invalidate all tokens immediately:

```bash
# 1. Change JWT secret without grace period
JWT_SECRET=NEW_EMERGENCY_SECRET
# Do NOT set JWT_SECRET_OLD

# 2. Force all users to re-authenticate
redis-cli KEYS "session:*" | xargs redis-cli DEL
redis-cli KEYS "refresh_token:*" | xargs redis-cli DEL

# 3. Clear token blacklist (it's now irrelevant)
redis-cli KEYS "blacklist:*" | xargs redis-cli DEL

# 4. Deploy immediately
kubectl set env deployment/urlshortener JWT_SECRET=NEW_EMERGENCY_SECRET JWT_SECRET_OLD-
kubectl rollout restart deployment/urlshortener --grace-period=0

# 5. Notify users via email that re-authentication is required
```

---

## API Key Rotation

### User API Keys

User API keys should be rotated on a regular schedule or on-demand.

#### Step 1: Generate New API Key

```go
// internal/services/api_key.go
func (s *APIKeyService) RotateAPIKey(userID int64) (*models.APIKey, error) {
    // Generate new key
    newKey := generateSecureAPIKey()

    // Create new API key record
    apiKey := &models.APIKey{
        UserID:    userID,
        Key:       hashAPIKey(newKey),
        ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
        CreatedAt: time.Now(),
    }

    // Save to database
    err := s.db.CreateAPIKey(apiKey)
    if err != nil {
        return nil, err
    }

    return apiKey, nil
}
```

#### Step 2: Grace Period Configuration

Implement a grace period where both old and new keys work:

```sql
-- Add grace_period column to api_keys table
ALTER TABLE api_keys ADD COLUMN grace_period_ends_at TIMESTAMP;

-- Set grace period (e.g., 30 days)
UPDATE api_keys
SET grace_period_ends_at = CURRENT_TIMESTAMP + INTERVAL '30 days'
WHERE user_id = <USER_ID> AND key = '<OLD_KEY_HASH>';
```

#### Step 3: Notify User

```go
// Send email notification
emailBody := fmt.Sprintf(`
Your API key has been rotated. Your new key is:

%s

Your old key will continue to work until %s. Please update your applications before this date.

For security best practices, we recommend:
1. Update your applications immediately
2. Never commit API keys to source control
3. Use environment variables to store keys
4. Rotate keys every 90 days
`, newKey, gracePeriodEnd)

emailService.SendEmail(user.Email, "API Key Rotated", emailBody)
```

#### Step 4: Monitor Usage

```bash
# Track old vs new key usage
SELECT
    key_id,
    COUNT(*) as request_count,
    MAX(used_at) as last_used
FROM api_key_usage_logs
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY key_id;
```

#### Step 5: Revoke Old Key

After grace period expires:

```sql
-- Revoke old API key
UPDATE api_keys
SET revoked = true, revoked_at = CURRENT_TIMESTAMP
WHERE user_id = <USER_ID>
  AND grace_period_ends_at < CURRENT_TIMESTAMP;
```

### Automated Rotation Script

Create `scripts/rotate-api-keys.sh`:

```bash
#!/bin/bash
# Rotate API keys approaching expiration

set -e

# Configuration
DAYS_BEFORE_EXPIRY=30
GRACE_PERIOD_DAYS=30

# Find keys expiring soon
psql -U postgres -d urlshortener <<EOF
SELECT user_id, email, key_id, expires_at
FROM api_keys
JOIN users ON api_keys.user_id = users.id
WHERE expires_at < CURRENT_TIMESTAMP + INTERVAL '$DAYS_BEFORE_EXPIRY days'
  AND revoked = false;
EOF

# TODO: Implement automated rotation
# 1. Generate new keys
# 2. Send email notifications
# 3. Set grace periods
```

---

## SSL/TLS Certificate Rotation

### Let's Encrypt Certificate Renewal (Automated)

#### Setup Automated Renewal with Certbot

```bash
# Install certbot
sudo apt-get install certbot python3-certbot-nginx

# Initial certificate generation
sudo certbot --nginx -d urlshort.example.com -d www.urlshort.example.com

# Test renewal
sudo certbot renew --dry-run

# Automated renewal via cron (runs twice daily)
sudo crontab -e
# Add: 0 0,12 * * * certbot renew --quiet --post-hook "systemctl reload nginx"
```

#### Manual Certificate Rotation

For custom certificates (not Let's Encrypt):

#### Step 1: Generate New Certificate

```bash
# Generate new private key
openssl genrsa -out urlshort.example.com.key 4096

# Generate CSR (Certificate Signing Request)
openssl req -new -key urlshort.example.com.key -out urlshort.example.com.csr

# Send CSR to Certificate Authority and receive signed certificate
# Save as urlshort.example.com.crt
```

#### Step 2: Verify Certificate

```bash
# Verify certificate matches private key
openssl x509 -noout -modulus -in urlshort.example.com.crt | openssl md5
openssl rsa -noout -modulus -in urlshort.example.com.key | openssl md5
# Both MD5 hashes should match

# Check certificate expiration
openssl x509 -in urlshort.example.com.crt -noout -dates

# Verify certificate chain
openssl verify -CAfile ca-bundle.crt urlshort.example.com.crt
```

#### Step 3: Deploy New Certificate

**For Nginx:**

```bash
# Backup old certificates
sudo cp /etc/nginx/ssl/urlshort.example.com.crt /etc/nginx/ssl/urlshort.example.com.crt.backup
sudo cp /etc/nginx/ssl/urlshort.example.com.key /etc/nginx/ssl/urlshort.example.com.key.backup

# Copy new certificates
sudo cp urlshort.example.com.crt /etc/nginx/ssl/
sudo cp urlshort.example.com.key /etc/nginx/ssl/
sudo chmod 600 /etc/nginx/ssl/urlshort.example.com.key

# Test nginx configuration
sudo nginx -t

# Reload nginx (zero-downtime)
sudo systemctl reload nginx
```

**For Kubernetes:**

```bash
# Create TLS secret
kubectl create secret tls urlshort-tls \
  --cert=urlshort.example.com.crt \
  --key=urlshort.example.com.key \
  --dry-run=client -o yaml | kubectl apply -f -

# Update Ingress to use new secret (if needed)
# Nginx Ingress Controller automatically reloads on secret update
```

#### Step 4: Verify Deployment

```bash
# Test HTTPS connection
curl -vI https://urlshort.example.com

# Check certificate details
echo | openssl s_client -servername urlshort.example.com \
  -connect urlshort.example.com:443 2>/dev/null | openssl x509 -noout -dates

# Use SSL Labs for comprehensive check
# https://www.ssllabs.com/ssltest/analyze.html?d=urlshort.example.com
```

---

## Kubernetes Secret Rotation

### Using Kubernetes Secrets

#### Step 1: Create New Secret Version

```bash
# Create new secret (v2)
kubectl create secret generic db-credentials-v2 \
  --from-literal=username=urlshortener_new \
  --from-literal=password=NEW_PASSWORD_HERE \
  --dry-run=client -o yaml > db-credentials-v2.yaml

# Apply new secret
kubectl apply -f db-credentials-v2.yaml
```

#### Step 2: Update Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: urlshortener
spec:
  template:
    spec:
      containers:
      - name: urlshortener
        env:
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: db-credentials-v2  # Updated reference
              key: username
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-credentials-v2  # Updated reference
              key: password
```

```bash
# Apply deployment update
kubectl apply -f deployment.yaml

# Rolling restart
kubectl rollout restart deployment urlshortener
kubectl rollout status deployment urlshortener
```

#### Step 3: Verify and Clean Up

```bash
# Verify all pods are running with new secret
kubectl get pods

# Check logs for errors
kubectl logs -l app=urlshortener --tail=100

# After confirmation (wait 24-48 hours), delete old secret
kubectl delete secret db-credentials
```

### Using Sealed Secrets

For GitOps workflows with Sealed Secrets:

```bash
# Install kubeseal CLI
wget https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.18.0/kubeseal-linux-amd64 -O kubeseal
chmod +x kubeseal
sudo mv kubeseal /usr/local/bin/

# Create sealed secret
kubectl create secret generic db-credentials-v2 \
  --from-literal=username=urlshortener_new \
  --from-literal=password=NEW_PASSWORD_HERE \
  --dry-run=client -o yaml | \
  kubeseal -o yaml > sealed-db-credentials-v2.yaml

# Commit to Git
git add sealed-db-credentials-v2.yaml
git commit -m "Rotate database credentials"
git push

# ArgoCD or FluxCD will automatically apply the new secret
```

---

## Third-Party Service Keys

### AWS Credentials Rotation

#### Step 1: Create New IAM Access Key

```bash
# Using AWS CLI
aws iam create-access-key --user-name urlshortener-service

# Output:
# {
#     "AccessKey": {
#         "UserName": "urlshortener-service",
#         "AccessKeyId": "AKIAIOSFODNN7EXAMPLE",
#         "Status": "Active",
#         "SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
#         "CreateDate": "2024-01-15T12:00:00Z"
#     }
# }
```

#### Step 2: Update Application Configuration

```bash
# Update environment variables
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

# For Kubernetes
kubectl create secret generic aws-credentials \
  --from-literal=access_key_id=AKIAIOSFODNN7EXAMPLE \
  --from-literal=secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart application
kubectl rollout restart deployment urlshortener
```

#### Step 3: Verify New Credentials Work

```bash
# Test AWS connection
aws s3 ls --profile urlshortener  # or using environment variables

# Check application logs
kubectl logs -l app=urlshortener | grep -i "aws\|s3"
```

#### Step 4: Deactivate Old Key

```bash
# List all access keys
aws iam list-access-keys --user-name urlshortener-service

# Deactivate old key
aws iam update-access-key \
  --access-key-id AKIAIOSFODNN7OLDKEY \
  --status Inactive \
  --user-name urlshortener-service

# After verification (24-48 hours), delete old key
aws iam delete-access-key \
  --access-key-id AKIAIOSFODNN7OLDKEY \
  --user-name urlshortener-service
```

### SendGrid API Key Rotation

```bash
# 1. Generate new API key in SendGrid dashboard
# Settings > API Keys > Create API Key

# 2. Update application
kubectl create secret generic sendgrid-credentials \
  --from-literal=api_key=SG.NEW_KEY_HERE \
  --dry-run=client -o yaml | kubectl apply -f -

# 3. Restart application
kubectl rollout restart deployment urlshortener

# 4. Verify email sending works
curl -X POST http://localhost:8080/api/v1/test-email

# 5. Delete old key in SendGrid dashboard
```

### Fastly API Token Rotation

```bash
# 1. Generate new token in Fastly dashboard
# Account > Personal API Tokens > Create Token

# 2. Update configuration
kubectl create secret generic fastly-credentials \
  --from-literal=api_token=NEW_TOKEN_HERE \
  --dry-run=client -o yaml | kubectl apply -f -

# 3. Test CDN purge
curl -X POST http://localhost:8080/api/v1/admin/purge-cache

# 4. Revoke old token in Fastly dashboard
```

---

## Emergency Rotation

When credentials are compromised or suspected to be compromised, immediate rotation is required.

### Emergency Rotation Checklist

1. **Identify Scope**
   - Which secret(s) were compromised?
   - What systems have access?
   - What is the potential impact?

2. **Immediate Actions**
   ```bash
   # Disable compromised credentials immediately
   # Example for database:
   psql -U postgres -c "ALTER USER urlshortener WITH PASSWORD 'TEMP_SECURE_PASSWORD';"

   # Example for AWS:
   aws iam update-access-key --access-key-id COMPROMISED_KEY --status Inactive
   ```

3. **Generate New Credentials**
   ```bash
   # Generate strong random passwords
   openssl rand -base64 32

   # Update all systems simultaneously
   ```

4. **Deploy Updates**
   ```bash
   # Force immediate rollout without grace period
   kubectl set env deployment/urlshortener DB_PASSWORD=NEW_PASSWORD
   kubectl rollout restart deployment urlshortener --grace-period=0
   ```

5. **Revoke Compromised Credentials**
   ```bash
   # Drop old user/key immediately after new credentials are deployed
   ```

6. **Audit and Investigation**
   - Review access logs
   - Check for unauthorized access
   - Determine how credentials were compromised
   - Implement preventive measures

### Emergency Rotation Script

Create `scripts/emergency-rotate.sh`:

```bash
#!/bin/bash
# Emergency credential rotation script

set -e

echo "EMERGENCY CREDENTIAL ROTATION"
echo "============================="
echo ""
echo "WARNING: This will immediately rotate credentials without grace period."
read -p "Are you sure you want to continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Aborted."
    exit 1
fi

# Function to generate secure password
generate_password() {
    openssl rand -base64 32
}

# Rotate database password
echo "Rotating database password..."
NEW_DB_PASS=$(generate_password)
psql -U postgres <<EOF
ALTER USER urlshortener WITH PASSWORD '$NEW_DB_PASS';
EOF

# Update Kubernetes secret
kubectl create secret generic db-credentials \
  --from-literal=password=$NEW_DB_PASS \
  --dry-run=client -o yaml | kubectl apply -f -

# Force immediate restart
kubectl rollout restart deployment urlshortener --grace-period=0

echo "Emergency rotation complete."
echo "NEW_DB_PASS has been set. Please store securely."
```

---

## Audit and Verification

### Rotation Audit Log

Maintain a rotation audit log in your database:

```sql
CREATE TABLE secret_rotation_log (
    id SERIAL PRIMARY KEY,
    secret_type VARCHAR(100) NOT NULL,
    rotated_by VARCHAR(255),
    rotation_reason VARCHAR(50), -- scheduled, emergency, compliance
    old_secret_hash VARCHAR(255), -- hash of old secret for tracking
    new_secret_hash VARCHAR(255), -- hash of new secret
    rotation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    grace_period_end TIMESTAMP,
    notes TEXT
);

-- Log rotation event
INSERT INTO secret_rotation_log (
    secret_type, rotated_by, rotation_reason, notes
) VALUES (
    'database_password', 'admin@example.com', 'scheduled', 'Quarterly rotation'
);
```

### Verification Checklist

After each rotation:

- [ ] New secret works correctly
- [ ] Application logs show no authentication errors
- [ ] Old secret is revoked/deleted (after grace period)
- [ ] Rotation is logged in audit trail
- [ ] Documentation is updated
- [ ] Team is notified
- [ ] Backup systems are updated

### Monitoring

Set up alerts for:

```yaml
# Prometheus alert rules
groups:
- name: secret_expiry
  rules:
  - alert: CertificateExpiringSoon
    expr: (ssl_certificate_expiry_seconds < 86400 * 30)
    annotations:
      summary: "SSL certificate expiring in 30 days"

  - alert: APIKeyExpiringSoon
    expr: (api_key_expiry_timestamp - time() < 86400 * 30)
    annotations:
      summary: "API key expiring in 30 days"
```

---

## Best Practices

1. **Use Secret Management Tools**
   - HashiCorp Vault
   - AWS Secrets Manager
   - Azure Key Vault
   - Google Secret Manager

2. **Automate Where Possible**
   - Use cert-manager for certificate rotation
   - Implement automated API key rotation
   - Schedule regular rotation via cron/scheduled jobs

3. **Test Rotation Procedures**
   - Practice rotation in staging environment
   - Document any issues encountered
   - Update procedures based on lessons learned

4. **Secure Secret Storage**
   - Never commit secrets to Git
   - Use environment variables or secret managers
   - Encrypt secrets at rest
   - Restrict access with RBAC

5. **Grace Periods**
   - Always use grace periods for non-emergency rotations
   - Typically 2x the token/session TTL
   - Monitor during grace period

6. **Documentation**
   - Keep rotation dates in a secure log
   - Document who performed rotation
   - Note any issues or special circumstances

---

## Additional Resources

- [OWASP Secret Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_CheatSheet.html)
- [NIST SP 800-57: Key Management](https://csrc.nist.gov/publications/detail/sp/800-57-part-1/rev-5/final)
- [CIS Controls: Secure Configuration](https://www.cisecurity.org/controls)

---

## Support

For questions or issues with secret rotation:

1. Check troubleshooting guide: [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
2. Review audit logs for rotation history
3. Contact security team: security@example.com
4. Create GitHub issue: https://github.com/yourusername/URLshorter/issues

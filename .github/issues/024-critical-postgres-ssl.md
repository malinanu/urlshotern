# [CRITICAL] Enable PostgreSQL SSL/TLS in Production

**Labels:** `critical`, `security`, `infra`, `quick-win`

## Description
PostgreSQL is configured with SSL/TLS disabled (`ssl = off`) in the production Kubernetes deployment, exposing database traffic to potential interception.

## Severity: CRITICAL
This is critical because:
- Database credentials transmitted in plaintext
- SQL queries and results visible on network
- Compliance violations (PCI DSS, HIPAA, SOC 2)
- Man-in-the-middle attack vulnerability
- Production security requirement

## Affected Files
- [ ] `k8s/production/postgres.yaml` (line 149)

## Current Behavior
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-config
data:
  postgresql.conf: |
    listen_addresses = '*'
    max_connections = 100
    shared_buffers = 128MB
    ssl = off  # <-- SSL DISABLED
```

Database connections are not encrypted, allowing anyone with network access to intercept database traffic.

## Expected Behavior
- SSL/TLS enabled for all database connections
- Server certificates properly configured
- Client certificate validation (optional but recommended)
- Encrypted communication between app and database
- Compliance with security standards

## Proposed Solution

### 1. Generate SSL Certificates
```bash
# Create certificate authority
openssl req -new -x509 -days 365 -nodes -text \
  -out server.crt -keyout server.key -subj "/CN=postgres"

# Set proper permissions
chmod 600 server.key
chown postgres:postgres server.key server.crt
```

### 2. Update PostgreSQL Configuration
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-config
data:
  postgresql.conf: |
    listen_addresses = '*'
    max_connections = 100
    shared_buffers = 128MB
    ssl = on
    ssl_cert_file = '/etc/ssl/certs/server.crt'
    ssl_key_file = '/etc/ssl/private/server.key'
    ssl_ca_file = '/etc/ssl/certs/ca.crt'  # if using client certs
    ssl_ciphers = 'HIGH:MEDIUM:+3DES:!aNULL'
    ssl_prefer_server_ciphers = on
    ssl_min_protocol_version = 'TLSv1.2'
```

### 3. Mount Certificates as Kubernetes Secrets
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: postgres-ssl-certs
type: Opaque
data:
  server.crt: <base64-encoded-cert>
  server.key: <base64-encoded-key>
  ca.crt: <base64-encoded-ca>
---
apiVersion: apps/v1
kind: StatefulSet
spec:
  template:
    spec:
      containers:
      - name: postgres
        volumeMounts:
        - name: ssl-certs
          mountPath: /etc/ssl/certs
          readOnly: true
      volumes:
      - name: ssl-certs
        secret:
          secretName: postgres-ssl-certs
          defaultMode: 0600
```

### 4. Update Connection Strings
```go
// In configs/config.go
DB_SSL_MODE=require  // or "verify-full" for cert validation
```

## Acceptance Criteria
- [ ] SSL/TLS enabled in PostgreSQL configuration
- [ ] Server certificates generated and mounted
- [ ] Connection strings updated to require SSL
- [ ] Existing connections still work
- [ ] Certificate expiration monitoring configured
- [ ] Documentation updated with SSL setup instructions
- [ ] Tested in staging environment
- [ ] Security scan confirms encrypted connections

## Implementation Steps
1. Generate SSL certificates (cert-manager or manual)
2. Create Kubernetes secret with certificates
3. Update PostgreSQL ConfigMap to enable SSL
4. Mount certificates in PostgreSQL pods
5. Update application connection strings
6. Test in staging
7. Deploy to production during maintenance window
8. Monitor logs for SSL errors
9. Update documentation

## Security Impact
- **Current Risk**: High - unencrypted database traffic
- **Mitigation**: Full encryption of database communications
- **Compliance**: Meets PCI DSS, HIPAA, SOC 2 requirements

## Rollback Plan
If issues arise:
1. Set `ssl = off` in ConfigMap
2. Remove SSL requirement from connection strings
3. Rollback pod configuration
4. Investigate and fix SSL issues
5. Retry deployment

## Related Issues
- #26 (Remove Hardcoded Database Credentials) - should be done together
- Security audit checklist

## References
- [PostgreSQL SSL Documentation](https://www.postgresql.org/docs/current/ssl-tcp.html)
- [Kubernetes Secrets for Certificate Management](https://kubernetes.io/docs/concepts/configuration/secret/)

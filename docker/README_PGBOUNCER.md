# PgBouncer Configuration

This document explains the database connection pooling setup using PgBouncer.

## Overview

PgBouncer is a lightweight connection pooler for PostgreSQL. It sits between the application and PostgreSQL database, managing a pool of database connections to improve performance and resource utilization.

## Architecture

```
Application (Port 8080)
    ↓
PgBouncer (Port 6432)
    ↓
PostgreSQL (Port 5432)
```

## Configuration

### Pool Mode
- **Mode**: `transaction` - Connection is returned to pool after transaction completes
- Alternative modes: `session` (one connection per client session) or `statement` (per SQL statement)

### Connection Limits
- **MAX_CLIENT_CONN**: 1000 - Maximum number of client connections
- **DEFAULT_POOL_SIZE**: 25 - Default connections per database
- **MIN_POOL_SIZE**: 10 - Minimum connections to keep open
- **RESERVE_POOL_SIZE**: 5 - Additional connections for emergencies
- **MAX_DB_CONNECTIONS**: 100 - Maximum total database connections
- **MAX_USER_CONNECTIONS**: 100 - Maximum connections per user

### Timeouts
- **SERVER_IDLE_TIMEOUT**: 600s - Close idle server connections after 10 minutes
- **SERVER_LIFETIME**: 3600s - Close server connections after 1 hour
- **SERVER_CONNECT_TIMEOUT**: 15s - Timeout for new server connections
- **QUERY_WAIT_TIMEOUT**: 120s - Maximum time queries can wait for connection

### Application Connection Pool Settings

The Go application also maintains its own connection pool with the following settings (configurable via environment variables):

- **DB_MAX_CONNECTIONS**: 25 - Maximum open connections
- **DB_MAX_IDLE_CONNECTIONS**: 10 - Maximum idle connections
- **DB_CONN_MAX_LIFETIME**: 3600s - Maximum connection lifetime
- **DB_CONN_MAX_IDLE_TIME**: 600s - Maximum idle time before closing

## Monitoring

### PgBouncer Admin Console

Connect to PgBouncer admin console:
```bash
psql -h localhost -p 6432 -U urlshortener -d pgbouncer
```

Useful commands:
```sql
-- Show pool statistics
SHOW POOLS;

-- Show client connections
SHOW CLIENTS;

-- Show server connections
SHOW SERVERS;

-- Show database statistics
SHOW STATS;

-- Show configuration
SHOW CONFIG;

-- Reload configuration
RELOAD;

-- Pause all activity
PAUSE;

-- Resume activity
RESUME;
```

### Application Pool Stats

Access connection pool statistics via HTTP endpoint:
```bash
curl http://localhost:8080/health/db-pool
```

Response example:
```json
{
  "status": "success",
  "data": {
    "max_open_connections": 25,
    "open_connections": 5,
    "in_use": 2,
    "idle": 3,
    "wait_count": 0,
    "wait_duration": "0s",
    "max_idle_closed": 0,
    "max_idle_time_closed": 0,
    "max_lifetime_closed": 0
  }
}
```

### Metrics Explanation

- **max_open_connections**: Configured maximum
- **open_connections**: Currently open connections
- **in_use**: Connections actively being used
- **idle**: Idle connections in pool
- **wait_count**: Number of times waited for connection
- **wait_duration**: Total time spent waiting
- **max_idle_closed**: Connections closed due to idle limit
- **max_idle_time_closed**: Connections closed due to idle time
- **max_lifetime_closed**: Connections closed due to max lifetime

## Performance Tuning

### For High Traffic

```env
DB_MAX_CONNECTIONS=50
DB_MAX_IDLE_CONNECTIONS=20
DEFAULT_POOL_SIZE=50
MAX_CLIENT_CONN=2000
```

### For Low Traffic

```env
DB_MAX_CONNECTIONS=10
DB_MAX_IDLE_CONNECTIONS=5
DEFAULT_POOL_SIZE=10
MAX_CLIENT_CONN=100
```

### For Long-Running Queries

```env
QUERY_WAIT_TIMEOUT=300
SERVER_IDLE_TIMEOUT=1800
```

## Troubleshooting

### High Wait Count
If `wait_count` is increasing rapidly:
1. Increase `DB_MAX_CONNECTIONS`
2. Increase `DEFAULT_POOL_SIZE` in PgBouncer
3. Check for slow queries
4. Consider horizontal scaling

### Connection Leaks
If `in_use` stays high:
1. Check application code for unclosed connections
2. Review transaction handling
3. Monitor `max_lifetime_closed` - should increase over time

### PgBouncer Not Accepting Connections
1. Check PgBouncer logs: `docker logs <pgbouncer-container>`
2. Verify PostgreSQL is healthy
3. Check network connectivity
4. Verify credentials in DATABASE_URL

## Best Practices

1. **Monitor Regularly**: Check pool stats during peak traffic
2. **Set Appropriate Limits**: Match pool size to expected concurrent queries
3. **Use Transaction Mode**: Best for most web applications
4. **Configure Timeouts**: Prevent resource exhaustion
5. **Log Connection Issues**: Enable logging for troubleshooting
6. **Scale Horizontally**: Add more app instances before increasing pool size too much
7. **Keep Connection Lifetime Reasonable**: Prevents stale connections
8. **Test Under Load**: Verify configuration handles expected traffic

## Production Deployment

For production environments:
1. Use strong passwords (minimum 32 characters)
2. Enable SSL/TLS for PostgreSQL connections
3. Set up monitoring and alerting on pool metrics
4. Configure appropriate resource limits in Docker
5. Use separate PgBouncer instances for different environments
6. Implement connection retry logic in application
7. Set up log aggregation for PgBouncer logs
8. Regular health checks on PgBouncer endpoint

## References

- [PgBouncer Documentation](https://www.pgbouncer.org/usage.html)
- [PostgreSQL Connection Pooling](https://www.postgresql.org/docs/current/runtime-config-connection.html)
- [Go database/sql Package](https://pkg.go.dev/database/sql)

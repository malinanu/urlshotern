# API Contract Tests

This directory contains Postman collections for API contract testing.

## Overview

The `api-tests.json` collection provides comprehensive contract tests for all public API endpoints of the URL Shortener service.

## Test Coverage

### Health Checks
- `/health` - Basic health check
- `/health/db-pool` - Database connection pool statistics
- `/health/goroutines` - Goroutine error monitoring

### URL Shortening
- Create short URL with valid URL
- Create short URL with invalid URL (validation test)
- Create short URL with missing URL (validation test)
- Create short URL with custom code

### URL Redirection
- Redirect with valid short code
- Redirect with invalid short code (404 test)

### Analytics
- Get analytics for valid short code
- Get analytics for invalid short code (404 test)

### Performance Tests
- Response time validation for shorten endpoint

## Running Tests Locally

### Prerequisites
- Node.js and npm installed
- Newman CLI tool

### Installation
```bash
npm install -g newman
```

### Run Tests
```bash
# Against local instance
newman run tests/postman/api-tests.json --env-var "base_url=http://localhost:8080"

# Against staging
newman run tests/postman/api-tests.json --env-var "base_url=https://staging.yourdomain.com"

# With HTML report
newman run tests/postman/api-tests.json \
  --env-var "base_url=http://localhost:8080" \
  --reporters cli,html \
  --reporter-html-export newman-report.html
```

## CI/CD Integration

These tests are automatically run in the CI/CD pipeline:

1. **On Pull Requests**: Run against development environment
2. **On Staging Deployment**: Run after successful staging deployment
3. **Before Production**: Must pass before production deployment

See `.github/workflows/ci-cd.yml` for the full CI/CD configuration.

## Test Structure

Each test includes:

1. **Pre-request Scripts**: Setup environment variables, generate test data
2. **Request Definition**: HTTP method, headers, body, URL
3. **Tests**: Assertions on response status, structure, and data
4. **Post-request Scripts**: Cleanup, store variables for dependent tests

## Writing New Tests

### Adding a New Test

1. **Choose Parent Folder**: Select the appropriate folder (e.g., "URL Shortening")

2. **Create Request**: Define the HTTP request

3. **Add Tests**: Write test assertions
```javascript
pm.test('Status code is 200', function () {
    pm.response.to.have.status(200);
});

pm.test('Response has required fields', function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property('field_name');
});
```

4. **Use Environment Variables**: Reference dynamic values
```javascript
pm.environment.get('variable_name')
pm.environment.set('variable_name', value)
```

## Test Categories

### Contract Tests
Verify API contracts:
- Response structure
- Field types and values
- Required fields
- HTTP status codes

### Validation Tests
Verify input validation:
- Invalid data handling
- Missing fields
- Boundary conditions
- Error messages

### Performance Tests
Verify performance requirements:
- Response time thresholds
- Throughput capacity
- Resource utilization

## Best Practices

1. **Independent Tests**: Each test should be self-contained
2. **Clear Names**: Use descriptive test names
3. **Meaningful Assertions**: Test what matters
4. **Clean Up**: Remove test data after execution
5. **Environment Variables**: Use variables for URLs and tokens
6. **Error Handling**: Test both success and failure scenarios

## Troubleshooting

### Test Failures

If tests fail:

1. **Check Service Health**: Ensure the service is running
2. **Verify Base URL**: Check the `base_url` environment variable
3. **Review Logs**: Check service logs for errors
4. **Validate Data**: Ensure test data is valid
5. **Check Dependencies**: Verify database and Redis are accessible

### Common Issues

**Connection Refused**
```
Error: connect ECONNREFUSED
```
Solution: Ensure the service is running at the specified base URL

**Authentication Errors**
```
Error: 401 Unauthorized
```
Solution: Check if authentication token is valid and not expired

**Timeout**
```
Error: ESOCKETTIMEDOUT
```
Solution: Increase timeout or check service performance

## Monitoring

### Success Metrics
- Pass Rate: > 95%
- Response Time: < 1000ms average
- Error Rate: < 1%

### Alerts
Set up alerts for:
- Test failure rate > 5%
- Average response time > 1500ms
- Any 5xx errors

## Future Enhancements

Planned improvements:

1. **Authentication Tests**: Add tests for OAuth and JWT flows
2. **Load Tests**: Integrate with K6 or JMeter
3. **Security Tests**: Add OWASP security checks
4. **Mock Data**: Generate realistic test data
5. **Parallel Execution**: Run tests concurrently
6. **Database Validation**: Verify data persistence

## References

- [Newman Documentation](https://learning.postman.com/docs/running-collections/using-newman-cli/command-line-integration-with-newman/)
- [Postman Test Scripts](https://learning.postman.com/docs/writing-scripts/test-scripts/)
- [API Testing Best Practices](https://www.postman.com/api-platform/api-testing/)

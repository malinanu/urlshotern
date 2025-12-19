# Documentation Index

Welcome to the URL Shortener documentation. This index provides quick access to all documentation resources organized by topic.

## Table of Contents

- [Quick Start](#quick-start)
- [Development](#development)
- [Architecture & Design](#architecture--design)
- [API Documentation](#api-documentation)
- [Security](#security)
- [Deployment](#deployment)
- [Operations](#operations)
- [Contributing](#contributing)

---

## Quick Start

**New to the project? Start here.**

- [README](../README.md) - Project overview, features, and quick setup
- [Getting Started Guide](guides/GETTING_STARTED.md) - Step-by-step setup for local development
- [Configuration Reference](CONFIGURATION.md) - Environment variables and configuration options

---

## Development

**Resources for developers working on the codebase.**

### Setup & Environment
- [Local Development Setup](guides/LOCAL_DEVELOPMENT.md) - Detailed local environment setup
- [Docker Development](guides/DOCKER_DEVELOPMENT.md) - Using Docker for development
- [Environment Variables](CONFIGURATION.md#environment-variables) - Complete variable reference

### Code Structure
- [Architecture Overview](ARCHITECTURE.md) - System architecture and design decisions
- [Database Schema](DATABASE_SCHEMA.md) - Complete database schema documentation
- [API Endpoints](API_DOCUMENTATION.md) - REST API endpoint reference

### Development Guides
- [Adding New Features](guides/ADDING_FEATURES.md) - Guidelines for feature development
- [Database Migrations](guides/MIGRATION_GUIDE.md) - Creating and running migrations
- [Testing Guide](guides/TESTING_GUIDE.md) - Writing and running tests
- [Code Style Guide](guides/CODE_STYLE.md) - Coding standards and best practices

### Troubleshooting
- [Troubleshooting Guide](TROUBLESHOOTING.md) - Common issues and solutions
- [Debug Guide](guides/DEBUGGING.md) - Debugging techniques and tools

---

## Architecture & Design

**Understanding the system design and technical decisions.**

- [Architecture Overview](ARCHITECTURE.md) - High-level system architecture
- [Design Decisions](guides/DESIGN_DECISIONS.md) - ADRs and architectural choices
- [Database Schema](DATABASE_SCHEMA.md) - Entity relationships and table structure
- [Caching Strategy](guides/CACHING_STRATEGY.md) - Redis caching implementation
- [ID Generation](guides/ID_GENERATION.md) - Snowflake ID and Base62 encoding
- [URL Shortening Algorithm](guides/URL_SHORTENING.md) - Core shortening logic

---

## API Documentation

**REST API reference and integration guides.**

- [API Documentation](API_DOCUMENTATION.md) - Complete API endpoint reference
- [Authentication](guides/AUTHENTICATION.md) - JWT authentication and authorization
- [Rate Limiting](guides/RATE_LIMITING.md) - API rate limiting configuration
- [Error Handling](guides/ERROR_HANDLING.md) - API error codes and responses
- [Webhooks](guides/WEBHOOKS.md) - Webhook integration guide

---

## Security

**Security best practices and procedures.**

- [Security Best Practices](SECURITY.md) - Security guidelines and threat model
- [Secret Rotation](SECRETS_ROTATION.md) - Rotating credentials and secrets
- [Authentication & Authorization](guides/AUTHENTICATION.md) - Auth implementation details
- [RBAC System](guides/RBAC.md) - Role-based access control
- [Security Audit Guide](guides/SECURITY_AUDIT.md) - Running security audits
- [Data Privacy](guides/DATA_PRIVACY.md) - GDPR compliance and data handling

---

## Deployment

**Deploying to production environments.**

### Cloud Platforms
- [AWS Deployment](DEPLOYMENT_GUIDES/AWS.md) - Deploying to Amazon Web Services
- [Azure Deployment](DEPLOYMENT_GUIDES/AZURE.md) - Deploying to Microsoft Azure
- [GCP Deployment](DEPLOYMENT_GUIDES/GCP.md) - Deploying to Google Cloud Platform
- [DigitalOcean Deployment](DEPLOYMENT_GUIDES/DIGITALOCEAN.md) - Deploying to DigitalOcean

### Container Orchestration
- [Docker Compose Production](DEPLOYMENT_GUIDES/DOCKER_COMPOSE_PRODUCTION.md) - Production Docker Compose setup
- [Kubernetes Deployment](DEPLOYMENT_GUIDES/KUBERNETES.md) - Kubernetes deployment guide
- [Helm Charts](DEPLOYMENT_GUIDES/HELM.md) - Using Helm for Kubernetes deployment

### Deployment Strategies
- [Zero-Downtime Deployment](guides/ZERO_DOWNTIME_DEPLOYMENT.md) - Blue-green and rolling deployments
- [Multi-Region Deployment](DEPLOYMENT_GUIDES/MULTI_REGION.md) - Multi-region architecture
- [CDN Integration](guides/CDN_INTEGRATION.md) - CloudFront and Fastly setup
- [SSL/TLS Setup](guides/SSL_SETUP.md) - Certificate management and HTTPS

---

## Operations

**Running and maintaining the production system.**

### Monitoring & Observability
- [Monitoring Guide](guides/MONITORING.md) - Metrics, logs, and traces
- [Alerting](guides/ALERTING.md) - Alert configuration and incident response
- [Performance Monitoring](guides/PERFORMANCE.md) - Performance tuning and optimization
- [Log Management](guides/LOGGING.md) - Centralized logging and analysis

### Maintenance
- [Backup & Recovery](guides/BACKUP_RECOVERY.md) - Backup strategies and disaster recovery
- [Database Maintenance](guides/DATABASE_MAINTENANCE.md) - PostgreSQL maintenance tasks
- [Cache Maintenance](guides/CACHE_MAINTENANCE.md) - Redis maintenance and optimization
- [Scaling Guide](guides/SCALING.md) - Horizontal and vertical scaling

### Troubleshooting
- [Troubleshooting Guide](TROUBLESHOOTING.md) - Common production issues
- [Incident Response](guides/INCIDENT_RESPONSE.md) - Handling production incidents
- [Runbook](guides/RUNBOOK.md) - Operational procedures and checklists

---

## Contributing

**Contributing to the project.**

- [Contributing Guidelines](../CONTRIBUTING.md) - How to contribute
- [Code of Conduct](../CODE_OF_CONDUCT.md) - Community guidelines
- [Development Workflow](guides/DEVELOPMENT_WORKFLOW.md) - Git workflow and PR process
- [Code Review Guide](guides/CODE_REVIEW.md) - Code review best practices
- [Release Process](guides/RELEASE_PROCESS.md) - Versioning and release procedures

---

## Quick Links

### Common Tasks
- [Create a new migration](guides/MIGRATION_GUIDE.md#creating-migrations)
- [Add a new API endpoint](guides/ADDING_FEATURES.md#adding-api-endpoints)
- [Configure email service](CONFIGURATION.md#email-configuration)
- [Setup CDN integration](guides/CDN_INTEGRATION.md)
- [Rotate secrets](SECRETS_ROTATION.md)
- [Scale the application](guides/SCALING.md)

### External Resources
- [GitHub Repository](https://github.com/URLshorter/url-shortener)
- [Issue Tracker](https://github.com/URLshorter/url-shortener/issues)
- [Discussions](https://github.com/URLshorter/url-shortener/discussions)

---

## Documentation Status

| Document | Status | Last Updated |
|----------|--------|--------------|
| README.md | ✅ Complete | Current |
| TROUBLESHOOTING.md | ✅ Complete | Current |
| SECRETS_ROTATION.md | ✅ Complete | Current |
| CONFIGURATION.md | ✅ Complete | Current |
| API_DOCUMENTATION.md | ⚠️ Partial | Needs update |
| ARCHITECTURE.md | ⚠️ Partial | Needs update |
| DATABASE_SCHEMA.md | ⚠️ Partial | Needs update |
| Deployment Guides | 📝 In Progress | Various |

**Legend:**
- ✅ Complete and up-to-date
- ⚠️ Partial or needs updating
- 📝 In progress
- ❌ Not yet created

---

## Contributing to Documentation

Found an error or want to improve the documentation? Please:

1. Check if an issue already exists
2. Create a new issue or submit a pull request
3. Follow the documentation style guide
4. Update this index if adding new documentation

For questions about documentation, open a [GitHub Discussion](https://github.com/URLshorter/url-shortener/discussions).

---

**Last Updated:** 2025-12-19
**Version:** 1.0.0

# Observability Stack for URL Shortener Service

**Complete production-grade monitoring, logging, tracing, and alerting infrastructure**

## Overview

This observability stack provides comprehensive visibility into the URL shortener service with:

- **MTTD < 2 minutes** (Mean Time To Detect)
- **99.95% Availability SLO** with error budget tracking
- **P95 Latency < 1s, P99 < 2s** SLO enforcement
- **< 0.1% Error Rate** SLO target
- Full distributed tracing for performance debugging
- Structured logging with centralized aggregation
- Business metrics tracking (URLs created, clicks, revenue)
- Automated alerting with intelligent routing
- Production-ready runbooks for common incidents

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      URL Shortener Application                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Metrics    │  │   Traces     │  │   Logs       │          │
│  │ (Prometheus) │  │ (OpenTel)    │  │   (Zap)      │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
└─────────┼──────────────────┼──────────────────┼──────────────────┘
          │                  │                  │
          ▼                  ▼                  ▼
   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
   │  Prometheus  │   │    Jaeger    │   │   Promtail   │
   │  (Metrics)   │   │  (Tracing)   │   │ (Log Fwd)    │
   └──────┬───────┘   └──────┬───────┘   └──────┬───────┘
          │                  │                  │
          │                  │                  ▼
          │                  │            ┌──────────────┐
          │                  │            │     Loki     │
          │                  │            │ (Log Aggr)   │
          │                  │            └──────┬───────┘
          │                  │                   │
          ▼                  ▼                   ▼
   ┌──────────────────────────────────────────────────────┐
   │                      Grafana                          │
   │    (Unified Visualization & Dashboards)               │
   └──────────────────────────────────────────────────────┘
          ▲
          │
   ┌──────┴───────┐
   │  Alertmanager│
   │  (Alerting)  │
   └──────┬───────┘
          │
          ├──────────┬──────────┬──────────────┐
          ▼          ▼          ▼              ▼
      PagerDuty   Slack    Email          Webhook
```

## Components

### 1. Metrics Collection
- **Prometheus**: Time-series database for metrics
- **Node Exporter**: System/hardware metrics
- **Postgres Exporter**: Database metrics
- **Redis Exporter**: Cache metrics
- **Application Metrics**: Custom business metrics

### 2. Distributed Tracing
- **Jaeger**: Distributed tracing system
- **OpenTelemetry**: Instrumentation and trace collection
- Service dependency mapping
- Performance bottleneck identification

### 3. Log Aggregation
- **Loki**: Log aggregation system
- **Promtail**: Log collection and forwarding
- Structured JSON logging with Zap
- Correlation with traces and metrics

### 4. Visualization
- **Grafana**: Unified dashboard platform
- Pre-built dashboards for all service aspects
- Custom business metrics dashboards
- Real-time alerting visualization

### 5. Alerting
- **Alertmanager**: Alert routing and management
- PagerDuty integration for critical alerts
- Slack notifications with smart routing
- Alert inhibition and grouping rules

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.23+ (for local development)
- 8GB+ RAM (recommended for full stack)
- 20GB+ disk space

### 1. Install Dependencies
```bash
# Install Go dependencies
go mod download

# Install Docker & Docker Compose (if not already installed)
# Instructions: https://docs.docker.com/get-docker/
```

### 2. Configure Environment
```bash
# Copy example environment file
cp .env.example .env

# Edit configuration
nano .env

# Required environment variables:
# - SLACK_WEBHOOK_URL: Slack webhook for alerts
# - PAGERDUTY_SERVICE_KEY: PagerDuty integration key
# - GRAFANA_PASSWORD: Grafana admin password
```

### 3. Start Observability Stack
```bash
# Start full stack with observability
docker-compose -f docker/docker-compose.observability.yml up -d

# Wait for all services to be healthy
docker-compose -f docker/docker-compose.observability.yml ps

# Check logs
docker-compose -f docker/docker-compose.observability.yml logs -f
```

### 4. Verify Deployment
```bash
# Check application health
curl http://localhost:8080/health

# Check Prometheus
curl http://localhost:9090/-/healthy

# Check Grafana
curl http://localhost:3000/api/health
```

### 5. Access Dashboards

| Service | URL | Credentials |
|---------|-----|-------------|
| **Grafana** | http://localhost:3000 | admin / ${GRAFANA_PASSWORD} |
| **Prometheus** | http://localhost:9090 | - |
| **Alertmanager** | http://localhost:9093 | - |
| **Jaeger** | http://localhost:16686 | - |
| **Loki** | http://localhost:3100 | - |

### 6. Import Dashboards

Grafana dashboards are automatically provisioned from `observability/grafana/dashboards/`.

Key dashboards:
- **Service Health Overview**: High-level service status
- **Application Performance**: Detailed app metrics
- **Database Performance**: DB and cache metrics
- **Business Metrics**: URLs, clicks, revenue
- **SLO Tracking**: SLO compliance and error budgets
- **Infrastructure**: System resource monitoring

## Configuration

### Prometheus Configuration
Location: `observability/prometheus/prometheus.yml`

Key settings:
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

storage:
  tsdb:
    retention:
      time: 30d
      size: 10GB
```

### Alert Rules
Location: `observability/prometheus/alerts/`

Alert rule files:
- `app_alerts.yml`: Application-level alerts
- `database_alerts.yml`: Database and cache alerts

### Recording Rules
Location: `observability/prometheus/rules/`

Pre-computed metrics for performance:
- `slo_recording_rules.yml`: SLO calculations

### Loki Configuration
Location: `observability/loki/loki-config.yml`

Retention: 31 days (configurable)

### Alertmanager Configuration
Location: `docker/alertmanager.yml`

Alert routing:
- Critical alerts → PagerDuty + Slack
- Warning alerts → Slack only
- Grouped by service, severity, and alertname

## SLO Definitions

### Availability SLO
- **Target**: 99.95%
- **Error Budget**: 0.05% (22 minutes/month downtime)
- **Measurement**: Successful requests / Total requests

### Latency SLO
- **P95 Target**: < 1 second
- **P99 Target**: < 2 seconds
- **Measurement**: Request duration percentiles

### Error Rate SLO
- **Target**: < 0.1%
- **Error Budget**: 0.1% error rate allowed
- **Measurement**: 5xx errors / Total requests

### Error Budget Tracking
Error budgets are tracked with burn rate alerts:
- **Critical**: 14.4x burn rate (budget exhausted in ~2 hours)
- **Warning**: 6x burn rate (budget exhausted in ~5 days)

## Alerting

### Alert Severity Levels

| Severity | Response Time | Examples |
|----------|---------------|----------|
| **P0 - Critical** | < 5 min | Application down, database unavailable |
| **P1 - High** | < 15 min | High error rate, SLO violation |
| **P2 - Medium** | < 1 hour | Cache down, elevated latency |
| **P3 - Low** | < 4 hours | Low cache hit rate, warnings |

### Alert Routing

```
Critical Alerts → PagerDuty (page on-call) + Slack (#incidents)
Warning Alerts → Slack (#alerts-url-shortener)
SLO Alerts → Slack (#slo-alerts)
Business Alerts → Slack (#business-metrics-alerts)
Database Alerts → Slack (#database-alerts)
```

### Configuring Alertmanager

1. **Set up Slack webhook**:
```bash
# Get webhook URL from Slack:
# https://api.slack.com/messaging/webhooks

# Add to .env
echo "SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL" >> .env
```

2. **Set up PagerDuty**:
```bash
# Get integration key from PagerDuty:
# Services → Service Directory → Your Service → Integrations

# Add to .env
echo "PAGERDUTY_SERVICE_KEY=your_service_key" >> .env
```

3. **Reload Alertmanager**:
```bash
docker-compose -f docker/docker-compose.observability.yml restart alertmanager
```

## Runbooks

Incident response runbooks are available in `observability/runbooks/`:

- [Runbook Index](runbooks/RUNBOOK_INDEX.md) - Start here
- [High Latency](runbooks/high-latency.md)
- [High Error Rate](runbooks/high-error-rate.md)

Each runbook includes:
- Symptoms and impact
- Investigation steps (< 5 min)
- Common root causes and solutions
- Mitigation checklists
- Communication templates
- Prevention strategies

## Operational Procedures

### Daily Operations

1. **Morning Check** (5 minutes):
   - Review Service Health dashboard
   - Check SLO compliance
   - Review overnight alerts

2. **Weekly Review** (30 minutes):
   - Analyze error budget consumption
   - Review slow query logs
   - Check capacity trends
   - Review alert noise

3. **Monthly Review** (2 hours):
   - SLO retrospective
   - Capacity planning
   - Runbook updates
   - Dashboard improvements

### Incident Response

1. **Alert Received** (< 1 min):
   - Acknowledge alert
   - Create incident channel
   - Open Service Health dashboard

2. **Triage** (< 2 min):
   - Assess severity
   - Notify stakeholders
   - Assign incident commander (for P0/P1)

3. **Investigate** (< 5 min):
   - Follow runbook
   - Check recent changes
   - Review error logs and traces

4. **Mitigate** (< 10 min for P1):
   - Apply fix or rollback
   - Monitor metrics for improvement

5. **Resolve**:
   - Verify stability (15 min)
   - Close alert
   - Update incident channel

6. **Post-Incident** (< 48 hours):
   - Write blameless postmortem
   - Create action items
   - Update runbooks

### Maintenance Windows

To prevent alerts during maintenance:

```bash
# Create silence in Alertmanager
curl -X POST http://localhost:9093/api/v2/silences \
  -H 'Content-Type: application/json' \
  -d '{
    "matchers": [{"name": "service", "value": "url-shortener", "isRegex": false}],
    "startsAt": "2025-01-08T02:00:00Z",
    "endsAt": "2025-01-08T04:00:00Z",
    "createdBy": "ops-team",
    "comment": "Planned maintenance window"
  }'
```

## Performance Tuning

### Prometheus Optimization

```yaml
# Reduce scrape frequency for non-critical metrics
scrape_interval: 30s  # Default: 15s

# Limit metric retention
storage:
  tsdb:
    retention:
      time: 15d  # Default: 30d
      size: 5GB  # Default: 10GB

# Drop high-cardinality metrics
metric_relabel_configs:
  - source_labels: [__name__]
    regex: 'url_clicks_total'
    action: drop
```

### Loki Optimization

```yaml
# Reduce log retention
limits_config:
  retention_period: 168h  # 7 days instead of 31

# Limit query range
limits_config:
  max_query_length: 168h  # 7 days instead of 30
```

### Jaeger Optimization

```yaml
# Reduce trace retention
environment:
  - SPAN_STORAGE_TYPE=badger
  - BADGER_SPAN_STORE_TTL=72h  # 3 days instead of default

# Implement sampling
environment:
  - SAMPLING_STRATEGIES_FILE=/etc/jaeger/sampling.json

# sampling.json
{
  "default_strategy": {
    "type": "probabilistic",
    "param": 0.1  # Sample 10% of traces
  }
}
```

## Troubleshooting

### Prometheus Not Scraping Targets

```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Check application metrics endpoint
curl http://localhost:8080/metrics

# Check Prometheus logs
docker logs prometheus
```

### Grafana Dashboard Not Loading

```bash
# Check Grafana logs
docker logs grafana

# Verify data source
curl -u admin:${GRAFANA_PASSWORD} http://localhost:3000/api/datasources

# Reimport dashboard
curl -X POST -u admin:${GRAFANA_PASSWORD} \
  http://localhost:3000/api/dashboards/import \
  -H 'Content-Type: application/json' \
  -d @observability/grafana/dashboards/service-health.json
```

### Alerts Not Firing

```bash
# Check alert rules
curl http://localhost:9090/api/v1/rules

# Check Alertmanager status
curl http://localhost:9093/api/v2/status

# Test alert
curl -X POST http://localhost:9093/api/v2/alerts \
  -H 'Content-Type: application/json' \
  -d '[{"labels":{"alertname":"Test","severity":"warning"}}]'
```

### Jaeger Not Receiving Traces

```bash
# Check application trace configuration
env | grep JAEGER

# Check Jaeger collector
curl http://localhost:14269/

# Check application logs for trace errors
docker logs url-shortener-app | grep -i trace
```

## Cost Optimization

### Estimated Monthly Costs

For 1M requests/day:

| Component | Storage | Monthly Cost (AWS) |
|-----------|---------|-------------------|
| Prometheus | 50GB | ~$2.50 |
| Loki | 100GB | ~$5.00 |
| Jaeger | 20GB | ~$1.00 |
| Grafana | Minimal | ~$0.50 |
| **Total** | **170GB** | **~$9/month** |

### Optimization Strategies

1. **Reduce retention periods** (50% cost savings):
   - Metrics: 30d → 15d
   - Logs: 31d → 7d
   - Traces: 7d → 3d

2. **Implement sampling** (70% cost savings):
   - Trace sampling: 10% of requests
   - Drop high-cardinality metrics
   - Aggregate logs before shipping

3. **Use remote storage** (scalability):
   - Cortex for Prometheus metrics
   - S3 for long-term log storage
   - Tempo for trace aggregation

## Security Considerations

1. **Authentication**:
   - Enable Grafana authentication (already configured)
   - Use strong passwords (minimum 16 characters)
   - Implement SSO/SAML for enterprise

2. **Network Security**:
   - Keep observability stack on private network
   - Use VPN or bastion host for access
   - Enable TLS for all communication

3. **Data Retention**:
   - Comply with data retention policies
   - Sanitize logs before aggregation
   - Avoid logging PII/sensitive data

4. **Access Control**:
   - Implement role-based access in Grafana
   - Limit Prometheus/Alertmanager access
   - Audit dashboard access

## Best Practices

1. **Start Simple**: Begin with Service Health dashboard
2. **Monitor What Matters**: Focus on SLIs that impact users
3. **Set Meaningful Alerts**: Avoid alert fatigue
4. **Test Runbooks**: Regular disaster recovery drills
5. **Keep Dashboards Updated**: Version control all configs
6. **Review Regularly**: Weekly metrics review meetings
7. **Document Everything**: Update runbooks after incidents

## Support & Resources

### Documentation
- [Prometheus Docs](https://prometheus.io/docs/)
- [Grafana Docs](https://grafana.com/docs/)
- [Jaeger Docs](https://www.jaegertracing.io/docs/)
- [Loki Docs](https://grafana.com/docs/loki/)
- [OpenTelemetry Docs](https://opentelemetry.io/docs/)

### Internal Resources
- Runbooks: `observability/runbooks/`
- Dashboards: `observability/grafana/dashboards/`
- Alert Rules: `observability/prometheus/alerts/`
- Incident Channel: `#incidents-url-shortener`
- On-Call Schedule: PagerDuty

### Getting Help
1. Check runbooks first
2. Review recent changes (deployments, config)
3. Search logs and traces
4. Escalate to senior engineer if needed

---

**Version**: 1.0.0
**Last Updated**: 2025-01-07
**Maintained By**: DevOps/SRE Team

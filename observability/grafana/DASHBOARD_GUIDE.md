# Grafana Dashboard Guide for URL Shortener

This document describes the Grafana dashboards for comprehensive observability of the URL Shortener service.

## Dashboard Overview

### 1. Service Health Overview Dashboard
**Purpose**: High-level view of service health for executives and on-call engineers
**Refresh**: 30s
**Time Range**: Last 24h

#### Panels:
- **SLO Compliance Gauges** (Top Row)
  - Availability SLO: 99.95% target
  - Latency P95 SLO: < 1s target
  - Error Rate SLO: < 0.1% target
  - Color coding: Green (meeting), Yellow (warning), Red (violation)

- **Error Budget Remaining** (Top Row)
  - 30-day error budget tracking
  - Burn rate visualization
  - Projected exhaustion date

- **Request Rate** (Graph)
  - Query: `sum(rate(http_requests_total{service="url-shortener"}[5m]))`
  - Shows requests per second over time

- **Error Rate %** (Graph)
  - Query: `(sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))) * 100`
  - Alert threshold line at 1%

- **Latency Percentiles** (Graph)
  - P50, P95, P99 latency over time
  - Alert threshold lines at 1s (P95) and 2s (P99)

- **Active Incidents** (Stat)
  - Current firing alerts
  - Link to Alertmanager

- **Service Status** (Stat Panels)
  - Application: UP/DOWN
  - Database: UP/DOWN
  - Cache: UP/DOWN
  - Color-coded status indicators

---

### 2. Application Performance Dashboard
**Purpose**: Detailed application performance metrics
**Refresh**: 15s
**Time Range**: Last 1h

#### Panels:
- **HTTP Request Rate by Endpoint** (Graph)
  - Query: `sum(rate(http_requests_total[5m])) by (endpoint, method)`
  - Stacked area chart

- **Request Duration Heatmap** (Heatmap)
  - Query: `sum(increase(http_request_duration_seconds_bucket[1m])) by (le)`
  - Visualize latency distribution

- **Top 10 Slowest Endpoints** (Bar Chart)
  - Query: `topk(10, histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, endpoint)))`

- **HTTP Status Code Distribution** (Pie Chart)
  - Query: `sum(rate(http_requests_total[5m])) by (status)`

- **Request Size Distribution** (Graph)
  - Query: `histogram_quantile(0.95, sum(rate(http_request_size_bytes_bucket[5m])) by (le))`

- **Response Size Distribution** (Graph)
  - Query: `histogram_quantile(0.95, sum(rate(http_response_size_bytes_bucket[5m])) by (le))`

- **Goroutine Count** (Graph)
  - Query: `goroutines_active`
  - Monitor for goroutine leaks

- **Memory Usage** (Graph)
  - Query: `memory_allocated_bytes`
  - Monitor memory allocation patterns

---

### 3. Database Performance Dashboard
**Purpose**: PostgreSQL and cache performance monitoring
**Refresh**: 30s
**Time Range**: Last 1h

#### Panels:
- **Database Connection Pool** (Graph)
  - Active connections: `db_connections_active`
  - Idle connections: `db_connections_idle`
  - Stacked area chart

- **Query Rate by Operation** (Graph)
  - Query: `sum(rate(db_queries_total[5m])) by (operation, table)`

- **Query Latency P95/P99** (Graph)
  - Query: `histogram_quantile(0.95, sum(rate(db_query_duration_seconds_bucket[5m])) by (le, operation))`

- **Slow Query Rate** (Stat)
  - Query: `sum(rate(db_slow_queries_total[5m]))`
  - Alert threshold

- **Database Error Rate** (Graph)
  - Query: `sum(rate(db_queries_total{status="error"}[5m])) / sum(rate(db_queries_total[5m]))`

- **Cache Hit Ratio** (Gauge)
  - Query: `sum(rate(cache_hits_total[5m])) / (sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m])))`
  - Target: > 85%

- **Cache Operations/sec** (Graph)
  - Hits: `sum(rate(cache_hits_total[5m]))`
  - Misses: `sum(rate(cache_misses_total[5m]))`

- **Redis Memory Usage** (Graph)
  - Query: `redis_memory_used_bytes / redis_memory_max_bytes * 100`

- **Redis Key Eviction Rate** (Graph)
  - Query: `rate(redis_evicted_keys_total[5m])`

---

### 4. Business Metrics Dashboard
**Purpose**: Business KPIs and product metrics
**Refresh**: 1m
**Time Range**: Last 24h

#### Panels:
- **URLs Created** (Stat)
  - Total URLs created today
  - Query: `sum(increase(urls_created_total[24h]))`

- **URL Creation Rate** (Graph)
  - Query: `sum(rate(urls_created_total[5m]))`
  - Show hourly patterns

- **Total Clicks/Redirects** (Stat)
  - Query: `sum(increase(http_requests_total{endpoint="/redirect"}[24h]))`

- **Click Rate** (Graph)
  - Query: `sum(rate(http_requests_total{endpoint="/redirect"}[5m]))`

- **Custom vs Auto-Generated URLs** (Pie Chart)
  - Query: `sum(increase(urls_created_total[24h])) by (custom_code)`

- **URLs with Expiry** (Pie Chart)
  - Query: `sum(increase(urls_created_total[24h])) by (with_expiry)`

- **Active Users** (Stat)
  - Query: `active_users_total`

- **Revenue by Plan Type** (Graph)
  - Query: `sum(rate(revenue_total_cents[1h])) by (plan_type) / 100`

- **Active Subscriptions** (Stat Panel Grid)
  - Free tier: `subscriptions_active{plan_type="free"}`
  - Pro tier: `subscriptions_active{plan_type="pro"}`
  - Enterprise: `subscriptions_active{plan_type="enterprise"}`

---

### 5. SLO Tracking Dashboard
**Purpose**: Detailed SLO compliance and error budget tracking
**Refresh**: 1m
**Time Range**: Last 30d

#### Panels:
- **30-Day Availability SLO** (Gauge)
  - Query: `sli:availability:ratio_rate30d * 100`
  - Target line at 99.95%

- **30-Day Latency SLO** (Gauge)
  - Query: `sli:latency:ratio_rate30d * 100`
  - Target: 95% of requests < 1s

- **30-Day Error Rate SLO** (Gauge)
  - Query: `(1 - sli:error_rate:ratio_rate30d) * 100`
  - Target: < 0.1%

- **Error Budget Remaining** (Graph)
  - Availability: `slo:error_budget:remaining{slo="availability"}`
  - Latency: `slo:error_budget:remaining{slo="latency"}`
  - Error Rate: `slo:error_budget:remaining{slo="error_rate"}`

- **Error Budget Burn Rate** (Graph - Multi-Window)
  - 1h window: `slo:error_budget:burn_rate_1h`
  - 6h window: `slo:error_budget:burn_rate_6h`
  - 24h window: `slo:error_budget:burn_rate_24h`
  - Alert threshold line at 2.0

- **SLI Time Series** (Graph)
  - Show 5m, 1h, 6h, 24h windows
  - Compare against SLO targets

- **Projected Error Budget Exhaustion** (Stat)
  - Calculation based on current burn rate
  - Alert if < 7 days

- **SLO Violation Events** (Table)
  - Query alerts where SLO was violated
  - Timestamp, duration, impact

---

### 6. Infrastructure Monitoring Dashboard
**Purpose**: System resource monitoring
**Refresh**: 30s
**Time Range**: Last 1h

#### Panels:
- **CPU Usage %** (Graph)
  - Query: `100 - (avg(irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
  - Alert thresholds at 80% and 90%

- **Memory Usage %** (Graph)
  - Query: `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`

- **Disk Usage %** (Gauge Panel Grid)
  - Per mount point
  - Alert threshold at 80%

- **Network I/O** (Graph)
  - Receive: `rate(node_network_receive_bytes_total[5m])`
  - Transmit: `rate(node_network_transmit_bytes_total[5m])`

- **Disk I/O** (Graph)
  - Read: `rate(node_disk_read_bytes_total[5m])`
  - Write: `rate(node_disk_written_bytes_total[5m])`

- **Load Average** (Graph)
  - 1m, 5m, 15m load averages

- **Open File Descriptors** (Graph)
  - Monitor for resource exhaustion

---

### 7. Distributed Tracing Dashboard
**Purpose**: Jaeger trace visualization
**Integration**: Links to Jaeger UI

#### Panels:
- **Trace Count by Service** (Graph)
- **Average Trace Duration** (Graph)
- **Error Trace Rate** (Graph)
- **Service Dependency Map** (Link to Jaeger)
- **Recent Slow Traces** (Table with links to Jaeger)

---

## Dashboard Access URLs

After deployment:
- Service Health: `http://grafana:3000/d/service-health`
- App Performance: `http://grafana:3000/d/app-performance`
- Database Performance: `http://grafana:3000/d/db-performance`
- Business Metrics: `http://grafana:3000/d/business-metrics`
- SLO Tracking: `http://grafana:3000/d/slo-tracking`
- Infrastructure: `http://grafana:3000/d/infrastructure`

## Alert Integration

Dashboards include:
- **Alert Annotations**: Show when alerts fired on graphs
- **Alert Status**: Real-time alert state in panels
- **Links to Runbooks**: Click-through to incident response procedures
- **Links to Traces**: Deep-link to relevant Jaeger traces

## Custom Variables

All dashboards support variables:
- `$datasource`: Prometheus data source selector
- `$environment`: production/staging/development
- `$service`: url-shortener (default)
- `$interval`: Auto-refresh interval
- `$timeRange`: Time range selector

## Best Practices

1. **Start with Service Health** dashboard for quick health check
2. **Drill down** to specific dashboards based on issues identified
3. **Use time range** controls to compare current vs historical data
4. **Enable auto-refresh** during incidents
5. **Share dashboard snapshots** in incident channels
6. **Export dashboards** regularly for version control

## Creating Dashboard JSON Files

To generate actual dashboard JSON files, use Grafana's UI:
1. Import these panel configurations
2. Save the dashboard
3. Export as JSON
4. Place in `observability/grafana/dashboards/`

Or use the Grafana API for programmatic generation.

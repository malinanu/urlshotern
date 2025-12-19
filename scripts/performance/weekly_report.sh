#!/bin/bash

# Weekly Performance Report Generator
# This script generates a weekly performance report and sends it via email/Slack

set -e

# Configuration
PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9090}"
LOOKBACK_HOURS=168  # 7 days
REPORT_DATE=$(date +%Y-%m-%d)
REPORT_DIR="./reports"
REPORT_FILE="$REPORT_DIR/weekly_performance_report_$REPORT_DATE.html"

# Ensure report directory exists
mkdir -p "$REPORT_DIR"

echo "Generating Weekly Performance Report for $REPORT_DATE"
echo "=================================================="

# Collect baseline metrics for the week
echo "Collecting performance metrics..."
cd tools/performance
go run baseline_collector.go \
    --prometheus-url="$PROMETHEUS_URL" \
    --lookback=$LOOKBACK_HOURS \
    --output="../../$REPORT_DIR/weekly_baseline_$REPORT_DATE.json"
cd ../..

# Generate HTML report
echo "Generating HTML report..."
cat > "$REPORT_FILE" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Weekly Performance Report</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .header {
            background-color: #2c3e50;
            color: white;
            padding: 20px;
            border-radius: 5px;
            margin-bottom: 20px;
        }
        .section {
            background-color: white;
            padding: 20px;
            margin-bottom: 20px;
            border-radius: 5px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .metric-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 15px;
            margin-top: 15px;
        }
        .metric-card {
            background-color: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            border-left: 4px solid #3498db;
        }
        .metric-card.warning {
            border-left-color: #f39c12;
        }
        .metric-card.critical {
            border-left-color: #e74c3c;
        }
        .metric-card.success {
            border-left-color: #2ecc71;
        }
        .metric-title {
            font-size: 14px;
            color: #7f8c8d;
            margin-bottom: 5px;
        }
        .metric-value {
            font-size: 28px;
            font-weight: bold;
            color: #2c3e50;
        }
        .metric-change {
            font-size: 14px;
            margin-top: 5px;
        }
        .metric-change.positive {
            color: #2ecc71;
        }
        .metric-change.negative {
            color: #e74c3c;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 15px;
        }
        th, td {
            text-align: left;
            padding: 12px;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #34495e;
            color: white;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
        .status-badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: bold;
        }
        .status-pass {
            background-color: #d4edda;
            color: #155724;
        }
        .status-warning {
            background-color: #fff3cd;
            color: #856404;
        }
        .status-fail {
            background-color: #f8d7da;
            color: #721c24;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Weekly Performance Report</h1>
        <p>Week of: <strong>{{REPORT_DATE}}</strong></p>
        <p>Report Generated: {{GENERATION_TIME}}</p>
    </div>

    <div class="section">
        <h2>Executive Summary</h2>
        <p>Overall system performance status: <span class="status-badge status-{{OVERALL_STATUS}}">{{OVERALL_STATUS}}</span></p>
        <div class="metric-grid">
            <div class="metric-card {{REDIRECT_STATUS}}">
                <div class="metric-title">URL Redirect Latency (p95)</div>
                <div class="metric-value">{{REDIRECT_P95}}ms</div>
                <div class="metric-change {{REDIRECT_TREND}}">{{REDIRECT_CHANGE}} vs. last week</div>
            </div>
            <div class="metric-card {{SHORTEN_STATUS}}">
                <div class="metric-title">URL Shortening Latency (p95)</div>
                <div class="metric-value">{{SHORTEN_P95}}ms</div>
                <div class="metric-change {{SHORTEN_TREND}}">{{SHORTEN_CHANGE}} vs. last week</div>
            </div>
            <div class="metric-card {{CACHE_STATUS}}">
                <div class="metric-title">Cache Hit Rate</div>
                <div class="metric-value">{{CACHE_HIT_RATE}}%</div>
                <div class="metric-change {{CACHE_TREND}}">{{CACHE_CHANGE}} vs. last week</div>
            </div>
            <div class="metric-card {{ERROR_STATUS}}">
                <div class="metric-title">Error Rate</div>
                <div class="metric-value">{{ERROR_RATE}}%</div>
                <div class="metric-change {{ERROR_TREND}}">{{ERROR_CHANGE}} vs. last week</div>
            </div>
        </div>
    </div>

    <div class="section">
        <h2>Performance Budget Compliance</h2>
        <table>
            <thead>
                <tr>
                    <th>Endpoint</th>
                    <th>Budget (p95)</th>
                    <th>Actual (p95)</th>
                    <th>Margin</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td>GET /{shortCode}</td>
                    <td>50ms</td>
                    <td>{{REDIRECT_P95}}ms</td>
                    <td>{{REDIRECT_MARGIN}}%</td>
                    <td><span class="status-badge status-{{REDIRECT_BUDGET_STATUS}}">{{REDIRECT_BUDGET_STATUS}}</span></td>
                </tr>
                <tr>
                    <td>POST /api/v1/shorten</td>
                    <td>200ms</td>
                    <td>{{SHORTEN_P95}}ms</td>
                    <td>{{SHORTEN_MARGIN}}%</td>
                    <td><span class="status-badge status-{{SHORTEN_BUDGET_STATUS}}">{{SHORTEN_BUDGET_STATUS}}</span></td>
                </tr>
                <tr>
                    <td>GET /api/v1/analytics</td>
                    <td>500ms</td>
                    <td>{{ANALYTICS_P95}}ms</td>
                    <td>{{ANALYTICS_MARGIN}}%</td>
                    <td><span class="status-badge status-{{ANALYTICS_BUDGET_STATUS}}">{{ANALYTICS_BUDGET_STATUS}}</span></td>
                </tr>
            </tbody>
        </table>
    </div>

    <div class="section">
        <h2>Alerts and Incidents</h2>
        <p><strong>Total Alerts:</strong> {{TOTAL_ALERTS}}</p>
        <p><strong>Incidents:</strong> {{TOTAL_INCIDENTS}}</p>
        <table>
            <thead>
                <tr>
                    <th>Date</th>
                    <th>Type</th>
                    <th>Severity</th>
                    <th>Duration</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                {{INCIDENTS_TABLE_ROWS}}
            </tbody>
        </table>
    </div>

    <div class="section">
        <h2>Infrastructure Metrics</h2>
        <div class="metric-grid">
            <div class="metric-card">
                <div class="metric-title">CPU Utilization</div>
                <div class="metric-value">{{CPU_UTIL}}%</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">Memory Usage</div>
                <div class="metric-value">{{MEMORY_UTIL}}%</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">Active Connections</div>
                <div class="metric-value">{{ACTIVE_CONNS}}</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">Goroutines</div>
                <div class="metric-value">{{GOROUTINES}}</div>
            </div>
        </div>
    </div>

    <div class="section">
        <h2>Capacity Status</h2>
        <table>
            <thead>
                <tr>
                    <th>Resource</th>
                    <th>Current</th>
                    <th>Threshold</th>
                    <th>Status</th>
                    <th>Action Required</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td>CPU</td>
                    <td>{{CPU_UTIL}}%</td>
                    <td>70%</td>
                    <td><span class="status-badge status-{{CPU_STATUS}}">{{CPU_STATUS}}</span></td>
                    <td>{{CPU_ACTION}}</td>
                </tr>
                <tr>
                    <td>Memory</td>
                    <td>{{MEMORY_UTIL}}%</td>
                    <td>80%</td>
                    <td><span class="status-badge status-{{MEMORY_STATUS}}">{{MEMORY_STATUS}}</span></td>
                    <td>{{MEMORY_ACTION}}</td>
                </tr>
            </tbody>
        </table>
    </div>

    <div class="section">
        <h2>Week-over-Week Comparison</h2>
        <table>
            <thead>
                <tr>
                    <th>Metric</th>
                    <th>Last Week</th>
                    <th>This Week</th>
                    <th>Change</th>
                    <th>Trend</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td>Avg Requests/sec</td>
                    <td>{{PREV_RPS}}</td>
                    <td>{{CURR_RPS}}</td>
                    <td>{{RPS_CHANGE}}</td>
                    <td>{{RPS_TREND}}</td>
                </tr>
                <tr>
                    <td>Redirect p95</td>
                    <td>{{PREV_REDIRECT_P95}}ms</td>
                    <td>{{REDIRECT_P95}}ms</td>
                    <td>{{REDIRECT_CHANGE}}</td>
                    <td>{{REDIRECT_TREND}}</td>
                </tr>
                <tr>
                    <td>Cache Hit Rate</td>
                    <td>{{PREV_CACHE_HIT}}%</td>
                    <td>{{CACHE_HIT_RATE}}%</td>
                    <td>{{CACHE_CHANGE}}</td>
                    <td>{{CACHE_TREND}}</td>
                </tr>
            </tbody>
        </table>
    </div>

    <div class="section">
        <h2>Action Items</h2>
        <ul>
            {{ACTION_ITEMS}}
        </ul>
    </div>

    <div class="section">
        <h2>Next Week's Focus</h2>
        <ul>
            <li>Continue monitoring performance budget compliance</li>
            <li>Review and prioritize optimization backlog</li>
            <li>Monitor capacity planning forecasts</li>
        </ul>
    </div>

    <p style="text-align: center; color: #7f8c8d; margin-top: 40px;">
        Generated by Performance Monitoring System<br>
        For questions, contact the Performance Engineering Team
    </p>
</body>
</html>
EOF

# TODO: Replace placeholders with actual data from collected metrics
# For now, this is a template. In production, parse the JSON baseline
# and populate the template with actual values.

echo "Report template generated: $REPORT_FILE"

# Send report via email (if configured)
if [ -n "$EMAIL_RECIPIENTS" ]; then
    echo "Sending report via email to: $EMAIL_RECIPIENTS"
    # TODO: Implement email sending
    # Example: cat "$REPORT_FILE" | mail -s "Weekly Performance Report - $REPORT_DATE" "$EMAIL_RECIPIENTS"
fi

# Send notification to Slack (if configured)
if [ -n "$SLACK_WEBHOOK_URL" ]; then
    echo "Sending notification to Slack..."
    curl -X POST "$SLACK_WEBHOOK_URL" \
        -H 'Content-Type: application/json' \
        -d "{
            \"text\": \"Weekly Performance Report for $REPORT_DATE is ready\",
            \"attachments\": [{
                \"color\": \"good\",
                \"fields\": [
                    {\"title\": \"Report Date\", \"value\": \"$REPORT_DATE\", \"short\": true},
                    {\"title\": \"Status\", \"value\": \"Available\", \"short\": true}
                ],
                \"footer\": \"Performance Monitoring System\"
            }]
        }"
fi

echo "Weekly performance report generation complete!"

#!/usr/bin/env python3
"""
Performance Dashboard Generator

Generates HTML dashboards with performance trends, regression alerts,
and historical tracking for CI/CD integration.
"""

import json
import os
import sys
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Any
import statistics


class PerformanceDashboard:
    def __init__(self, trends_dir: str = 'performance-trends'):
        self.trends_dir = Path(trends_dir)
        self.trends_dir.mkdir(exist_ok=True)
        self.history_file = self.trends_dir / 'history.jsonl'

    def load_history(self, days: int = 30) -> List[Dict]:
        """Load performance history for the specified number of days."""
        if not self.history_file.exists():
            return []

        cutoff_date = datetime.now() - timedelta(days=days)
        history = []

        with open(self.history_file, 'r') as f:
            for line in f:
                entry = json.loads(line.strip())
                entry_date = datetime.fromisoformat(entry.get('timestamp', ''))

                if entry_date >= cutoff_date:
                    history.append(entry)

        return history

    def save_result(self, result: Dict):
        """Save a performance test result to history."""
        result['timestamp'] = datetime.now().isoformat()

        with open(self.history_file, 'a') as f:
            f.write(json.dumps(result) + '\n')

    def detect_regressions(self, current: Dict, baseline: Dict,
                          threshold: float = 0.10) -> List[Dict]:
        """Detect performance regressions between current and baseline."""
        regressions = []

        def compare_metrics(current_metrics: Dict, baseline_metrics: Dict,
                           prefix: str = ''):
            for key, current_value in current_metrics.items():
                baseline_value = baseline_metrics.get(key)

                if baseline_value is None:
                    continue

                # Handle nested dictionaries
                if isinstance(current_value, dict) and isinstance(baseline_value, dict):
                    compare_metrics(current_value, baseline_value,
                                   f"{prefix}{key}.")
                    continue

                # Compare numeric values
                if isinstance(current_value, (int, float)) and \
                   isinstance(baseline_value, (int, float)):
                    if baseline_value > 0:
                        change = (current_value - baseline_value) / baseline_value

                        if change > threshold:
                            regressions.append({
                                'metric': f"{prefix}{key}",
                                'current': current_value,
                                'baseline': baseline_value,
                                'change_percentage': change * 100,
                                'severity': 'critical' if change > 0.25 else 'warning'
                            })

        # Compare K6 metrics
        if 'k6' in current and 'k6' in baseline:
            compare_metrics(current['k6'], baseline['k6'], 'k6.')

        # Compare Lighthouse metrics
        if 'lighthouse' in current and 'lighthouse' in baseline:
            compare_metrics(current['lighthouse'], baseline['lighthouse'], 'lighthouse.')

        # Compare database metrics
        if 'database' in current and 'database' in baseline:
            compare_metrics(current['database'], baseline['database'], 'database.')

        return regressions

    def calculate_statistics(self, history: List[Dict],
                            metric_path: str) -> Dict[str, float]:
        """Calculate statistics for a specific metric over time."""
        values = []

        for entry in history:
            value = self._get_nested_value(entry, metric_path)
            if value is not None and isinstance(value, (int, float)):
                values.append(value)

        if not values:
            return {}

        return {
            'mean': statistics.mean(values),
            'median': statistics.median(values),
            'stdev': statistics.stdev(values) if len(values) > 1 else 0,
            'min': min(values),
            'max': max(values),
            'p95': self._percentile(values, 0.95),
            'p99': self._percentile(values, 0.99),
        }

    def _get_nested_value(self, data: Dict, path: str) -> Any:
        """Get a nested value from a dictionary using dot notation."""
        keys = path.split('.')
        value = data

        for key in keys:
            if isinstance(value, dict):
                value = value.get(key)
            else:
                return None

        return value

    def _percentile(self, values: List[float], p: float) -> float:
        """Calculate percentile of a list of values."""
        sorted_values = sorted(values)
        index = int(len(sorted_values) * p)
        return sorted_values[min(index, len(sorted_values) - 1)]

    def generate_html_report(self, history: List[Dict],
                            regressions: List[Dict],
                            output_file: str = 'performance-report.html'):
        """Generate an HTML performance report."""
        html = f"""
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Performance Dashboard - URL Shortener</title>
    <style>
        * {{
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }}

        body {{
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: #f5f7fa;
            color: #333;
            line-height: 1.6;
        }}

        .container {{
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }}

        header {{
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 30px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }}

        h1 {{
            font-size: 2.5em;
            margin-bottom: 10px;
        }}

        .timestamp {{
            opacity: 0.9;
            font-size: 0.9em;
        }}

        .grid {{
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }}

        .card {{
            background: white;
            padding: 25px;
            border-radius: 10px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }}

        .card h2 {{
            font-size: 1.5em;
            margin-bottom: 15px;
            color: #667eea;
        }}

        .metric {{
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 10px 0;
            border-bottom: 1px solid #eee;
        }}

        .metric:last-child {{
            border-bottom: none;
        }}

        .metric-name {{
            font-weight: 500;
        }}

        .metric-value {{
            font-size: 1.2em;
            font-weight: 600;
        }}

        .status {{
            display: inline-block;
            padding: 5px 15px;
            border-radius: 20px;
            font-size: 0.85em;
            font-weight: 600;
        }}

        .status.pass {{
            background: #d4edda;
            color: #155724;
        }}

        .status.warning {{
            background: #fff3cd;
            color: #856404;
        }}

        .status.critical {{
            background: #f8d7da;
            color: #721c24;
        }}

        .regression {{
            background: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 15px;
            margin-bottom: 15px;
            border-radius: 5px;
        }}

        .regression.critical {{
            background: #f8d7da;
            border-left-color: #dc3545;
        }}

        .regression-metric {{
            font-weight: 600;
            font-size: 1.1em;
            margin-bottom: 5px;
        }}

        .regression-details {{
            font-size: 0.9em;
            color: #666;
        }}

        .chart {{
            margin-top: 20px;
            height: 300px;
            background: #f8f9fa;
            border-radius: 5px;
            padding: 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #999;
        }}

        table {{
            width: 100%;
            border-collapse: collapse;
            margin-top: 15px;
        }}

        th, td {{
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #eee;
        }}

        th {{
            background: #f8f9fa;
            font-weight: 600;
        }}

        .trend {{
            display: inline-flex;
            align-items: center;
            font-size: 0.9em;
        }}

        .trend.up {{
            color: #dc3545;
        }}

        .trend.down {{
            color: #28a745;
        }}

        .trend.stable {{
            color: #6c757d;
        }}

        .footer {{
            text-align: center;
            padding: 20px;
            color: #666;
            font-size: 0.9em;
        }}
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🚀 Performance Dashboard</h1>
            <p class="timestamp">Generated on {datetime.now().strftime('%Y-%m-%d %H:%M:%S UTC')}</p>
            <p>URL Shortener Application - Performance Monitoring</p>
        </header>

        <!-- Performance Status Overview -->
        <div class="grid">
            <div class="card">
                <h2>Overall Status</h2>
                <div class="metric">
                    <span class="metric-name">Health</span>
                    <span class="status {'critical' if regressions else 'pass'}">
                        {'REGRESSION DETECTED' if regressions else 'HEALTHY'}
                    </span>
                </div>
                <div class="metric">
                    <span class="metric-name">Regressions</span>
                    <span class="metric-value">{len(regressions)}</span>
                </div>
                <div class="metric">
                    <span class="metric-name">Test Runs</span>
                    <span class="metric-value">{len(history)}</span>
                </div>
            </div>

            <div class="card">
                <h2>API Performance</h2>
                <div class="metric">
                    <span class="metric-name">P95 Response Time</span>
                    <span class="metric-value">-</span>
                </div>
                <div class="metric">
                    <span class="metric-name">Error Rate</span>
                    <span class="metric-value">-</span>
                </div>
                <div class="metric">
                    <span class="metric-name">Throughput</span>
                    <span class="metric-value">-</span>
                </div>
            </div>

            <div class="card">
                <h2>Frontend Performance</h2>
                <div class="metric">
                    <span class="metric-name">LCP</span>
                    <span class="metric-value">-</span>
                </div>
                <div class="metric">
                    <span class="metric-name">Performance Score</span>
                    <span class="metric-value">-</span>
                </div>
                <div class="metric">
                    <span class="metric-name">Bundle Size</span>
                    <span class="metric-value">-</span>
                </div>
            </div>
        </div>

        <!-- Regressions -->
        {self._generate_regressions_html(regressions)}

        <!-- Performance Trends -->
        <div class="card">
            <h2>Performance Trends (Last 30 Days)</h2>
            <div class="chart">
                Chart visualization would go here
                <br>(Integrate with Chart.js or similar library)
            </div>
        </div>

        <!-- Historical Data -->
        <div class="card">
            <h2>Historical Performance Data</h2>
            <table>
                <thead>
                    <tr>
                        <th>Date</th>
                        <th>Commit</th>
                        <th>P95 API</th>
                        <th>LCP</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    {self._generate_history_table_html(history)}
                </tbody>
            </table>
        </div>

        <footer class="footer">
            <p>Automated Performance Testing Dashboard | Generated by CI/CD Pipeline</p>
        </footer>
    </div>
</body>
</html>
        """

        output_path = self.trends_dir / output_file
        with open(output_path, 'w') as f:
            f.write(html)

        print(f"HTML report generated: {output_path}")
        return output_path

    def _generate_regressions_html(self, regressions: List[Dict]) -> str:
        if not regressions:
            return """
            <div class="card">
                <h2>✅ No Regressions Detected</h2>
                <p>All performance metrics are within acceptable thresholds.</p>
            </div>
            """

        html = '<div class="card"><h2>⚠️ Performance Regressions</h2>'

        for regression in regressions:
            severity_class = regression['severity']
            html += f"""
            <div class="regression {severity_class}">
                <div class="regression-metric">{regression['metric']}</div>
                <div class="regression-details">
                    Current: {regression['current']:.2f} |
                    Baseline: {regression['baseline']:.2f} |
                    Change: <strong>{regression['change_percentage']:+.2f}%</strong>
                </div>
            </div>
            """

        html += '</div>'
        return html

    def _generate_history_table_html(self, history: List[Dict]) -> str:
        if not history:
            return '<tr><td colspan="5">No historical data available</td></tr>'

        html = ''
        for entry in sorted(history, key=lambda x: x.get('timestamp', ''), reverse=True)[:10]:
            timestamp = entry.get('timestamp', 'N/A')
            commit = entry.get('commit', 'N/A')[:7]
            status = 'pass' if not entry.get('regressions') else 'warning'

            html += f"""
            <tr>
                <td>{timestamp[:10]}</td>
                <td><code>{commit}</code></td>
                <td>-</td>
                <td>-</td>
                <td><span class="status {status}">{status.upper()}</span></td>
            </tr>
            """

        return html

    def generate_summary(self, history: List[Dict]) -> Dict:
        """Generate a performance summary."""
        if not history:
            return {'error': 'No historical data available'}

        latest = history[-1] if history else {}

        return {
            'total_runs': len(history),
            'latest_run': latest.get('timestamp'),
            'commit': latest.get('commit'),
            'metrics': latest,
        }


def main():
    """Main entry point for the dashboard generator."""
    dashboard = PerformanceDashboard()

    # Load history
    history = dashboard.load_history(days=30)

    # Get current and baseline results
    current = history[-1] if history else {}
    baseline = history[-2] if len(history) > 1 else {}

    # Detect regressions
    regressions = dashboard.detect_regressions(current, baseline) if baseline else []

    # Generate HTML report
    dashboard.generate_html_report(history, regressions)

    # Print summary
    summary = dashboard.generate_summary(history)
    print(json.dumps(summary, indent=2))

    # Exit with error code if regressions detected
    if regressions:
        print(f"\n⚠️  {len(regressions)} regression(s) detected!")
        sys.exit(1)
    else:
        print("\n✅ No regressions detected")
        sys.exit(0)


if __name__ == '__main__':
    main()

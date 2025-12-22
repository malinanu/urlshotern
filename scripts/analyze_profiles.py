#!/usr/bin/env python3
"""
Performance Profile Analysis Script
Analyzes pprof data and generates performance reports
"""

import json
import sys
from pathlib import Path
from datetime import datetime


def analyze_metrics(baseline_file, load_file, output_file):
    """Compare baseline and under-load metrics"""

    print("Analyzing metrics...")

    # Load JSON files
    with open(baseline_file, 'r') as f:
        baseline = json.load(f)

    with open(load_file, 'r') as f:
        load = json.load(f)

    # Generate analysis report
    report = []
    report.append("=" * 80)
    report.append("PERFORMANCE METRICS ANALYSIS")
    report.append("=" * 80)
    report.append(f"\nGenerated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")

    # Runtime metrics comparison
    report.append("\n1. RUNTIME METRICS COMPARISON")
    report.append("-" * 80)

    baseline_runtime = baseline.get('runtime', {})
    load_runtime = load.get('runtime', {})

    metrics = [
        ('Goroutines', 'goroutines', ''),
        ('Heap Alloc', 'heap_alloc_mb', 'MB'),
        ('Heap Sys', 'heap_sys_mb', 'MB'),
        ('Heap Objects', 'heap_objects', ''),
        ('GC Runs', 'gc_runs', ''),
    ]

    for label, key, unit in metrics:
        baseline_val = baseline_runtime.get(key, 0)
        load_val = load_runtime.get(key, 0)
        diff = load_val - baseline_val
        if baseline_val > 0:
            pct_change = (diff / baseline_val) * 100
        else:
            pct_change = 0

        report.append(f"{label:20s}: {baseline_val:>12.2f} {unit:>3s} → {load_val:>12.2f} {unit:>3s} "
                     f"(Δ {diff:>+10.2f}, {pct_change:>+6.1f}%)")

    # Cache metrics
    report.append("\n2. CACHE PERFORMANCE")
    report.append("-" * 80)

    baseline_cache_rate = baseline.get('cache_hit_rate', 0)
    load_cache_rate = load.get('cache_hit_rate', 0)

    report.append(f"Baseline Cache Hit Rate: {baseline_cache_rate:.2f}%")
    report.append(f"Under Load Cache Hit Rate: {load_cache_rate:.2f}%")
    report.append(f"Change: {load_cache_rate - baseline_cache_rate:+.2f}%")

    baseline_hits = baseline.get('cache_hits', 0)
    baseline_misses = baseline.get('cache_misses', 0)
    load_hits = load.get('cache_hits', 0)
    load_misses = load.get('cache_misses', 0)

    report.append(f"\nCache Hits: {baseline_hits} → {load_hits} (Δ {load_hits - baseline_hits:+d})")
    report.append(f"Cache Misses: {baseline_misses} → {load_misses} (Δ {load_misses - baseline_misses:+d})")

    # Database metrics
    report.append("\n3. DATABASE PERFORMANCE")
    report.append("-" * 80)

    baseline_queries = baseline.get('database_queries', 0)
    load_queries = load.get('database_queries', 0)
    baseline_slow = baseline.get('slow_queries', 0)
    load_slow = load.get('slow_queries', 0)

    report.append(f"Total Queries: {baseline_queries} → {load_queries} (Δ {load_queries - baseline_queries:+d})")
    report.append(f"Slow Queries (>10ms): {baseline_slow} → {load_slow} (Δ {load_slow - baseline_slow:+d})")

    if load_queries > 0:
        slow_pct = (load_slow / load_queries) * 100
        report.append(f"Slow Query Rate: {slow_pct:.2f}%")

    # Endpoint metrics
    report.append("\n4. ENDPOINT PERFORMANCE")
    report.append("-" * 80)

    baseline_endpoints = baseline.get('endpoints', {})
    load_endpoints = load.get('endpoints', {})

    # Find common endpoints
    common_endpoints = set(baseline_endpoints.keys()) & set(load_endpoints.keys())

    if common_endpoints:
        report.append(f"\n{'Endpoint':<40s} {'P95 (ms)':<12s} {'Error Rate':<12s} {'RPS':<12s}")
        report.append("-" * 80)

        for endpoint in sorted(common_endpoints):
            load_stats = load_endpoints[endpoint]
            p95 = load_stats.get('p95_ms', 0)
            error_rate = load_stats.get('error_rate', 0)
            count = load_stats.get('count', 0)

            # Estimate RPS (assuming metrics collected over load test duration)
            rps = count / 60  # Assuming 60-second collection window

            report.append(f"{endpoint:<40s} {p95:>10.2f} ms {error_rate:>10.2f}% {rps:>10.1f}")

    # Performance issues
    report.append("\n5. IDENTIFIED PERFORMANCE ISSUES")
    report.append("-" * 80)

    issues = []

    # Check goroutine growth
    goroutine_growth = load_runtime.get('goroutines', 0) - baseline_runtime.get('goroutines', 0)
    if goroutine_growth > 1000:
        issues.append(f"⚠ CRITICAL: Excessive goroutine growth (+{goroutine_growth})")
        issues.append("  → Possible goroutine leak or unbounded spawning")

    # Check memory growth
    heap_growth = load_runtime.get('heap_alloc_mb', 0) - baseline_runtime.get('heap_alloc_mb', 0)
    if heap_growth > 50:
        issues.append(f"⚠ WARNING: High memory growth (+{heap_growth:.2f} MB)")
        issues.append("  → Check for memory leaks or excessive allocations")

    # Check cache hit rate
    if load_cache_rate < 70:
        issues.append(f"⚠ WARNING: Low cache hit rate ({load_cache_rate:.2f}%)")
        issues.append("  → Consider increasing cache TTL or cache size")

    # Check slow queries
    if load_slow > 0 and load_queries > 0:
        slow_rate = (load_slow / load_queries) * 100
        if slow_rate > 10:
            issues.append(f"⚠ WARNING: High slow query rate ({slow_rate:.2f}%)")
            issues.append("  → Optimize database queries or add indexes")

    # Check endpoint performance
    for endpoint, stats in load_endpoints.items():
        p95 = stats.get('p95_ms', 0)
        error_rate = stats.get('error_rate', 0)

        if p95 > 100:
            issues.append(f"⚠ WARNING: Slow endpoint {endpoint} (P95: {p95:.2f}ms)")

        if error_rate > 1:
            issues.append(f"⚠ CRITICAL: High error rate on {endpoint} ({error_rate:.2f}%)")

    if issues:
        for issue in issues:
            report.append(issue)
    else:
        report.append("✓ No critical performance issues detected")

    # Recommendations
    report.append("\n6. OPTIMIZATION RECOMMENDATIONS")
    report.append("-" * 80)

    recommendations = []

    if goroutine_growth > 500:
        recommendations.append("1. Implement worker pool pattern to limit goroutine creation")
        recommendations.append("   See: PERFORMANCE_PROFILING_REPORT.md Section 7.1")

    if heap_growth > 20:
        recommendations.append("2. Investigate heap allocations with 'go tool pprof heap.prof'")
        recommendations.append("   Focus on frequently called functions")

    if load_cache_rate < 80:
        recommendations.append("3. Optimize cache strategy:")
        recommendations.append("   - Increase cache TTL for stable data")
        recommendations.append("   - Implement cache warming for popular URLs")

    if load_slow > 0:
        recommendations.append("4. Optimize database queries:")
        recommendations.append("   - Add missing indexes")
        recommendations.append("   - Implement batch operations")
        recommendations.append("   - Use prepared statements")

    if not recommendations:
        recommendations.append("✓ System performing well under current load")
        recommendations.append("  Consider increasing load to find capacity limits")

    for rec in recommendations:
        report.append(rec)

    # Summary
    report.append("\n" + "=" * 80)
    report.append("SUMMARY")
    report.append("=" * 80)

    # Calculate overall health score
    health_score = 100

    if goroutine_growth > 1000:
        health_score -= 30
    elif goroutine_growth > 500:
        health_score -= 15

    if heap_growth > 50:
        health_score -= 20
    elif heap_growth > 20:
        health_score -= 10

    if load_cache_rate < 70:
        health_score -= 20
    elif load_cache_rate < 80:
        health_score -= 10

    if load_slow > 0 and load_queries > 0:
        slow_rate = (load_slow / load_queries) * 100
        if slow_rate > 20:
            health_score -= 20
        elif slow_rate > 10:
            health_score -= 10

    report.append(f"\nPerformance Health Score: {max(0, health_score)}/100")

    if health_score >= 80:
        report.append("Status: ✓ GOOD - System performing well")
    elif health_score >= 60:
        report.append("Status: ⚠ MODERATE - Some optimization needed")
    else:
        report.append("Status: ✗ POOR - Critical optimization required")

    report.append("\nNext Steps:")
    report.append("1. Review detailed profiles: go tool pprof -http=:8081 cpu.prof")
    report.append("2. Implement Priority 1 optimizations from PERFORMANCE_PROFILING_REPORT.md")
    report.append("3. Re-run profiling after optimizations")
    report.append("4. Compare before/after metrics")

    report.append("\n" + "=" * 80)

    # Write report
    report_text = "\n".join(report)
    with open(output_file, 'w') as f:
        f.write(report_text)

    print(report_text)
    print(f"\nReport saved to: {output_file}")


def main():
    if len(sys.argv) != 4:
        print("Usage: python analyze_profiles.py <baseline_metrics.json> <load_metrics.json> <output_report.txt>")
        sys.exit(1)

    baseline_file = sys.argv[1]
    load_file = sys.argv[2]
    output_file = sys.argv[3]

    # Check if files exist
    if not Path(baseline_file).exists():
        print(f"Error: {baseline_file} not found")
        sys.exit(1)

    if not Path(load_file).exists():
        print(f"Error: {load_file} not found")
        sys.exit(1)

    analyze_metrics(baseline_file, load_file, output_file)


if __name__ == '__main__':
    main()

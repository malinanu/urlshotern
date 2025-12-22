# URL Shortener Performance Profiling Script (PowerShell)
# Windows-compatible version of run_profiling.sh

param(
    [int]$ServerPort = 8080,
    [int]$ProfileDuration = 30,
    [int]$LoadDuration = 60,
    [int]$LoadConnections = 100
)

$ErrorActionPreference = "Continue"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "URL Shortener Performance Profiling Suite" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Configuration
$OutputDir = "profiling_results_$(Get-Date -Format 'yyyyMMdd_HHmmss')"
New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

Write-Host "Output directory: $OutputDir" -ForegroundColor Yellow
Write-Host "Server port: $ServerPort" -ForegroundColor Yellow
Write-Host "Profile duration: ${ProfileDuration}s" -ForegroundColor Yellow
Write-Host ""

# Check if server is running
function Test-Server {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:$ServerPort/health" -UseBasicParsing -TimeoutSec 5
        Write-Host "✓ Server is running" -ForegroundColor Green
        return $true
    }
    catch {
        Write-Host "ERROR: Server is not running on port $ServerPort" -ForegroundColor Red
        Write-Host "Start the profiling server with: go run cmd/profiling/profile_server.go" -ForegroundColor Yellow
        return $false
    }
}

# Function to run CPU profiling
function Start-CPUProfiling {
    Write-Host ""
    Write-Host "1. CPU Profiling" -ForegroundColor Cyan
    Write-Host "   Duration: ${ProfileDuration}s"
    Write-Host "   Starting CPU profile collection..."

    # Start CPU profile download in background
    $job = Start-Job -ScriptBlock {
        param($Port, $Duration, $OutputFile)
        Invoke-WebRequest -Uri "http://localhost:$Port/debug/pprof/profile?seconds=$Duration" `
            -OutFile $OutputFile -UseBasicParsing
    } -ArgumentList $ServerPort, $ProfileDuration, "$OutputDir\cpu.prof"

    # Generate load during profiling
    Write-Host "   Generating load during profiling..."
    if (Get-Command wrk -ErrorAction SilentlyContinue) {
        wrk -t4 -c50 -d${ProfileDuration}s --latency "http://localhost:$ServerPort/health" `
            > "$OutputDir\cpu_load_results.txt" 2>&1
    }
    else {
        Write-Host "   Warning: wrk not found, skipping load generation" -ForegroundColor Yellow
        Start-Sleep -Seconds $ProfileDuration
    }

    # Wait for profile download
    Wait-Job $job | Out-Null
    Receive-Job $job | Out-Null
    Remove-Job $job

    Write-Host "   ✓ CPU profile saved to $OutputDir\cpu.prof" -ForegroundColor Green
    Write-Host "   Generating flame graph..."
    go tool pprof -png "$OutputDir\cpu.prof" > "$OutputDir\cpu_flamegraph.png" 2>$null
    Write-Host "   ✓ Flame graph saved to $OutputDir\cpu_flamegraph.png" -ForegroundColor Green
}

# Function to run heap profiling
function Start-HeapProfiling {
    Write-Host ""
    Write-Host "2. Heap Memory Profiling" -ForegroundColor Cyan
    Write-Host "   Collecting heap snapshot..."

    Invoke-WebRequest -Uri "http://localhost:$ServerPort/debug/pprof/heap" `
        -OutFile "$OutputDir\heap.prof" -UseBasicParsing

    Write-Host "   ✓ Heap profile saved to $OutputDir\heap.prof" -ForegroundColor Green
    Write-Host "   Generating visualization..."
    go tool pprof -png "$OutputDir\heap.prof" > "$OutputDir\heap_visualization.png" 2>$null
    go tool pprof -top "$OutputDir\heap.prof" > "$OutputDir\heap_top.txt" 2>$null
    Write-Host "   ✓ Heap analysis saved" -ForegroundColor Green
}

# Function to run goroutine profiling
function Start-GoroutineProfiling {
    Write-Host ""
    Write-Host "3. Goroutine Profiling" -ForegroundColor Cyan
    Write-Host "   Collecting goroutine snapshot..."

    Invoke-WebRequest -Uri "http://localhost:$ServerPort/debug/pprof/goroutine" `
        -OutFile "$OutputDir\goroutine.prof" -UseBasicParsing

    Write-Host "   ✓ Goroutine profile saved to $OutputDir\goroutine.prof" -ForegroundColor Green
    Write-Host "   Analyzing goroutines..."
    go tool pprof -top "$OutputDir\goroutine.prof" > "$OutputDir\goroutine_analysis.txt" 2>$null
    Write-Host "   ✓ Goroutine analysis saved" -ForegroundColor Green
}

# Function to collect allocation profiling
function Start-AllocationProfiling {
    Write-Host ""
    Write-Host "4. Allocation Profiling" -ForegroundColor Cyan
    Write-Host "   Collecting allocation data..."

    Invoke-WebRequest -Uri "http://localhost:$ServerPort/debug/pprof/allocs" `
        -OutFile "$OutputDir\allocs.prof" -UseBasicParsing

    Write-Host "   ✓ Allocation profile saved to $OutputDir\allocs.prof" -ForegroundColor Green
    go tool pprof -top "$OutputDir\allocs.prof" > "$OutputDir\allocs_top.txt" 2>$null
    Write-Host "   ✓ Allocation analysis saved" -ForegroundColor Green
}

# Function to collect block profiling
function Start-BlockingProfiling {
    Write-Host ""
    Write-Host "5. Blocking Operations Profiling" -ForegroundColor Cyan
    Write-Host "   Collecting blocking profile..."

    Invoke-WebRequest -Uri "http://localhost:$ServerPort/debug/pprof/block" `
        -OutFile "$OutputDir\block.prof" -UseBasicParsing

    Write-Host "   ✓ Block profile saved to $OutputDir\block.prof" -ForegroundColor Green
    go tool pprof -top "$OutputDir\block.prof" > "$OutputDir\block_top.txt" 2>$null
    Write-Host "   ✓ Block analysis saved" -ForegroundColor Green
}

# Function to collect mutex profiling
function Start-MutexProfiling {
    Write-Host ""
    Write-Host "6. Mutex Contention Profiling" -ForegroundColor Cyan
    Write-Host "   Collecting mutex profile..."

    Invoke-WebRequest -Uri "http://localhost:$ServerPort/debug/pprof/mutex" `
        -OutFile "$OutputDir\mutex.prof" -UseBasicParsing

    Write-Host "   ✓ Mutex profile saved to $OutputDir\mutex.prof" -ForegroundColor Green
    go tool pprof -top "$OutputDir\mutex.prof" > "$OutputDir\mutex_top.txt" 2>$null
    Write-Host "   ✓ Mutex analysis saved" -ForegroundColor Green
}

# Function to collect application metrics
function Start-MetricsCollection {
    Write-Host ""
    Write-Host "7. Application Metrics" -ForegroundColor Cyan
    Write-Host "   Collecting custom metrics..."

    # Baseline metrics
    Invoke-WebRequest -Uri "http://localhost:$ServerPort/debug/metrics" `
        -OutFile "$OutputDir\metrics_baseline.json" -UseBasicParsing
    Write-Host "   ✓ Baseline metrics collected" -ForegroundColor Green

    # Run load test
    Write-Host "   Running load test..."
    if (Get-Command wrk -ErrorAction SilentlyContinue) {
        wrk -t8 -c$LoadConnections -d${LoadDuration}s --latency `
            "http://localhost:$ServerPort/health" `
            > "$OutputDir\load_test_results.txt" 2>&1
    }
    else {
        Write-Host "   Warning: wrk not found, using basic load generation" -ForegroundColor Yellow
        # Simple PowerShell-based load generation
        $jobs = @()
        for ($i = 0; $i -lt 10; $i++) {
            $jobs += Start-Job -ScriptBlock {
                param($Port, $Duration)
                $endTime = (Get-Date).AddSeconds($Duration)
                while ((Get-Date) -lt $endTime) {
                    try {
                        Invoke-WebRequest -Uri "http://localhost:$Port/health" -UseBasicParsing -TimeoutSec 1 | Out-Null
                    }
                    catch {}
                }
            } -ArgumentList $ServerPort, $LoadDuration
        }
        Wait-Job $jobs | Out-Null
        $jobs | Remove-Job
    }

    # Post-load metrics
    Start-Sleep -Seconds 2
    Invoke-WebRequest -Uri "http://localhost:$ServerPort/debug/metrics" `
        -OutFile "$OutputDir\metrics_under_load.json" -UseBasicParsing
    Write-Host "   ✓ Under-load metrics collected" -ForegroundColor Green
}

# Function to generate summary report
function New-SummaryReport {
    Write-Host ""
    Write-Host "8. Generating Summary Report" -ForegroundColor Cyan

    $summary = @"
URL Shortener - Profiling Summary
Generated: $(Get-Date)
=====================================

Configuration:
- Server Port: $ServerPort
- Profile Duration: ${ProfileDuration}s
- Load Test Duration: ${LoadDuration}s
- Concurrent Connections: $LoadConnections

Files Generated:
- cpu.prof              : CPU profile (view with: go tool pprof -http=:8081 cpu.prof)
- cpu_flamegraph.png    : CPU flame graph visualization
- heap.prof             : Heap memory profile
- heap_visualization.png: Heap usage visualization
- heap_top.txt          : Top memory allocators
- goroutine.prof        : Goroutine profile
- goroutine_analysis.txt: Goroutine stack analysis
- allocs.prof           : Allocation profile
- allocs_top.txt        : Top allocators
- block.prof            : Blocking operations profile
- mutex.prof            : Mutex contention profile
- metrics_baseline.json : Metrics before load
- metrics_under_load.json: Metrics during load
- load_test_results.txt : Load test output

Analysis Commands:
==================

1. Interactive CPU Profile:
   go tool pprof -http=:8081 $OutputDir\cpu.prof

2. Interactive Heap Profile:
   go tool pprof -http=:8081 $OutputDir\heap.prof

3. View Top CPU Consumers:
   go tool pprof -top $OutputDir\cpu.prof

4. View Top Memory Allocators:
   go tool pprof -top $OutputDir\heap.prof

5. View Goroutine Stacks:
   go tool pprof -text $OutputDir\goroutine.prof

Performance Recommendations:
============================
See PERFORMANCE_PROFILING_REPORT.md for detailed analysis and optimization strategies.

Next Steps:
===========
1. Review generated profiles using 'go tool pprof'
2. Compare metrics_baseline.json vs metrics_under_load.json
3. Identify performance bottlenecks
4. Implement optimizations from PERFORMANCE_PROFILING_REPORT.md
5. Re-run profiling to measure improvements
"@

    Set-Content -Path "$OutputDir\PROFILING_SUMMARY.txt" -Value $summary
    Write-Host "   ✓ Summary report generated" -ForegroundColor Green
}

# Main execution
function Main {
    if (-not (Test-Server)) {
        exit 1
    }

    Write-Host ""
    $totalTime = $ProfileDuration + $LoadDuration + 60
    Write-Host "Starting comprehensive profiling..." -ForegroundColor Cyan
    Write-Host "This will take approximately $totalTime seconds" -ForegroundColor Yellow
    Write-Host ""

    Start-CPUProfiling
    Start-HeapProfiling
    Start-GoroutineProfiling
    Start-AllocationProfiling
    Start-BlockingProfiling
    Start-MutexProfiling
    Start-MetricsCollection
    New-SummaryReport

    Write-Host ""
    Write-Host "==========================================" -ForegroundColor Cyan
    Write-Host "Profiling Complete!" -ForegroundColor Green
    Write-Host "==========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Results saved to: $OutputDir\" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Quick Start:" -ForegroundColor Cyan
    Write-Host "  View CPU profile:  go tool pprof -http=:8081 $OutputDir\cpu.prof" -ForegroundColor White
    Write-Host "  View heap profile: go tool pprof -http=:8081 $OutputDir\heap.prof" -ForegroundColor White
    Write-Host "  Read summary:      Get-Content $OutputDir\PROFILING_SUMMARY.txt" -ForegroundColor White
    Write-Host ""

    # Try to analyze with Python if available
    if (Get-Command python -ErrorAction SilentlyContinue) {
        Write-Host "Analyzing metrics with Python..." -ForegroundColor Cyan
        python scripts\analyze_profiles.py `
            "$OutputDir\metrics_baseline.json" `
            "$OutputDir\metrics_under_load.json" `
            "$OutputDir\analysis_report.txt"
        Write-Host ""
        Write-Host "Analysis report: $OutputDir\analysis_report.txt" -ForegroundColor Yellow
    }
}

# Run main function
Main

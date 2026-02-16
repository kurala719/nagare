# Redis Task Queue Test Script
# This script demonstrates the async queue and alert generation functionality

param(
    [string]$BaseUrl = "http://localhost:8080/api/v1",
    [string]$Token = $null
)

if (-not $Token) {
    Write-Host "Usage: .\test_queue.ps1 -Token 'your-jwt-token'" -ForegroundColor Yellow
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type"  = "application/json"
}

function Test-QueueStats {
    Write-Host "`n=== Queue Stats ===" -ForegroundColor Cyan
    try {
        $response = Invoke-WebRequest -Uri "$BaseUrl/queue/stats" `
            -Method GET `
            -Headers $headers
        $data = $response.Content | ConvertFrom-Json
        $data | ConvertTo-Json | Write-Host -ForegroundColor Green
    }
    catch {
        Write-Host "Error: $_" -ForegroundColor Red
    }
}

function Test-PullHostsAsync {
    param([int]$MonitorId = 1)
    Write-Host "`n=== Queue Async Host Pull ===" -ForegroundColor Cyan
    Write-Host "Monitor ID: $MonitorId" -ForegroundColor Yellow
    try {
        $response = Invoke-WebRequest -Uri "$BaseUrl/monitors/$MonitorId/hosts/pull-async" `
            -Method POST `
            -Headers $headers
        $data = $response.Content | ConvertFrom-Json
        $data | ConvertTo-Json | Write-Host -ForegroundColor Green
    }
    catch {
        Write-Host "Error: $_" -ForegroundColor Red
    }
}

function Test-PullItemsAsync {
    param(
        [int]$MonitorId = 1,
        [int]$HostId = 1
    )
    Write-Host "`n=== Queue Async Item Pull ===" -ForegroundColor Cyan
    Write-Host "Monitor ID: $MonitorId, Host ID: $HostId" -ForegroundColor Yellow
    try {
        $response = Invoke-WebRequest -Uri "$BaseUrl/monitors/$MonitorId/hosts/$HostId/items/pull-async" `
            -Method POST `
            -Headers $headers
        $data = $response.Content | ConvertFrom-Json
        $data | ConvertTo-Json | Write-Host -ForegroundColor Green
    }
    catch {
        Write-Host "Error: $_" -ForegroundColor Red
    }
}

function Test-GenerateAlerts {
    param([int]$Count = 5)
    Write-Host "`n=== Generate Test Alerts ===" -ForegroundColor Cyan
    Write-Host "Alert Count: $Count" -ForegroundColor Yellow
    try {
        $response = Invoke-WebRequest -Uri "$BaseUrl/alerts/generate-test?count=$Count" `
            -Method POST `
            -Headers $headers
        $data = $response.Content | ConvertFrom-Json
        $data | ConvertTo-Json | Write-Host -ForegroundColor Green
    }
    catch {
        Write-Host "Error: $_" -ForegroundColor Red
    }
}

function Test-AlertScore {
    Write-Host "`n=== Get Alert Score ===" -ForegroundColor Cyan
    try {
        $response = Invoke-WebRequest -Uri "$BaseUrl/alerts/score" `
            -Method GET `
            -Headers $headers
        $data = $response.Content | ConvertFrom-Json
        Write-Host "Alert Score: $($data.score)" -ForegroundColor Green
    }
    catch {
        Write-Host "Error: $_" -ForegroundColor Red
    }
}

# Run tests
Write-Host "Testing Redis Task Queue Implementation" -ForegroundColor Magenta
Write-Host "Base URL: $BaseUrl" -ForegroundColor Gray

# Initial stats
Test-QueueStats

# Queue some tasks
Write-Host "`n--- Queuing Tasks ---" -ForegroundColor Magenta
Test-PullHostsAsync -MonitorId 1
Start-Sleep -Milliseconds 500
Test-PullItemsAsync -MonitorId 1 -HostId 1
Start-Sleep -Milliseconds 500

# Check queue stats after queueing
Test-QueueStats

# Generate alerts
Write-Host "`n--- Generating Test Alerts ---" -ForegroundColor Magenta
Test-GenerateAlerts -Count 10
Start-Sleep -Seconds 1

# Check alert score
Test-AlertScore

# Wait for workers to process
Write-Host "`nWaiting 3 seconds for workers to process tasks..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# Final stats
Test-QueueStats

Write-Host "`n✓ Test completed" -ForegroundColor Green

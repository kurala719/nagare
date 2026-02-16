#!/usr/bin/env pwsh

# Test script for QQ webhook endpoint

Write-Host "Testing QQ Webhook Endpoint" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"
$endpoint = "/api/v1/media/qq/message"

# Test 1: Private message with /status command
Write-Host "Test 1: Private message - /status command" -ForegroundColor Yellow
$body1 = @{
    post_type = "message"
    message_type = "private"
    user_id = 123456789
    message = "/status"
    message_id = 1
    time = [DateTimeOffset]::Now.ToUnixTimeSeconds()
} | ConvertTo-Json

try {
    $response1 = Invoke-WebRequest -Uri "$baseUrl$endpoint" `
        -Method POST `
        -Headers @{"Content-Type"="application/json"} `
        -Body $body1 `
        -UseBasicParsing

    Write-Host "✓ Status: $($response1.StatusCode)" -ForegroundColor Green
    Write-Host "Response: $($response1.Content)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Status Code: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = [System.IO.StreamReader]::new($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response Body: $responseBody" -ForegroundColor Red
    }
}

Write-Host ""

# Test 2: Testing other working webhooks for comparison
Write-Host "Test 2: Testing /im/command (should work)" -ForegroundColor Yellow
$body2 = @{
    message = "/status"
} | ConvertTo-Json

try {
    $response2 = Invoke-WebRequest -Uri "$baseUrl/api/v1/im/command" `
        -Method POST `
        -Headers @{"Content-Type"="application/json"} `
        -Body $body2 `
        -UseBasicParsing

    Write-Host "✓ Status: $($response2.StatusCode)" -ForegroundColor Green
    Write-Host "Response: $($response2.Content)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 3: Testing authenticated endpoint (should fail)
Write-Host "Test 3: Testing /media (should return 401 without auth)" -ForegroundColor Yellow
try {
    $response3 = Invoke-WebRequest -Uri "$baseUrl/api/v1/media" `
        -Method GET `
        -UseBasicParsing

    Write-Host "✓ Status: $($response3.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "✓ Expected 401: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Testing complete!" -ForegroundColor Cyan

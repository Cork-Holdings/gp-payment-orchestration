# Merchant API Verification Script

$GatewayURL = "http://localhost:8080"
if ($env:LISTEN_ADDR) { $GatewayURL = "http://localhost:$($env:LISTEN_ADDR)" }

Write-Host "============================================="
Write-Host "Starting GeePay Merchant API Verification"
Write-Host "============================================="
Write-Host ""

# Helper function to compute HMAC-SHA256 signature in PowerShell
function Get-HMACSignature($message, $secret) {
    $hmac = New-Object System.Security.Cryptography.HMACSHA256
    $hmac.Key = [System.Text.Encoding]::UTF8.GetBytes($secret)
    $bodyBytes = [System.Text.Encoding]::UTF8.GetBytes($message)
    $sigBytes = $hmac.ComputeHash($bodyBytes)
    return [System.BitConverter]::ToString($sigBytes).Replace("-", "").ToLower()
}

# 1. OAuth Client Credentials token generation
Write-Host "1. Testing Token Generation (POST /oauth/token)..."
try {
    $tokenBody = @{
        client_id = "merchant_123"
        client_secret = "secret_456"
    }
    $res = Invoke-RestMethod -Uri "$GatewayURL/oauth/token" -Method Post -Body $tokenBody
    $token = $res.access_token
    Write-Host "SUCCESS: Access token generated successfully!" -ForegroundColor Green
    Write-Host "Token: $token"
} catch {
    Write-Error "Token Generation Failed: $_"
    return
}
Write-Host ""

# 2. Token generation with invalid credentials
Write-Host "2. Testing Token Generation with Invalid Credentials..."
try {
    $badTokenBody = @{
        client_id = "merchant_123"
        client_secret = "wrong_secret"
    }
    $res = Invoke-RestMethod -Uri "$GatewayURL/oauth/token" -Method Post -Body $badTokenBody
    Write-Host "failed: Expected token generation to fail, but it succeeded!" -ForegroundColor Red
} catch {
    $status = $_.Exception.Response.StatusCode.value__
    if ($status -eq 401) {
        Write-Host "SUCCESS: Rejected with HTTP 401 Unauthorized as expected!" -ForegroundColor Green
    } else {
        Write-Error "Unexpected failure status ($status): $_"
    }
}
Write-Host ""

# 3. Mobile money collections (Valid 12-digit phone number, expects 202 Accepted)
Write-Host "3. Testing Collections (POST /api/v1/mobile-money/collect)..."
try {
    $collectBody = @{
        phone_number = "260971234567" # Exactly 12 digits
        amount = 150.0
        currency = "USD"
    } | ConvertTo-Json

    $headers = @{
        "Authorization" = "Bearer $token"
        "X-Client-ID" = "merchant_123"
    }

    $res = Invoke-RestMethod -Uri "$GatewayURL/api/v1/mobile-money/collect" -Method Post -Body $collectBody -ContentType "application/json" -Headers $headers
    Write-Host "SUCCESS: Collection request accepted!" -ForegroundColor Green
    Write-Host "Tracking Ref: $($res.tracking_ref)"
    Write-Host "Status: $($res.status)"
} catch {
    Write-Error "Collections Failed: $_"
}
Write-Host ""

# 4. Mobile money collections with invalid phone number
Write-Host "4. Testing Collections with Invalid Phone Number (10 digits)..."
try {
    $badPhoneBody = @{
        phone_number = "0971234567" # Not 12 digits
        amount = 50.0
        currency = "USD"
    } | ConvertTo-Json

    $headers = @{
        "Authorization" = "Bearer $token"
        "X-Client-ID" = "merchant_123"
    }

    $res = Invoke-RestMethod -Uri "$GatewayURL/api/v1/mobile-money/collect" -Method Post -Body $badPhoneBody -ContentType "application/json" -Headers $headers
    Write-Host "failed: Accepted invalid phone number!" -ForegroundColor Red
} catch {
    $status = $_.Exception.Response.StatusCode.value__
    if ($status -eq 400) {
        Write-Host "SUCCESS: Rejected with HTTP 400 Bad Request as expected!" -ForegroundColor Green
    } else {
        Write-Error "Unexpected failure status ($status): $_"
    }
}
Write-Host ""

# 5. IP Whitelisting Edge Security check (Mocking bad IP via X-Forwarded-For)
Write-Host "5. Testing IP Whitelisting (Mocking unapproved IP)..."
try {
    $collectBody = @{
        phone_number = "260971234567"
        amount = 10.0
        currency = "USD"
    } | ConvertTo-Json

    $headers = @{
        "Authorization" = "Bearer $token"
        "X-Client-ID" = "merchant_123"
        "X-Forwarded-For" = "8.8.8.8" # Blocked IP
    }

    $res = Invoke-RestMethod -Uri "$GatewayURL/api/v1/mobile-money/collect" -Method Post -Body $collectBody -ContentType "application/json" -Headers $headers
    Write-Host "failed: Allowed request from unapproved IP address!" -ForegroundColor Red
} catch {
    $status = $_.Exception.Response.StatusCode.value__
    if ($status -eq 401) {
        Write-Host "SUCCESS: Rejected with HTTP 401 Unauthorized as expected!" -ForegroundColor Green
    } else {
        Write-Error "Unexpected failure status ($status): $_"
    }
}
Write-Host ""

# 6. Disbursements with correct signature and balance ($150)
Write-Host "6. Testing Successful Disbursement (POST /api/v1/mobile-money/disburse)..."
try {
    $disburseObj = @{
        phone_number = "260971234567"
        amount = 150.0
        currency = "USD"
    }
    $disburseBody = $disburseObj | ConvertTo-Json -Compress
    
    # Compute signature using merchant's client secret
    $signature = Get-HMACSignature -message $disburseBody -secret "secret_456"

    $headers = @{
        "Authorization" = "Bearer $token"
        "X-Client-ID" = "merchant_123"
        "X-Auth-Signature" = $signature
    }

    $res = Invoke-RestMethod -Uri "$GatewayURL/api/v1/mobile-money/disburse" -Method Post -Body $disburseBody -ContentType "application/json" -Headers $headers
    Write-Host "SUCCESS: Disbursement approved!" -ForegroundColor Green
    Write-Host "Tracking Ref: $($res.tracking_ref)"
    Write-Host "Status: $($res.status)"
} catch {
    Write-Error "Disbursement Failed: $_"
}
Write-Host ""

# 7. Disbursements with tampered signature
Write-Host "7. Testing Disbursement with Tampered Signature..."
try {
    $disburseObj = @{
        phone_number = "260971234567"
        amount = 150.0
        currency = "USD"
    }
    $disburseBody = $disburseObj | ConvertTo-Json -Compress

    $headers = @{
        "Authorization" = "Bearer $token"
        "X-Client-ID" = "merchant_123"
        "X-Auth-Signature" = "invalid_signature_hex"
    }

    $res = Invoke-RestMethod -Uri "$GatewayURL/api/v1/mobile-money/disburse" -Method Post -Body $disburseBody -ContentType "application/json" -Headers $headers
    Write-Host "failed: Approved disbursement with invalid signature!" -ForegroundColor Red
} catch {
    $status = $_.Exception.Response.StatusCode.value__
    if ($status -eq 400) {
        Write-Host "SUCCESS: Rejected with HTTP 400 and error code!" -ForegroundColor Green
    } else {
        Write-Error "Unexpected failure status ($status): $_"
    }
}
Write-Host ""

# 8. Disbursements with insufficient balance (Requesting $5000, balance is $1000)
Write-Host "8. Testing Disbursement with Insufficient Balance..."
try {
    $disburseObj = @{
        phone_number = "260971234567"
        amount = 5000.0 # Exceeds limit
        currency = "USD"
    }
    $disburseBody = $disburseObj | ConvertTo-Json -Compress
    
    $signature = Get-HMACSignature -message $disburseBody -secret "secret_456"

    $headers = @{
        "Authorization" = "Bearer $token"
        "X-Client-ID" = "merchant_123"
        "X-Auth-Signature" = $signature
    }

    $res = Invoke-RestMethod -Uri "$GatewayURL/api/v1/mobile-money/disburse" -Method Post -Body $disburseBody -ContentType "application/json" -Headers $headers
    Write-Host "failed: Approved disbursement with insufficient balance!" -ForegroundColor Red
} catch {
    $status = $_.Exception.Response.StatusCode.value__
    if ($status -eq 400) {
        Write-Host "SUCCESS: Rejected with HTTP 400 and INSUFFICIENT_BALANCE as expected!" -ForegroundColor Green
    } else {
        Write-Error "Unexpected failure status ($status): $_"
    }
}
Write-Host ""

Write-Host "============================================="
Write-Host "Verification completed!"
Write-Host "============================================="

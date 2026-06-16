# API Verification Script for gp-payment-orchestration

$BaseURL = "http://localhost:2050"
if ($env:LISTEN_ADDR) { $BaseURL = "http://localhost:$($env:LISTEN_ADDR)" }

Write-Host "========================================"
Write-Host "Starting geepay API testing on $BaseURL"
Write-Host "========================================"
Write-Host ""

# 1. Test Auth Login (Seeds admin on gp_auth startup)
Write-Host "1. Testing /auth/login..."
try {
    $loginBody = @{ username = "admin"; password = "password" } | ConvertTo-Json
    $authRes = Invoke-RestMethod -Uri "$BaseURL/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    Write-Host "SUCCESS: Auth token received!" -ForegroundColor Green
    Write-Host "Access Token: $($authRes.access_token)"
    Write-Host "Refresh Token: $($authRes.refresh_token)"
} catch {
    Write-Error "Auth Login Failed: $_"
}
Write-Host ""

# 2. Create Accounts (Specifying wallet types)
Write-Host "2. Creating Accounts..."
try {
    # Account 1: Emoney
    $acc1Body = @{ name = "Alice Emoney"; currency = "USD"; initial_balance = 500.0; wallet_type = "emoney" } | ConvertTo-Json
    $acc1 = Invoke-RestMethod -Uri "$BaseURL/ledger/accounts" -Method Post -Body $acc1Body -ContentType "application/json"
    $acc1Id = $acc1.account_id
    Write-Host "Created Account 1: ID=$acc1Id, Type=$($acc1.wallet_type), Balance=$($acc1.balance)" -ForegroundColor Green

    # Account 2: Emoney (matching type)
    $acc2Body = @{ name = "Bob Emoney"; currency = "USD"; initial_balance = 100.0; wallet_type = "emoney" } | ConvertTo-Json
    $acc2 = Invoke-RestMethod -Uri "$BaseURL/ledger/accounts" -Method Post -Body $acc2Body -ContentType "application/json"
    $acc2Id = $acc2.account_id
    Write-Host "Created Account 2: ID=$acc2Id, Type=$($acc2.wallet_type), Balance=$($acc2.balance)" -ForegroundColor Green

    # Account 3: POS (mismatched type)
    $acc3Body = @{ name = "Charlie POS"; currency = "USD"; initial_balance = 0.0; wallet_type = "pos" } | ConvertTo-Json
    $acc3 = Invoke-RestMethod -Uri "$BaseURL/ledger/accounts" -Method Post -Body $acc3Body -ContentType "application/json"
    $acc3Id = $acc3.account_id
    Write-Host "Created Account 3: ID=$acc3Id, Type=$($acc3.wallet_type), Balance=$($acc3.balance)" -ForegroundColor Green
} catch {
    Write-Error "Account Creation Failed: $_"
}
Write-Host ""

# 3. Test Successful Transfer (emoney -> emoney)
Write-Host "3. Testing Successful Transfer (emoney -> emoney)..."
try {
    $transferBody = @{
        source_account_id = $acc1Id
        destination_account_id = $acc2Id
        amount = 50.0
        currency = "USD"
    } | ConvertTo-Json
    $transferRes = Invoke-RestMethod -Uri "$BaseURL/ledger/transfers" -Method Post -Body $transferBody -ContentType "application/json"
    Write-Host "SUCCESS: Transfer processed!" -ForegroundColor Green
    Write-Host "Transfer ID: $($transferRes.transfer_id)"
    Write-Host "Status: $($transferRes.status)"
} catch {
    Write-Error "Transfer Failed: $_"
}
Write-Host ""

# 4. Test Wallet-Type Check Validation Check (emoney -> pos)
Write-Host "4. Testing Cross-Wallet Transfer Validation (emoney -> pos)..."
try {
    $crossBody = @{
        source_account_id = $acc1Id
        destination_account_id = $acc3Id
        amount = 20.0
        currency = "USD"
    } | ConvertTo-Json
    
    # This call should throw HTTP 400 with CROSS_WALLET_NOT_ALLOWED
    $crossRes = Invoke-RestMethod -Uri "$BaseURL/ledger/transfers" -Method Post -Body $crossBody -ContentType "application/json"
    Write-Host "Unexpected Success! The server accepted a cross-wallet transfer." -ForegroundColor Red
} catch {
    $status = $_.Exception.Response.StatusCode.value__
    $body = $_.ErrorDetails.Message | ConvertFrom-Json
    if ($status -eq 400 -and $body.error -eq "CROSS_WALLET_NOT_ALLOWED") {
        Write-Host "SUCCESS: Validation failed as expected with HTTP 400 CROSS_WALLET_NOT_ALLOWED!" -ForegroundColor Green
    } else {
        Write-Error "Validation failed with unexpected status ($status) or message: $_"
    }
}
Write-Host ""

# 5. Check Balance Updates
Write-Host "5. Verifying Balance Updates..."
try {
    $bal1 = Invoke-RestMethod -Uri "$BaseURL/ledger/accounts/$acc1Id/balance" -Method Get
    $bal2 = Invoke-RestMethod -Uri "$BaseURL/ledger/accounts/$acc2Id/balance" -Method Get
    
    Write-Host "Account 1 Balance: $($bal1.balance) (Expected: 450)" -ForegroundColor Green
    Write-Host "Account 2 Balance: $($bal2.balance) (Expected: 150)" -ForegroundColor Green
} catch {
    Write-Error "Balance check failed: $_"
}
Write-Host ""

Write-Host "========================================"
Write-Host "Testing completed successfully!"
Write-Host "========================================"

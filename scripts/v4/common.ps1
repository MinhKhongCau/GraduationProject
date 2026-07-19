Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$script:RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$script:BackendRoot = Join-Path $script:RepoRoot "app\backend"
$script:ComposeFile = Join-Path $script:RepoRoot "docker-compose.dev.huy.yml"
$script:ComposeProject = "mindcare-v4"
$script:Gateway = "http://localhost:8000"
$script:Results = [ordered]@{}
$script:TxnSequence = 91000000

function Write-Step([string]$Message) { Write-Host "`n[V4] $Message" -ForegroundColor Cyan }

function Assert-True([bool]$Condition, [string]$Message) {
    if (-not $Condition) { throw "ASSERTION FAILED: $Message" }
}

function Assert-Equal($Expected, $Actual, [string]$Message) {
    if ([string]$Expected -ne [string]$Actual) {
        throw "ASSERTION FAILED: $Message (expected=$Expected actual=$Actual)"
    }
}

function Get-EnvValue([string]$Name, [string]$Default = "") {
    $line = Get-Content (Join-Path $script:RepoRoot ".env") |
        Where-Object { $_ -match "^$([regex]::Escape($Name))=" } |
        Select-Object -First 1
    if (-not $line) { return $Default }
    return $line.Substring($line.IndexOf("=") + 1).Trim().Trim('"').Trim("'")
}

function Invoke-Compose {
    & docker compose -p $script:ComposeProject -f $script:ComposeFile @args
    if ($LASTEXITCODE -ne 0) { throw "docker compose failed: $($args -join ' ')" }
}

function Invoke-ComposeWithOverride {
    $override = Join-Path $PSScriptRoot "docker-compose.auth-failure.yml"
    & docker compose -p $script:ComposeProject -f $script:ComposeFile -f $override @args
    if ($LASTEXITCODE -ne 0) { throw "docker compose override failed: $($args -join ' ')" }
}

function Wait-Http([string]$Url, [int]$TimeoutSeconds = 120) {
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    do {
        try {
            $response = Invoke-WebRequest -UseBasicParsing -Uri $Url -TimeoutSec 5
            if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 500) { return }
        } catch {}
        Start-Sleep -Seconds 2
    } while ((Get-Date) -lt $deadline)
    throw "Timed out waiting for $Url"
}

function Convert-ResponseBody($Response) {
    if ([string]::IsNullOrWhiteSpace($Response.Content)) { return $null }
    try { return $Response.Content | ConvertFrom-Json } catch { return $null }
}

function Invoke-Api {
    param([string]$Method, [string]$Path, $Body = $null, [string]$Token = "", [hashtable]$Headers = @{})
    $requestHeaders = @{}
    foreach ($key in $Headers.Keys) { $requestHeaders[$key] = $Headers[$key] }
    if ($Token) { $requestHeaders.Authorization = "Bearer $Token" }
    $params = @{
        UseBasicParsing = $true
        Method = $Method
        Uri = if ($Path.StartsWith("http")) { $Path } else { "$script:Gateway$Path" }
        Headers = $requestHeaders
        TimeoutSec = 30
    }
    if ($null -ne $Body) {
        $params.ContentType = "application/json"
        $params.Body = if ($Body -is [string]) { $Body } else { $Body | ConvertTo-Json -Depth 12 -Compress }
    }
    try {
        $response = Invoke-WebRequest @params
        return [pscustomobject]@{ Status = [int]$response.StatusCode; Raw = $response.Content; Json = Convert-ResponseBody $response }
    } catch {
        if ($null -eq $_.Exception.Response) { throw }
        $response = $_.Exception.Response
        $reader = [System.IO.StreamReader]::new($response.GetResponseStream())
        try { $raw = $reader.ReadToEnd() } finally { $reader.Dispose() }
        $json = $null
        if ($raw) { try { $json = $raw | ConvertFrom-Json } catch {} }
        return [pscustomobject]@{ Status = [int]$response.StatusCode; Raw = $raw; Json = $json }
    }
}

function Login([string]$Email, [string]$Password) {
    $response = Invoke-Api POST "/api/v1/auth/login" @{ email = $Email; password = $Password }
    Assert-Equal 200 $response.Status "login must succeed for $Email"
    Assert-True ([bool]$response.Json.data.accessToken) "login response must include data.accessToken"
    return [pscustomobject]@{
        Token = [string]$response.Json.data.accessToken
        ID = [string]$response.Json.data.accountId
        Role = [string]$response.Json.data.role
        Sanitized = [ordered]@{
            success = $response.Json.success
            message = $response.Json.message
            data = [ordered]@{
                accessToken = "<REDACTED>"
                accountId = $response.Json.data.accountId
                fullName = $response.Json.data.fullName
                refreshToken = "<REDACTED>"
                role = $response.Json.data.role
            }
        }
    }
}

function Invoke-Sql([string]$Database, [string]$Sql) {
    $dbUser = Get-EnvValue "DB_USER" "admin"
    $previousPreference = $ErrorActionPreference
    try {
        $ErrorActionPreference = "Continue"
        $output = $Sql | & docker exec -e PGOPTIONS=--client-min-messages=warning -i postgres-db psql -X -q -v ON_ERROR_STOP=1 -U $dbUser -d $Database -At -F "|" 2>&1
        $exitCode = $LASTEXITCODE
    } finally { $ErrorActionPreference = $previousPreference }
    if ($exitCode -ne 0) { throw "SQL failed on $Database`: $($output -join "`n")" }
    return (($output | ForEach-Object { [string]$_ }) -join "`n").Trim()
}

function Invoke-Migration([string]$Database, [string]$Path) {
    [void](Invoke-Sql $Database (Get-Content -Raw $Path))
}

function Wait-Sql([string]$Database, [string]$Sql, [string]$Expected, [int]$TimeoutSeconds = 45) {
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    do {
        $actual = Invoke-Sql $Database $Sql
        if ($actual -eq $Expected) { return $actual }
        Start-Sleep -Seconds 1
    } while ((Get-Date) -lt $deadline)
    throw "Timed out waiting for SQL result '$Expected'; last result '$actual'"
}

function ConvertTo-GoQueryEscape([string]$Value) { return [Uri]::EscapeDataString($Value).Replace("%20", "+") }

function New-SignedIPN {
    param(
        [string]$OrderID,
        [long]$AmountVND,
        [string]$ResponseCode = "00",
        [string]$TransactionNumber = "",
        [switch]$OmitAmount,
        [switch]$OmitResponseCode,
        [string]$RawAmount = ""
    )
    if (-not $TransactionNumber) { $script:TxnSequence++; $TransactionNumber = [string]$script:TxnSequence }
    $params = @{
        vnp_TxnRef = $OrderID
        vnp_TransactionNo = $TransactionNumber
        vnp_TransactionStatus = if ($ResponseCode -eq "00") { "00" } else { $ResponseCode }
        vnp_PayDate = (Get-Date).ToString("yyyyMMddHHmmss")
    }
    if (-not $OmitAmount) { $params.vnp_Amount = if ($RawAmount) { $RawAmount } else { [string]($AmountVND * 100) } }
    if (-not $OmitResponseCode) { $params.vnp_ResponseCode = $ResponseCode }
    $pairs = foreach ($key in ($params.Keys | Sort-Object)) {
        "$(ConvertTo-GoQueryEscape $key)=$(ConvertTo-GoQueryEscape ([string]$params[$key]))"
    }
    $raw = $pairs -join "&"
    $secret = Get-EnvValue "VNP_HASH_SECRET"
    Assert-True (-not [string]::IsNullOrWhiteSpace($secret)) "VNP_HASH_SECRET must be configured"
    $hmac = [Security.Cryptography.HMACSHA512]::new([Text.Encoding]::UTF8.GetBytes($secret))
    try {
        $hash = ([BitConverter]::ToString($hmac.ComputeHash([Text.Encoding]::UTF8.GetBytes($raw)))).Replace("-", "").ToLowerInvariant()
    } finally { $hmac.Dispose() }
    return "$script:Gateway/api/v1/payments/vnpay-ipn?$raw&vnp_SecureHash=$hash"
}

function Send-IPN {
    param([string]$OrderID, [long]$AmountVND, [string]$ResponseCode = "00", [string]$TransactionNumber = "")
    return Invoke-Api GET (New-SignedIPN -OrderID $OrderID -AmountVND $AmountVND -ResponseCode $ResponseCode -TransactionNumber $TransactionNumber)
}

function Get-AvailableSlots([string]$ExpertID, [string]$Date) {
    $response = Invoke-Api GET "/api/v1/public/booking/slots/available-times?expert_id=$ExpertID&date=$Date"
    Assert-Equal 200 $response.Status "available times query"
    return @($response.Json.data.available_times)
}

function New-AppointmentAndOrder {
    param([string]$SlotID, [string]$ExpertID, [string]$PatientToken, [switch]$NoOrder)
    $lock = Invoke-Api POST "/api/v1/booking/slots/$SlotID/lock" $null $PatientToken
    Assert-Equal 200 $lock.Status "slot lock"
    $appointment = Invoke-Api POST "/api/v1/booking/appointments" @{ slot_id = $SlotID; expert_id = $ExpertID } $PatientToken
    Assert-Equal 200 $appointment.Status "appointment create"
    $appointmentID = [string]$appointment.Json.data.appointment_id
    if ($NoOrder) { return [pscustomobject]@{ SlotID = $SlotID; AppointmentID = $appointmentID; Lock = $lock; Appointment = $appointment } }
    $order = Invoke-Api POST "/api/v1/payments/orders" @{ appointment_id = $appointmentID } $PatientToken
    Assert-Equal 200 $order.Status "payment order create"
    return [pscustomobject]@{
        SlotID = $SlotID; AppointmentID = $appointmentID; OrderID = [string]$order.Json.data.order_id
        Amount = [long]$order.Json.data.gross_amount; ExpiresAt = [long]$order.Json.data.expires_at
        Lock = $lock; Appointment = $appointment; Order = $order
    }
}

function Wait-Outbox([string]$OrderID, [string]$Status, [int]$TimeoutSeconds = 45) {
    return Wait-Sql "payment_db" "SELECT status FROM payment_outbox_events WHERE aggregate_id='$OrderID' ORDER BY created_at DESC LIMIT 1;" $Status $TimeoutSeconds
}

function Get-OrderRow([string]$OrderID) {
    return Invoke-Sql "payment_db" "SELECT status || '|' || gateway_capture_status || '|' || fulfillment_status || '|' || gross_amount || '|' || commission_amount || '|' || net_amount FROM payment_orders WHERE id='$OrderID';"
}

. (Join-Path $PSScriptRoot "v4\common.ps1")

try {
    Write-Step "Starting a fresh isolated Docker environment"
    Push-Location $script:RepoRoot
    try {
        Invoke-Compose down -v --remove-orphans
        Invoke-Compose up -d --build
    } finally { Pop-Location }
    Wait-Http "$script:Gateway/api/v1/public/booking/templates" 180
    Wait-Http "$script:Gateway/api/v1/auth/public-key" 180

    Write-Step "Applying and rerunning explicit SQL migrations"
    $bookingMigrations = @("migrate_v0_5_ab.sql", "migrate_v0_5_cd.sql", "migrate_v0_5_ef.sql")
    $paymentMigrations = @("migrate_v1.sql", "migrate_v2.sql", "migrate_v3.sql")
    foreach ($pass in 1..2) {
        foreach ($file in $bookingMigrations) { Invoke-Migration "booking_db" (Join-Path $script:BackendRoot "booking-service\$file") }
        foreach ($file in $paymentMigrations) { Invoke-Migration "payment_db" (Join-Path $script:BackendRoot "payment-service\$file") }
    }
    Assert-Equal "expires_at" (Invoke-Sql "payment_db" "SELECT column_name FROM information_schema.columns WHERE table_name='payment_orders' AND column_name='expires_at';") "payment expiry column"
    Assert-Equal "ux_payment_orders_active_pending_appointment" (Invoke-Sql "payment_db" "SELECT indexname FROM pg_indexes WHERE indexname='ux_payment_orders_active_pending_appointment';") "pending order index"
    Assert-Equal "ux_payment_orders_success_appointment" (Invoke-Sql "payment_db" "SELECT indexname FROM pg_indexes WHERE indexname='ux_payment_orders_success_appointment';") "success order index"
    Assert-Equal "ix_payment_outbox_events_delivery_poll" (Invoke-Sql "payment_db" "SELECT indexname FROM pg_indexes WHERE indexname='ix_payment_outbox_events_delivery_poll';") "outbox polling index"
    Assert-Equal "ux_payment_compensation_cases_order_reason" (Invoke-Sql "payment_db" "SELECT indexname FROM pg_indexes WHERE indexname='ux_payment_compensation_cases_order_reason';") "compensation idempotency index"
    Assert-Equal "price" (Invoke-Sql "booking_db" "SELECT column_name FROM information_schema.columns WHERE table_name='Booking_Config_Availability' AND column_name='price';") "availability price column"
    Assert-Equal "availability_id" (Invoke-Sql "booking_db" "SELECT column_name FROM information_schema.columns WHERE table_name='Booking_Expert_Slots' AND column_name='availability_id';") "slot provenance column"
    $script:Results.Migrations = "fresh database + idempotent rerun passed"

    Write-Step "Authenticating ADMIN, EXPERT, and PATIENT through Kong"
    $admin = Login "admin@mindcare.com" "admin@mindcare.com"
    $expert = Login "expert@mindcare.com" "expert@mindcare.com"
    $patient = Login "patient@mindcare.com" "patient@mindcare.com"
    Assert-Equal "ADMIN" $admin.Role "admin role"
    Assert-Equal "EXPERT" $expert.Role "expert role"
    Assert-Equal "PATIENT" $patient.Role "patient role"
    $unauthorized = Invoke-Api POST "/api/v1/payments/orders" @{ appointment_id = "00000000-0000-0000-0000-000000000000" }
    Assert-Equal 401 $unauthorized.Status "payment order without JWT"
    $wrongRole = Invoke-Api POST "/api/v1/booking/templates" @{ shift_name = "forbidden"; start_time = "01:00"; end_time = "02:00"; slot_duration_minutes = 30 } $patient.Token
    Assert-Equal 403 $wrongRole.Status "patient cannot create template"
    $script:Results.Authentication = [ordered]@{ admin = $admin.Sanitized; expert = $expert.Sanitized; patient = $patient.Sanitized; unauthorized_status = 401; wrong_role_status = 403 }

    Write-Step "Creating deterministic priced availability and generated slots"
    $targetDate = (Get-Date).Date.AddDays(2)
    $dayOfWeek = [int]$targetDate.DayOfWeek
    if ($dayOfWeek -eq 0) { $dayOfWeek = 7 }
    $template = Invoke-Api POST "/api/v1/booking/templates" @{
        shift_name = "V4 deterministic full day"
        start_time = "06:00"
        end_time = "22:00"
        slot_duration_minutes = 30
    } $admin.Token
    Assert-Equal 200 $template.Status "template create"
    $templateID = [string]$template.Json.data.TemplateID
    $effectiveFrom = [DateTimeOffset]::new((Get-Date).Date).ToUnixTimeMilliseconds()
    $availability = Invoke-Api POST "/api/v1/booking/availabilities" @{
        template_id = $templateID
        day_of_week = $dayOfWeek
        effective_from = $effectiveFrom
        price = 300000
    } $expert.Token
    Assert-Equal 200 $availability.Status "availability create"
    $availabilityID = [string]$availability.Json.data.availability_id
    $overlap = Invoke-Api POST "/api/v1/booking/availabilities" @{
        template_id = $templateID
        day_of_week = $dayOfWeek
        effective_from = $effectiveFrom
        price = 300000
    } $expert.Token
    Assert-Equal 409 $overlap.Status "overlapping availability"
    $generated = Invoke-Api POST "/api/v1/booking/slots/generate" @{ days_to_generate = 30 } $expert.Token
    Assert-Equal 200 $generated.Status "slot generation"
    Assert-True ([int]$generated.Json.data.slots_created -ge 32) "generation must create a full target day"
    $idempotentGeneration = Invoke-Api POST "/api/v1/booking/slots/generate" @{ days_to_generate = 30 } $expert.Token
    Assert-Equal 200 $idempotentGeneration.Status "idempotent generation request"
    Assert-Equal 0 $idempotentGeneration.Json.data.slots_created "second generation creates no duplicate slots"
    $targetDateText = $targetDate.ToString("yyyy-MM-dd")
    $slots = @(Get-AvailableSlots $expert.ID $targetDateText)
    Assert-True ($slots.Count -ge 24) "enough generated slots for V4 scenarios"
    $firstSlotID = [string]$slots[0].slot_id
    Assert-Equal "$($expert.ID)|$availabilityID|300000.00|0" (Invoke-Sql "booking_db" "SELECT expert_id || '|' || availability_id || '|' || price || '|' || status FROM `"Booking_Expert_Slots`" WHERE slot_id='$firstSlotID';") "generated slot provenance, price, and status"
    $script:slotIndex = 0
    function Take-Slot {
        if ($script:slotIndex -ge $slots.Count) { throw "No generated V4 slots remain" }
        $value = [string]$slots[$script:slotIndex].slot_id
        $script:slotIndex++
        return $value
    }
    $script:Results.Fixture = [ordered]@{ target_date = $targetDateText; template_id = $templateID; availability_id = $availabilityID; generated = [int]$generated.Json.data.slots_created; price_vnd = 300000 }

    Write-Step "Running the complete successful booking and payment flow"
    $happy = New-AppointmentAndOrder (Take-Slot) $expert.ID $patient.Token
    Assert-Equal "$($patient.ID)|$($expert.ID)|0|300000.00" (Invoke-Sql "booking_db" "SELECT patient_id || '|' || expert_id || '|' || status || '|' || (SELECT price FROM `"Booking_Expert_Slots`" WHERE slot_id='$($happy.SlotID)') FROM `"Booking_Appointments`" WHERE appointment_id='$($happy.AppointmentID)';") "appointment authority and initial status"
    $lockState = Invoke-Sql "booking_db" "SELECT status || '|' || locked_by || '|' || (locked_expires_at IS NOT NULL) FROM `"Booking_Expert_Slots`" WHERE slot_id='$($happy.SlotID)';"
    Assert-Equal "1|$($patient.ID)|true" $lockState "slot lock persistence"
    Assert-Equal "1|PENDING|PENDING|300000|45000|255000" (Get-OrderRow $happy.OrderID) "pending payment order values"
    Assert-True ($happy.ExpiresAt -le [long](Invoke-Sql "booking_db" "SELECT locked_expires_at FROM `"Booking_Expert_Slots`" WHERE slot_id='$($happy.SlotID)';")) "payment expiry cannot exceed booking lock"
    $expiryMatch = [regex]::Match([string]$happy.Order.Json.data.payment_url, "vnp_ExpireDate=([0-9]{14})")
    Assert-True $expiryMatch.Success "VNPay URL contains vnp_ExpireDate"
    $vnZone = [TimeZoneInfo]::FindSystemTimeZoneById("SE Asia Standard Time")
    $expectedExpiry = [TimeZoneInfo]::ConvertTimeFromUtc([DateTimeOffset]::FromUnixTimeMilliseconds($happy.ExpiresAt).UtcDateTime, $vnZone).ToString("yyyyMMddHHmmss")
    Assert-Equal $expectedExpiry $expiryMatch.Groups[1].Value "VNPay expiry equals persisted expiry"
    $retryOrder = Invoke-Api POST "/api/v1/payments/orders" @{ appointment_id = $happy.AppointmentID } $patient.Token
    Assert-Equal 200 $retryOrder.Status "create order retry"
    Assert-Equal $happy.OrderID $retryOrder.Json.data.order_id "active order ID reused"
    Assert-Equal $happy.ExpiresAt $retryOrder.Json.data.expires_at "active order expiry reused"
    Assert-Equal 1 (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_orders WHERE appointment_id='$($happy.AppointmentID)' AND status=1;") "one pending order"
    $happyTxn = "92000001"
    $ipn = Send-IPN $happy.OrderID $happy.Amount "00" $happyTxn
    Assert-Equal 200 $ipn.Status "success IPN HTTP status"
    Assert-Equal "00" $ipn.Json.RspCode "success IPN response code"
    Wait-Outbox $happy.OrderID "DELIVERED" | Out-Null
    Assert-Equal "2|CAPTURED|BOOKING_CONFIRMED|300000|45000|255000" (Get-OrderRow $happy.OrderID) "successful payment and fulfillment"
    Assert-Equal "1" (Invoke-Sql "booking_db" "SELECT status FROM `"Booking_Appointments`" WHERE appointment_id='$($happy.AppointmentID)';") "appointment confirmed"
    Assert-Equal "2||" (Invoke-Sql "booking_db" "SELECT status || '|' || COALESCE(locked_by::text,'') || '|' || COALESCE(locked_expires_at::text,'') FROM `"Booking_Expert_Slots`" WHERE slot_id='$($happy.SlotID)';") "slot occupied and lock metadata cleared"
    Assert-Equal "255000" (Invoke-Sql "payment_db" "SELECT pending_balance FROM payment_wallets WHERE user_id='$($expert.ID)';") "expert pending balance receives net amount"
    Assert-Equal "2" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_wallet_transactions WHERE reference_id='$($happy.OrderID)';") "two wallet ledger rows"
    Assert-Equal "1" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_outbox_events WHERE aggregate_id='$($happy.OrderID)' AND event_type='booking.appointment.confirm';") "one confirmation outbox"
    $confirmedAt = Invoke-Sql "booking_db" "SELECT confirmed_at FROM `"Booking_Appointments`" WHERE appointment_id='$($happy.AppointmentID)';"
    $replay = Send-IPN $happy.OrderID $happy.Amount "00" $happyTxn
    Assert-Equal "02" $replay.Json.RspCode "duplicate success acknowledged"
    Assert-Equal "2" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_wallet_transactions WHERE reference_id='$($happy.OrderID)';") "replay creates no wallet rows"
    Assert-Equal $confirmedAt (Invoke-Sql "booking_db" "SELECT confirmed_at FROM `"Booking_Appointments`" WHERE appointment_id='$($happy.AppointmentID)';") "replay does not rewrite confirmed_at"
    Assert-Equal "0" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_compensation_cases WHERE payment_order_id='$($happy.OrderID)';") "happy flow has no compensation"
    $script:Results.Happy = "PASS"

    Write-Step "Running failed payment and booking release flow"
    $failed = New-AppointmentAndOrder (Take-Slot) $expert.ID $patient.Token
    $failedIPN = Send-IPN $failed.OrderID $failed.Amount "24" "92000002"
    Assert-Equal "00" $failedIPN.Json.RspCode "failed gateway result accepted"
    Wait-Outbox $failed.OrderID "DELIVERED" | Out-Null
    Assert-Equal "3|FAILED|BOOKING_FAILED|300000|45000|255000" (Get-OrderRow $failed.OrderID) "failed payment state"
    Assert-Equal "2" (Invoke-Sql "booking_db" "SELECT status FROM `"Booking_Appointments`" WHERE appointment_id='$($failed.AppointmentID)';") "failed payment cancels appointment"
    Assert-Equal "0" (Invoke-Sql "booking_db" "SELECT status FROM `"Booking_Expert_Slots`" WHERE slot_id='$($failed.SlotID)';") "failed payment releases slot"
    Assert-Equal "0" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_wallet_transactions WHERE reference_id='$($failed.OrderID)';") "failed payment does not settle wallet"
    $failedReplay = Send-IPN $failed.OrderID $failed.Amount "24" "92000002"
    Assert-Equal "02" $failedReplay.Json.RspCode "failed IPN replay acknowledged"
    Assert-Equal "1" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_outbox_events WHERE aggregate_id='$($failed.OrderID)';") "failed replay creates no outbox duplicate"
    $script:Results.Failed = "PASS"

    Write-Step "Running time-off-aware failed payment flow"
    $timeOffFailure = New-AppointmentAndOrder (Take-Slot) $expert.ID $patient.Token
    $slotWindow = (Invoke-Sql "booking_db" "SELECT start_time || '|' || end_time FROM `"Booking_Expert_Slots`" WHERE slot_id='$($timeOffFailure.SlotID)';").Split('|')
    $forceTimeOff = Invoke-Api POST "/api/v1/booking/time-off/confirm" @{
        start_datetime = [long]$slotWindow[0]
        end_datetime = [long]$slotWindow[1]
        reason = "V4 pending-payment coverage"
    } $expert.Token
    Assert-Equal 200 $forceTimeOff.Status "force time-off for pending appointment"
    $timeOffID = [string]$forceTimeOff.Json.data.time_off.time_off_id
    Assert-Equal "2|3" (Invoke-Sql "booking_db" "SELECT appointment.status || '|' || slot.status FROM `"Booking_Appointments`" appointment JOIN `"Booking_Expert_Slots`" slot ON slot.slot_id=appointment.slot_id WHERE appointment.appointment_id='$($timeOffFailure.AppointmentID)';") "time-off cancels pending appointment and makes slot unavailable"
    $timeOffFailedIPN = Send-IPN $timeOffFailure.OrderID $timeOffFailure.Amount "24" "92000003"
    Assert-Equal "00" $timeOffFailedIPN.Json.RspCode "time-off failure IPN accepted"
    Wait-Outbox $timeOffFailure.OrderID "DELIVERED" | Out-Null
    Assert-Equal "2|3" (Invoke-Sql "booking_db" "SELECT appointment.status || '|' || slot.status FROM `"Booking_Appointments`" appointment JOIN `"Booking_Expert_Slots`" slot ON slot.slot_id=appointment.slot_id WHERE appointment.appointment_id='$($timeOffFailure.AppointmentID)';") "failure callback preserves unavailable coverage"
    $coveredLock = Invoke-Api POST "/api/v1/booking/slots/$($timeOffFailure.SlotID)/lock" $null $patient.Token
    Assert-True ($coveredLock.Status -ne 200) "time-off-covered slot cannot be locked"
    $script:Results.TimeOffFailure = "PASS"

    Write-Step "Running expiry, replacement, and old/new successful capture flow"
    $replacement = New-AppointmentAndOrder (Take-Slot) $expert.ID $patient.Token
    $orderA = $replacement.OrderID
    $exactNow = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()
    [void](Invoke-Sql "payment_db" "UPDATE payment_orders SET expires_at=$exactNow WHERE id='$orderA';")
    $orderBResponse = Invoke-Api POST "/api/v1/payments/orders" @{ appointment_id = $replacement.AppointmentID } $patient.Token
    Assert-Equal 200 $orderBResponse.Status "replacement order create"
    $orderB = [string]$orderBResponse.Json.data.order_id
    Assert-True ($orderA -ne $orderB) "replacement has a new ID and transaction reference"
    Assert-Equal "4|1" (Invoke-Sql "payment_db" "SELECT STRING_AGG(status::text,'|' ORDER BY created_at) FROM payment_orders WHERE appointment_id='$($replacement.AppointmentID)';") "expired A and pending B"
    Assert-Equal "1" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_orders WHERE appointment_id='$($replacement.AppointmentID)' AND status=1;") "one replacement pending"
    $oldSuccess = Send-IPN $orderA $replacement.Amount "00" "92000004"
    Assert-Equal "00" $oldSuccess.Json.RspCode "expired old order success accepted"
    Wait-Outbox $orderA "DELIVERED" | Out-Null
    Assert-Equal "2|4" (Invoke-Sql "payment_db" "SELECT STRING_AGG(status::text,'|' ORDER BY created_at) FROM payment_orders WHERE appointment_id='$($replacement.AppointmentID)';") "old success wins and replacement expires"
    $walletRowsBeforeDuplicate = Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_wallet_transactions WHERE reference_id IN ('$orderA','$orderB');"
    $newCapture = Send-IPN $orderB $replacement.Amount "00" "92000005"
    Assert-Equal "02" $newCapture.Json.RspCode "duplicate capture receives stable acknowledgement"
    Assert-Equal "4|CAPTURED_DUPLICATE|REFUND_REQUIRED|300000|45000|255000" (Get-OrderRow $orderB) "duplicate capture evidence"
    Assert-Equal $walletRowsBeforeDuplicate (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_wallet_transactions WHERE reference_id IN ('$orderA','$orderB');") "duplicate capture does not settle wallet"
    Assert-Equal "1" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_outbox_events WHERE aggregate_id IN ('$orderA','$orderB') AND event_type='booking.appointment.confirm';") "duplicate capture creates no second booking outbox"
    Assert-Equal "1" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_compensation_cases WHERE payment_order_id='$orderB' AND status='REFUND_REQUIRED' AND reason_code='DUPLICATE_GATEWAY_CAPTURE';") "one refund-required duplicate case"
    $evidenceBefore = Invoke-Sql "payment_db" "SELECT paid_at || '|' || gateway_txn_ref || '|' || gateway_payment_date FROM payment_orders WHERE id='$orderB';"
    $newCaptureReplay = Send-IPN $orderB $replacement.Amount "00" "92000005"
    Assert-Equal "02" $newCaptureReplay.Json.RspCode "duplicate capture replay acknowledged"
    Assert-Equal $evidenceBefore (Invoke-Sql "payment_db" "SELECT paid_at || '|' || gateway_txn_ref || '|' || gateway_payment_date FROM payment_orders WHERE id='$orderB';") "duplicate capture evidence remains stable"
    Assert-Equal "1" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_compensation_cases WHERE payment_order_id='$orderB';") "duplicate capture case idempotency"
    $script:Results.ExpiryReplacement = "PASS"
    $script:Results.DuplicateCapture = "PASS"

    Write-Step "Running invalid VNPay IPN matrix"
    $invalid = New-AppointmentAndOrder (Take-Slot) $expert.ID $patient.Token
    $baselineMutationCounts = Invoke-Sql "payment_db" "SELECT (SELECT COUNT(*) FROM payment_wallet_transactions) || '|' || (SELECT COUNT(*) FROM payment_outbox_events) || '|' || (SELECT COUNT(*) FROM payment_compensation_cases);"
    $invalidResults = [ordered]@{}
    $noHash = Invoke-Api GET "$script:Gateway/api/v1/payments/vnpay-ipn?vnp_TxnRef=$($invalid.OrderID)&vnp_Amount=$($invalid.Amount * 100)&vnp_ResponseCode=00"
    $invalidResults.no_checksum = $noHash.Json.RspCode
    $badHash = Invoke-Api GET "$script:Gateway/api/v1/payments/vnpay-ipn?vnp_TxnRef=$($invalid.OrderID)&vnp_Amount=$($invalid.Amount * 100)&vnp_ResponseCode=00&vnp_SecureHash=bad"
    $invalidResults.invalid_checksum = $badHash.Json.RspCode
    $missingResponse = Invoke-Api GET (New-SignedIPN -OrderID $invalid.OrderID -AmountVND $invalid.Amount -OmitResponseCode)
    $invalidResults.missing_response_code = $missingResponse.Json.RspCode
    $missingAmount = Invoke-Api GET (New-SignedIPN -OrderID $invalid.OrderID -AmountVND $invalid.Amount -OmitAmount)
    $invalidResults.missing_amount = $missingAmount.Json.RspCode
    foreach ($case in @(@("nonnumeric_amount", "abc"), @("zero_amount", "0"), @("negative_amount", "-100"), @("amount_mismatch", "100"))) {
        $response = Invoke-Api GET (New-SignedIPN -OrderID $invalid.OrderID -AmountVND $invalid.Amount -RawAmount $case[1])
        $invalidResults[$case[0]] = $response.Json.RspCode
        Assert-Equal "04" $response.Json.RspCode $case[0]
    }
    $unknownID = [guid]::NewGuid().ToString()
    $unknown = Invoke-Api GET (New-SignedIPN -OrderID $unknownID -AmountVND $invalid.Amount)
    $invalidResults.unknown_txn_reference = $unknown.Json.RspCode
    Assert-Equal "97" $noHash.Json.RspCode "missing checksum contract"
    Assert-Equal "97" $badHash.Json.RspCode "invalid checksum contract"
    Assert-Equal "99" $missingResponse.Json.RspCode "missing response code contract"
    Assert-Equal "04" $missingAmount.Json.RspCode "missing amount contract"
    Assert-Equal "01" $unknown.Json.RspCode "unknown order contract"
    Assert-Equal "1|PENDING|PENDING|300000|45000|255000" (Get-OrderRow $invalid.OrderID) "invalid IPNs leave order unchanged"
    Assert-Equal $baselineMutationCounts (Invoke-Sql "payment_db" "SELECT (SELECT COUNT(*) FROM payment_wallet_transactions) || '|' || (SELECT COUNT(*) FROM payment_outbox_events) || '|' || (SELECT COUNT(*) FROM payment_compensation_cases);") "invalid IPNs create no wallet, outbox, or compensation mutation"
    $script:Results.InvalidIPN = $invalidResults

    Write-Step "Running durable outbox retry across payment-service restart"
    $retryFlow = New-AppointmentAndOrder (Take-Slot) $expert.ID $patient.Token
    Push-Location $script:RepoRoot
    try { Invoke-Compose stop booking-service } finally { Pop-Location }
    $retryIPN = Send-IPN $retryFlow.OrderID $retryFlow.Amount "00" "92000006"
    Assert-Equal "00" $retryIPN.Json.RspCode "retry scenario success IPN"
    Wait-Sql "payment_db" "SELECT status FROM payment_outbox_events WHERE aggregate_id='$($retryFlow.OrderID)';" "RETRY_WAIT" 30 | Out-Null
    $retryMetadata = Invoke-Sql "payment_db" "SELECT attempt_count || '|' || (next_attempt_at IS NOT NULL) || '|' || (last_error IS NOT NULL AND LENGTH(last_error) <= 500) FROM payment_outbox_events WHERE aggregate_id='$($retryFlow.OrderID)';"
    Assert-True ($retryMetadata -match '^[1-9][0-9]*\|true\|true$') "retry metadata is persisted and bounded"
    Push-Location $script:RepoRoot
    try {
        Invoke-Compose restart payment-service
        Invoke-Compose start booking-service
        Invoke-Compose restart api-gateway
    } finally { Pop-Location }
    Wait-Http "$script:Gateway/api/v1/public/booking/templates" 120
    Wait-Outbox $retryFlow.OrderID "DELIVERED" 60 | Out-Null
    Assert-Equal "BOOKING_CONFIRMED" (Invoke-Sql "payment_db" "SELECT fulfillment_status FROM payment_orders WHERE id='$($retryFlow.OrderID)';") "retry delivery confirms fulfillment"
    Assert-Equal "1" (Invoke-Sql "booking_db" "SELECT status FROM `"Booking_Appointments`" WHERE appointment_id='$($retryFlow.AppointmentID)';") "retry delivery confirms booking once"
    $script:Results.OutboxRetry = "PASS ($retryMetadata before restart)"

    Write-Step "Running controlled crash/idempotent redelivery"
    $redeliveryEvent = Invoke-Sql "payment_db" "SELECT id FROM payment_outbox_events WHERE aggregate_id='$($happy.OrderID)' AND event_type='booking.appointment.confirm';"
    $redeliveryAttempts = [int](Invoke-Sql "payment_db" "SELECT attempt_count FROM payment_outbox_events WHERE id='$redeliveryEvent';")
    $redeliveryConfirmedAt = Invoke-Sql "booking_db" "SELECT confirmed_at FROM `"Booking_Appointments`" WHERE appointment_id='$($happy.AppointmentID)';"
    [void](Invoke-Sql "payment_db" "UPDATE payment_outbox_events SET status='PENDING', published=false, delivered_at=NULL, next_attempt_at=NULL WHERE id='$redeliveryEvent';")
    Wait-Outbox $happy.OrderID "DELIVERED" 30 | Out-Null
    Assert-Equal ($redeliveryAttempts + 1) (Invoke-Sql "payment_db" "SELECT attempt_count FROM payment_outbox_events WHERE id='$redeliveryEvent';") "redelivery records exactly one more attempt"
    Assert-Equal $redeliveryConfirmedAt (Invoke-Sql "booking_db" "SELECT confirmed_at FROM `"Booking_Appointments`" WHERE appointment_id='$($happy.AppointmentID)';") "idempotent booking no-op preserves confirmed_at"
    Assert-Equal "2" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_wallet_transactions WHERE reference_id='$($happy.OrderID)' AND type IN (1,2);") "redelivery does not duplicate credit or commission debit"
    $script:Results.CrashRedelivery = "PASS (controlled lost-local-result fixture)"

    Write-Step "Running permanent booking conflict compensation"
    $conflict = New-AppointmentAndOrder (Take-Slot) $expert.ID $patient.Token
    $cancel = Invoke-Api PATCH "/api/v1/booking/appointments/$($conflict.AppointmentID)/cancel" @{ reason = "V4 conflict fixture" } $patient.Token
    Assert-Equal 200 $cancel.Status "patient cancellation before gateway success"
    $conflictIPN = Send-IPN $conflict.OrderID $conflict.Amount "00" "92000007"
    Assert-Equal "00" $conflictIPN.Json.RspCode "conflict payment success accepted"
    Wait-Outbox $conflict.OrderID "DEAD" 30 | Out-Null
    Assert-Equal "conflict" (Invoke-Sql "payment_db" "SELECT terminal_reason_code FROM payment_outbox_events WHERE aggregate_id='$($conflict.OrderID)';") "typed conflict reason"
    Assert-Equal "2|CAPTURED|REFUND_REQUIRED|300000|45000|255000" (Get-OrderRow $conflict.OrderID) "payment truth retained with refund required"
    Assert-Equal "1" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_compensation_cases WHERE payment_order_id='$($conflict.OrderID)' AND status='REFUND_REQUIRED' AND reason_code='BOOKING_CONFLICT';") "one booking conflict case"
    $deadAttemptCount = Invoke-Sql "payment_db" "SELECT attempt_count FROM payment_outbox_events WHERE aggregate_id='$($conflict.OrderID)';"
    Start-Sleep -Seconds 7
    Assert-Equal $deadAttemptCount (Invoke-Sql "payment_db" "SELECT attempt_count FROM payment_outbox_events WHERE aggregate_id='$($conflict.OrderID)';") "DEAD conflict is not retried"
    $script:Results.Conflict = "PASS"

    Write-Step "Running authenticated wrong-service callback compensation"
    $authFailure = New-AppointmentAndOrder (Take-Slot) $expert.ID $patient.Token
    Push-Location $script:RepoRoot
    try { Invoke-Compose stop booking-service } finally { Pop-Location }
    $authIPN = Send-IPN $authFailure.OrderID $authFailure.Amount "00" "92000008"
    Assert-Equal "00" $authIPN.Json.RspCode "auth failure payment success accepted"
    Wait-Sql "payment_db" "SELECT status FROM payment_outbox_events WHERE aggregate_id='$($authFailure.OrderID)';" "RETRY_WAIT" 30 | Out-Null
    Push-Location $script:RepoRoot
    try {
        Invoke-ComposeWithOverride up -d --force-recreate payment-service
        Invoke-Compose start booking-service
        Invoke-Compose restart api-gateway
    } finally { Pop-Location }
    Wait-Http "$script:Gateway/api/v1/public/booking/templates" 120
    [void](Invoke-Sql "payment_db" "UPDATE payment_outbox_events SET next_attempt_at=0 WHERE aggregate_id='$($authFailure.OrderID)' AND status='RETRY_WAIT';")
    Wait-Outbox $authFailure.OrderID "DEAD" 45 | Out-Null
    Assert-Equal "authentication" (Invoke-Sql "payment_db" "SELECT terminal_reason_code FROM payment_outbox_events WHERE aggregate_id='$($authFailure.OrderID)';") "typed M2M authentication failure"
    Assert-Equal "2|CAPTURED|MANUAL_REVIEW|300000|45000|255000" (Get-OrderRow $authFailure.OrderID) "authentication failure requires manual review"
    Assert-Equal "1" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_compensation_cases WHERE payment_order_id='$($authFailure.OrderID)' AND status='MANUAL_REVIEW' AND reason_code='BOOKING_AUTHENTICATION_REJECTED';") "one M2M compensation case"
    Push-Location $script:RepoRoot
    try {
        Invoke-Compose up -d --force-recreate payment-service
        Invoke-Compose restart api-gateway
    } finally { Pop-Location }
    Wait-Http "$script:Gateway/api/v1/payments/vnpay-ipn?x=1" 120
    $script:Results.AuthenticationFailure = "PASS"

    Write-Step "Running schedule reconciliation and time-off lifecycle checks"
    $protectedPrice = Invoke-Sql "booking_db" "SELECT price FROM `"Booking_Expert_Slots`" WHERE slot_id='$($happy.SlotID)';"
    $availabilityUpdate = Invoke-Api PATCH "/api/v1/booking/availabilities/$availabilityID" @{ price = 310000 } $expert.Token
    Assert-Equal 200 $availabilityUpdate.Status "availability price update"
    Assert-Equal $protectedPrice (Invoke-Sql "booking_db" "SELECT price FROM `"Booking_Expert_Slots`" WHERE slot_id='$($happy.SlotID)';") "appointment-linked occupied slot price snapshot protected"
    Assert-Equal "0" (Invoke-Sql "booking_db" "SELECT COUNT(*) FROM `"Booking_Expert_Slots`" slot WHERE availability_id='$availabilityID' AND status=0 AND start_time > $([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()) AND price <> 310000 AND NOT EXISTS (SELECT 1 FROM `"Booking_Appointments`" appointment WHERE appointment.slot_id=slot.slot_id);") "future safe available slots reconciled to authoritative price"
    $timeOffSlot = (Get-AvailableSlots $expert.ID $targetDateText | Select-Object -First 1)
    Assert-True ($null -ne $timeOffSlot) "available slot remains for time-off lifecycle"
    $expiredCoveredLock = Invoke-Api POST "/api/v1/booking/slots/$($timeOffSlot.slot_id)/lock" $null $patient.Token
    Assert-Equal 200 $expiredCoveredLock.Status "appointment-free slot locked before time-off"
    [void](Invoke-Sql "booking_db" "UPDATE `"Booking_Expert_Slots`" SET locked_expires_at=$([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds() - 1000) WHERE slot_id='$($timeOffSlot.slot_id)' AND status=1;")
    Assert-Equal "1|true" (Invoke-Sql "booking_db" "SELECT status || '|' || (locked_expires_at < $([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds())) FROM `"Booking_Expert_Slots`" WHERE slot_id='$($timeOffSlot.slot_id)';") "controlled standalone lock is expired"
    $normalTimeOff = Invoke-Api POST "/api/v1/booking/time-off" @{
        start_datetime = [long]$timeOffSlot.start_time
        end_datetime = [long]$timeOffSlot.end_time
        reason = "V4 public coverage"
    } $expert.Token
    Assert-Equal 200 $normalTimeOff.Status "normal time-off create"
    $normalTimeOffID = [string]$normalTimeOff.Json.data.time_off.time_off_id
    Assert-Equal "3|true|true" (Invoke-Sql "booking_db" "SELECT status || '|' || (locked_by IS NULL AND locked_expires_at IS NULL) || '|' || (EXISTS(SELECT 1 FROM `"Booking_Expert_Time_Off`" WHERE time_off_id='$normalTimeOffID' AND processed_at IS NOT NULL)) FROM `"Booking_Expert_Slots`" WHERE slot_id='$($timeOffSlot.slot_id)';") "expired covered lock becomes unavailable with lock metadata cleared"
    $hidden = @(Get-AvailableSlots $expert.ID $targetDateText | Where-Object { $_.slot_id -eq $timeOffSlot.slot_id })
    Assert-Equal 0 $hidden.Count "active time-off hides public slot"
    $directCoveredLock = Invoke-Api POST "/api/v1/booking/slots/$($timeOffSlot.slot_id)/lock" $null $patient.Token
    Assert-True ($directCoveredLock.Status -ne 200) "active time-off rejects direct lock"
    $deleteTimeOff = Invoke-Api DELETE "/api/v1/booking/time-off/$normalTimeOffID" $null $expert.Token
    Assert-Equal 200 $deleteTimeOff.Status "time-off delete"
    $regenerate = Invoke-Api POST "/api/v1/booking/slots/generate" @{ days_to_generate = 30 } $expert.Token
    Assert-Equal 200 $regenerate.Status "safe generation after time-off delete"
    $timeOffList = Invoke-Api GET "/api/v1/booking/time-off" $null $expert.Token
    Assert-Equal 200 $timeOffList.Status "time-off list"
    $script:Results.ScheduleTimeOff = "PASS"

    Write-Step "Running real PostgreSQL concurrency checks"
    $slots = @(Get-AvailableSlots $expert.ID $targetDateText)
    $script:slotIndex = 0
    $concurrentLockSlot = Take-Slot
    $gateway = $script:Gateway
    $patientToken = $patient.Token
    $lockJobs = 1..2 | ForEach-Object {
        Start-Job -ScriptBlock {
            param($gateway, $token, $slot)
            try {
                $r = Invoke-WebRequest -UseBasicParsing -Method Post -Uri "$gateway/api/v1/booking/slots/$slot/lock" -Headers @{ Authorization = "Bearer $token" }
                [int]$r.StatusCode
            } catch { [int]$_.Exception.Response.StatusCode }
        } -ArgumentList $gateway, $patientToken, $concurrentLockSlot
    }
    try { $lockStatuses = @($lockJobs | Wait-Job | Receive-Job) } finally { $lockJobs | Remove-Job -Force }
    Assert-Equal 1 (@($lockStatuses | Where-Object { $_ -eq 200 })).Count "exactly one concurrent slot lock succeeds"
    Assert-Equal 1 (Invoke-Sql "booking_db" "SELECT COUNT(*) FROM `"Booking_Expert_Slots`" WHERE slot_id='$concurrentLockSlot' AND status=1;") "slot is locked once"

    $concurrentCreate = New-AppointmentAndOrder (Take-Slot) $expert.ID $patient.Token -NoOrder
    $appointmentIDForConcurrency = $concurrentCreate.AppointmentID
    $createJobs = 1..2 | ForEach-Object {
        Start-Job -ScriptBlock {
            param($gateway, $token, $appointment)
            $body = @{ appointment_id = $appointment } | ConvertTo-Json -Compress
            try {
                $r = Invoke-WebRequest -UseBasicParsing -Method Post -Uri "$gateway/api/v1/payments/orders" -Headers @{ Authorization = "Bearer $token" } -ContentType "application/json" -Body $body
                $j = $r.Content | ConvertFrom-Json
                "200|$($j.data.order_id)"
            } catch { "$([int]$_.Exception.Response.StatusCode)|" }
        } -ArgumentList $gateway, $patientToken, $appointmentIDForConcurrency
    }
    try { $createResults = @($createJobs | Wait-Job | Receive-Job) } finally { $createJobs | Remove-Job -Force }
    Assert-True (@($createResults | Where-Object { $_ -match '^200\|' }).Count -eq 2) "both concurrent create-order callers receive the winner"
    Assert-Equal 1 (@($createResults | ForEach-Object { ($_ -split '\|')[1] } | Select-Object -Unique)).Count "concurrent callers receive one order ID"
    Assert-Equal "1" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_orders WHERE appointment_id='$appointmentIDForConcurrency' AND status=1;") "one concurrent PENDING row"

    $concurrentCapture = New-AppointmentAndOrder (Take-Slot) $expert.ID $patient.Token
    $concurrentA = $concurrentCapture.OrderID
    [void](Invoke-Sql "payment_db" "UPDATE payment_orders SET expires_at=$([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()) WHERE id='$concurrentA';")
    $concurrentBResponse = Invoke-Api POST "/api/v1/payments/orders" @{ appointment_id = $concurrentCapture.AppointmentID } $patient.Token
    Assert-Equal 200 $concurrentBResponse.Status "concurrent capture replacement"
    $concurrentB = [string]$concurrentBResponse.Json.data.order_id
    $urlA = New-SignedIPN -OrderID $concurrentA -AmountVND $concurrentCapture.Amount -TransactionNumber "92000009"
    $urlB = New-SignedIPN -OrderID $concurrentB -AmountVND $concurrentCapture.Amount -TransactionNumber "92000010"
    $ipnJobs = @($urlA, $urlB) | ForEach-Object {
        Start-Job -ScriptBlock {
            param($url)
            $r = Invoke-WebRequest -UseBasicParsing -Uri $url
            ($r.Content | ConvertFrom-Json).RspCode
        } -ArgumentList $_
    }
    try { $concurrentIPNResults = @($ipnJobs | Wait-Job | Receive-Job) } finally { $ipnJobs | Remove-Job -Force }
    Assert-True (@($concurrentIPNResults | Where-Object { $_ -in @('00', '02') }).Count -eq 2) "concurrent IPNs receive stable acknowledgements"
    $winnerID = Invoke-Sql "payment_db" "SELECT id FROM payment_orders WHERE appointment_id='$($concurrentCapture.AppointmentID)' AND status=2;"
    $loserID = if ($winnerID -eq $concurrentA) { $concurrentB } else { $concurrentA }
    Wait-Outbox $winnerID "DELIVERED" 45 | Out-Null
    Assert-Equal "1" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_orders WHERE appointment_id='$($concurrentCapture.AppointmentID)' AND status=2;") "one concurrent success winner"
    Assert-Equal "2" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_wallet_transactions WHERE reference_id IN ('$concurrentA','$concurrentB');") "one wallet settlement under concurrent captures"
    Assert-Equal "1" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_outbox_events WHERE aggregate_id IN ('$concurrentA','$concurrentB') AND event_type='booking.appointment.confirm';") "one booking event under concurrent captures"
    Assert-Equal "1" (Invoke-Sql "payment_db" "SELECT COUNT(*) FROM payment_compensation_cases WHERE payment_order_id='$loserID' AND reason_code='DUPLICATE_GATEWAY_CAPTURE';") "one losing-capture compensation case"
    $script:Results.Concurrency = [ordered]@{ lock_statuses = $lockStatuses; create_order = $createResults; ipn_codes = $concurrentIPNResults; winner = $winnerID; loser = $loserID }

    Write-Step "Verifying ADMIN compensation visibility and safe fields"
    $caseList = Invoke-Api GET "/api/v1/payments/compensation-cases?status=REFUND_REQUIRED&page=0&size=100" $null $admin.Token
    Assert-Equal 200 $caseList.Status "admin compensation list"
    Assert-True ([int]$caseList.Json.data.total -ge 2) "admin sees refund-required cases"
    $caseID = [string]$caseList.Json.data.items[0].id
    $caseDetail = Invoke-Api GET "/api/v1/payments/compensation-cases/$caseID" $null $admin.Token
    Assert-Equal 200 $caseDetail.Status "admin compensation detail"
    Assert-True ($caseDetail.Raw.Contains("gateway_transaction_number")) "admin detail includes trusted refund-investigation evidence"
    Assert-True (-not $caseDetail.Raw.ToLowerInvariant().Contains("securehash")) "admin response hides secure hashes"
    Assert-True (-not $caseDetail.Raw.ToLowerInvariant().Contains("signature")) "admin response hides signatures"
    Assert-True (-not $caseDetail.Raw.ToLowerInvariant().Contains("secret")) "admin response hides secrets"
    $patientCaseList = Invoke-Api GET "/api/v1/payments/compensation-cases" $null $patient.Token
    Assert-Equal 403 $patientCaseList.Status "non-admin compensation access denied"
    $script:Results.CompensationVisibility = "PASS"

    Write-Step "Verifying V4.5 booking and payment read APIs through Kong"
    $patientAppointments = Invoke-Api GET "/api/v1/booking/appointments?page=0&size=20" $null $patient.Token
    Assert-Equal 200 $patientAppointments.Status "patient appointment calendar"
    Assert-True ([int]$patientAppointments.Json.data.total_items -ge 1) "patient calendar returns owned rows"
    Assert-Equal 0 ([int]$patientAppointments.Json.data.page) "patient pagination is zero-based"
    $expertAppointments = Invoke-Api GET "/api/v1/booking/appointments/expert?from=$targetDateText&to=$targetDateText&status=CONFIRMED&page=0&size=20" $null $expert.Token
    Assert-Equal 200 $expertAppointments.Status "expert appointment calendar filters"
    Assert-True (@($expertAppointments.Json.data.items | Where-Object { $_.appointment_id -eq $happy.AppointmentID }).Count -eq 1) "expert calendar contains confirmed appointment"
    $patientDetail = Invoke-Api GET "/api/v1/booking/appointments/$($happy.AppointmentID)" $null $patient.Token
    Assert-Equal 200 $patientDetail.Status "patient appointment detail ownership"
    Assert-Equal $happy.AppointmentID ([string]$patientDetail.Json.data.appointment_id) "appointment detail identity"
    $expertDetail = Invoke-Api GET "/api/v1/booking/appointments/$($happy.AppointmentID)" $null $expert.Token
    Assert-Equal 200 $expertDetail.Status "assigned expert appointment detail"
    $expertSlots = Invoke-Api GET "/api/v1/booking/slots/expert?from=$targetDateText&to=$targetDateText&status=OCCUPIED&page=0&size=50" $null $expert.Token
    Assert-Equal 200 $expertSlots.Status "expert slot calendar filters"
    Assert-True (@($expertSlots.Json.data.items | Where-Object { $_.slot_id -eq $happy.SlotID }).Count -eq 1) "expert slot calendar contains occupied slot"
    $badBookingPage = Invoke-Api GET "/api/v1/booking/appointments?page=-1" $null $patient.Token
    Assert-Equal 400 $badBookingPage.Status "booking rejects negative page"

    $paymentList = Invoke-Api GET "/api/v1/payments/orders?appointment_id=$($happy.AppointmentID)&page=0&size=20" $null $patient.Token
    Assert-Equal 200 $paymentList.Status "patient payment order list"
    Assert-Equal 1 ([int]$paymentList.Json.data.total_items) "appointment payment list has one order"
    Assert-Equal "SUCCESS" ([string]$paymentList.Json.data.items[0].status) "payment list exposes authoritative status"
    Assert-Equal "BOOKING_CONFIRMED" ([string]$paymentList.Json.data.items[0].fulfillment_status) "payment list exposes fulfillment"
    $paymentDetail = Invoke-Api GET "/api/v1/payments/orders/$($happy.OrderID)" $null $patient.Token
    Assert-Equal 200 $paymentDetail.Status "patient payment order detail"
    Assert-Equal "CAPTURED" ([string]$paymentDetail.Json.data.gateway_capture_status) "payment detail exposes capture state"
    Assert-True (-not $paymentDetail.Raw.ToLowerInvariant().Contains("securehash")) "payment detail hides secure hash"
    $badPaymentSize = Invoke-Api GET "/api/v1/payments/orders?size=101" $null $patient.Token
    Assert-Equal 400 $badPaymentSize.Status "payment rejects oversized page"
    $walletHistory = Invoke-Api GET "/api/v1/payments/wallets/history?page=0&size=100" $null $expert.Token
    Assert-Equal 200 $walletHistory.Status "expert wallet history pagination"
    Assert-True ([int]$walletHistory.Json.data.total_items -ge 2) "wallet history contains settlement ledger"

    $topUp = Invoke-Api POST "/api/v1/payments/wallets/top-up" @{ amount = 6000000 } $expert.Token
    Assert-Equal 200 $topUp.Status "dev-only withdrawal fixture top-up"
    $bankAccount = Invoke-Api POST "/api/v1/payments/bank-accounts" @{ bank_code = "VCB"; account_number = "V45TEST001"; account_holder_name = "V4 5 TEST EXPERT" } $expert.Token
    Assert-Equal 200 $bankAccount.Status "expert links bank account"
    $bankAccountID = [string]$bankAccount.Json.data.id
    $withdrawal = Invoke-Api POST "/api/v1/payments/withdrawals" @{ bank_account_id = $bankAccountID; amount = 5000001 } $expert.Token
    Assert-Equal 200 $withdrawal.Status "expert creates pending-approval withdrawal"
    $withdrawalID = [string]$withdrawal.Json.data.id
    Assert-Equal "PENDING_APPROVAL" ([string]$withdrawal.Json.data.status) "withdrawal requires admin approval"
    $expertWithdrawals = Invoke-Api GET "/api/v1/payments/withdrawals?status=PENDING_APPROVAL&page=0&size=20" $null $expert.Token
    Assert-Equal 200 $expertWithdrawals.Status "expert withdrawal list"
    Assert-True (@($expertWithdrawals.Json.data.items | Where-Object { $_.id -eq $withdrawalID }).Count -eq 1) "expert sees owned withdrawal"
    $withdrawalDetail = Invoke-Api GET "/api/v1/payments/withdrawals/$withdrawalID" $null $expert.Token
    Assert-Equal 200 $withdrawalDetail.Status "expert withdrawal detail"
    $adminWithdrawals = Invoke-Api GET "/api/v1/payments/admin/withdrawals?status=PENDING_APPROVAL&expert_id=$($expert.ID)&page=0&size=20" $null $admin.Token
    Assert-Equal 200 $adminWithdrawals.Status "admin pending withdrawal list"
    Assert-True (@($adminWithdrawals.Json.data.items | Where-Object { $_.id -eq $withdrawalID }).Count -eq 1) "admin can discover pending withdrawal before action"
    $filteredCases = Invoke-Api GET "/api/v1/payments/compensation-cases?reason_code=BOOKING_CONFLICT&page=0&size=20" $null $admin.Token
    Assert-Equal 200 $filteredCases.Status "compensation reason filter"
    Assert-True ([int]$filteredCases.Json.data.total_items -ge 1) "compensation pagination metadata"
    $publicInternal = Invoke-Api POST "/internal/appointments/$($happy.AppointmentID)/webhook" @{ status = "SUCCESS" } $patient.Token
    Assert-True ($publicInternal.Status -in @(401, 403, 404)) "Kong blocks frontend access to internal webhook"
    $script:Results.ReadCRUD = "PASS"

    Write-Step "Writing sanitized execution summary"
    $summaryPath = Join-Path $env:TEMP "mindcare-v4-last-run-summary.json"
    $script:Results | ConvertTo-Json -Depth 20 | Set-Content -Encoding UTF8 $summaryPath
    Write-Host "`nV4 E2E PASS" -ForegroundColor Green
    foreach ($entry in $script:Results.GetEnumerator()) { Write-Host ("  {0}: {1}" -f $entry.Key, ($entry.Value | ConvertTo-Json -Compress -Depth 5)) }
    exit 0
} catch {
    Write-Host "`nV4 E2E FAIL: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

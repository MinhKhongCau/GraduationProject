package com.mindcare.auth.presentation.rest;

import com.mindcare.auth.application.usecase.InternalTokenUseCase;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

/**
 * Endpoint nội bộ — chỉ các container trong Docker network mới truy cập được.
 *
 * <p>Kong Gateway CHẶN hoàn toàn path /internal/** từ Internet.
 * (Xem block-internal-routes trong kong.yml)</p>
 *
 * <pre>
 * POST /internal/auth/token
 * Body: { "client_id": "booking-service", "client_secret": "booking_internal_secret_2024" }
 * Response: { "access_token": "eyJ...", "expires_in": 900, "token_type": "Bearer" }
 * </pre>
 */
@RestController
@RequestMapping("/internal/auth")
@RequiredArgsConstructor
@Tag(name = "Internal Auth", description = "M2M token endpoint — Docker internal network only")
public class InternalAuthController {

    private static final Logger log = LoggerFactory.getLogger(InternalAuthController.class);

    private final InternalTokenUseCase internalTokenUseCase;

    @Operation(
        summary = "Cấp Internal JWT",
        description = "Service dùng client_id + client_secret để lấy JWT nội bộ (TTL 15 phút). "
                    + "Path này bị Kong chặn từ external, chỉ dùng được trong Docker network."
    )
    @PostMapping("/token")
    public ResponseEntity<?> issueToken(@RequestBody TokenRequest request) {
        try {
            String token = internalTokenUseCase.issueInternalToken(
                    request.clientId(),
                    request.clientSecret()
            );

            return ResponseEntity.ok(Map.of(
                    "access_token", token,
                    "token_type", "Bearer",
                    "expires_in", 900  // 15 phút = 900 giây
            ));
        } catch (IllegalArgumentException e) {
            log.warn("Internal token request denied: {}", e.getMessage());
            return ResponseEntity.status(401).body(Map.of(
                    "success", false,
                    "error", "Unauthorized",
                    "message", e.getMessage()
            ));
        }
    }

    /** DTO request — dùng Java Record cho gọn */
    public record TokenRequest(
            String clientId,
            String clientSecret
    ) {}
}

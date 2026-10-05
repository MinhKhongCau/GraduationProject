package com.mindcare.auth.application.token;

import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.SignatureAlgorithm;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

import com.mindcare.auth.config.InternalClientsConfig;
import com.mindcare.auth.config.RsaKeyConfig;

import java.util.Date;
import java.util.HashMap;
import java.util.Map;

/**
 * Use case cấp phát JWT M2M (Machine-to-Machine) cho giao tiếp nội bộ giữa các service.
 *
 * <p>Flow:
 * <ol>
 *   <li>Nhận {@code client_id} + {@code client_secret} (plain text) từ service khách.</li>
 *   <li>Tra cứu bcrypt-hash của {@code client_id} trong {@link InternalClientsConfig}.</li>
 *   <li>So khớp bằng BCrypt — KHÔNG truy vấn DB.</li>
 *   <li>Nếu khớp → sinh JWT RS256 TTL 15 phút, payload: sub=client_id, role=internal.</li>
 * </ol>
 * </p>
 */
@Service
@RequiredArgsConstructor
public class InternalTokenUseCase {

    private static final Logger log = LoggerFactory.getLogger(InternalTokenUseCase.class);

    private final RsaKeyConfig rsaKeyConfig;
    private final PasswordEncoder passwordEncoder;

    /** Danh sách client hợp lệ (clientId → bcrypt hash), bind từ application.yaml */
    private final InternalClientsConfig internalClientsConfig;

    @Value("${jwt.internal-expiration-ms:900000}")
    private long internalExpirationMs;

    /**
     * Xác thực client và cấp token nội bộ.
     *
     * @param clientId     ID của service (vd: "booking-service")
     * @param clientSecret Secret plain-text do service gửi lên
     * @return JWT string đã ký RS256
     * @throws IllegalArgumentException nếu client_id không tồn tại hoặc secret sai
     */
    public String issueInternalToken(String clientId, String clientSecret) {
        String rawHashedSecret = internalClientsConfig.getHashedSecret(clientId);
        log.info("🔍 [DIAGNOSTIC] clientId: '{}', clientSecret (plain): '{}', rawHashedSecret: '{}'", clientId, clientSecret, rawHashedSecret);

        String hashedSecret = rawHashedSecret;
        if (hashedSecret != null) {
            hashedSecret = hashedSecret.replace("$$", "$");
            log.info("🔍 [DIAGNOSTIC] processed hashedSecret (escaped): '{}'", hashedSecret);
        }

        if (hashedSecret == null) {
            log.warn("⛔ Internal token request for unknown client_id: {}", clientId);
            throw new IllegalArgumentException("Unknown client_id: " + clientId);
        }

        if (!passwordEncoder.matches(clientSecret, hashedSecret)) {
            log.warn("⛔ Invalid client_secret for client_id: {}", clientId);
            throw new IllegalArgumentException("Invalid client_secret for client_id: " + clientId);
        }

        log.info("✅ Issuing internal token for service: {}", clientId);
        return buildInternalToken(clientId);
    }

    private String buildInternalToken(String clientId) {
        Map<String, Object> claims = new HashMap<>();
        claims.put("role", "internal");
        claims.put("client_id", clientId);

        long now = System.currentTimeMillis();
        return Jwts.builder()
                .setClaims(claims)
                .setSubject(clientId)
                .setIssuedAt(new Date(now))
                .setExpiration(new Date(now + internalExpirationMs))
                .signWith(rsaKeyConfig.getPrivateKey(), SignatureAlgorithm.RS256)
                .compact();
    }
}


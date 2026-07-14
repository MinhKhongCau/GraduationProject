package com.mindcare.auth.utils;

import com.mindcare.auth.domain.entity.Account;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.SignatureAlgorithm;
import io.jsonwebtoken.io.Decoders;
import io.jsonwebtoken.security.Keys;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import com.mindcare.auth.config.RsaKeyConfig;

import java.util.Date;
import java.util.HashMap;
import java.util.Map;

@Component
public class JwtUtils {

    private final RsaKeyConfig rsaKeyConfig;

    public JwtUtils(RsaKeyConfig rsaKeyConfig) {
        this.rsaKeyConfig = rsaKeyConfig;
    }

    @Value("${jwt.access-token-expiration}")
    private long jwtExpiration;

    @Value("${jwt.refresh-token-expiration}")
    private long refreshExpiration;

    // 1. Hàm in Access Token
    public String generateAccessToken(Account account) {
        Map<String, Object> extraClaims = new HashMap<>();
        extraClaims.put("role", account.getRole().name());
        extraClaims.put("accountId", account.getAccountId().toString()); // Đính kèm ID để Booking Service biết ai đang gọi
        
        return buildToken(extraClaims, account.getEmail(), jwtExpiration);
    }

    // 2. Hàm in Refresh Token (Không cần đính kèm claim phụ, chỉ cần Email)
    public String generateRefreshToken(Account account) {
        return buildToken(new HashMap<>(), account.getEmail(), refreshExpiration);
    }

    // Hàm lõi: Lắp ráp các bộ phận thành 1 chuỗi JWT hoàn chỉnh
    private String buildToken(Map<String, Object> extraClaims, String email, long expiration) {
        return Jwts.builder()
                .setClaims(extraClaims)
                .setSubject(email)
                .setIssuedAt(new Date(System.currentTimeMillis()))
                .setExpiration(new Date(System.currentTimeMillis() + expiration))
                .signWith(rsaKeyConfig.getPrivateKey(), SignatureAlgorithm.RS256)
                .compact();
    }

    public boolean validateJwtToken(String authToken) {
        try {
            Jwts.parserBuilder().setSigningKey(rsaKeyConfig.getPublicKey()).build().parseClaimsJws(authToken);
            return true;
        } catch (Exception e) {
            return false;
        }
    }

    // 4. Hàm lấy Email (Subject) từ trong Token
    public String getEmailFromToken(String token) {
        return Jwts.parserBuilder()
                .setSigningKey(rsaKeyConfig.getPublicKey())
                .build()
                .parseClaimsJws(token)
                .getBody()
                .getSubject();
    }

    // 5. Hàm băm Refresh Token bằng SHA-256 để lưu trữ gọn nhẹ và bảo mật trong DB
    public String hashToken(String token) {
        if (token == null) return null;
        try {
            java.security.MessageDigest digest = java.security.MessageDigest.getInstance("SHA-256");
            byte[] hashBytes = digest.digest(token.getBytes(java.nio.charset.StandardCharsets.UTF_8));
            StringBuilder hexString = new StringBuilder();
            for (byte b : hashBytes) {
                String hex = Integer.toHexString(0xff & b);
                if (hex.length() == 1) hexString.append('0');
                hexString.append(hex);
            }
            return hexString.toString();
        } catch (Exception e) {
            throw new RuntimeException("Error hashing token", e);
        }
    }
}
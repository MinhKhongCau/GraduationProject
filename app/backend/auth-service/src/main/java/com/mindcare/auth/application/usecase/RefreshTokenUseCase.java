package com.mindcare.auth.application.usecase;
import com.mindcare.auth.application.dto.command.TokenRefreshCommand;
import com.mindcare.auth.application.dto.response.LoginResponse;
import com.mindcare.auth.application.port.out.RefreshTokenPort;
import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.domain.entity.RefreshToken;
import com.mindcare.auth.utils.JwtUtils;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import java.time.LocalDateTime;

@Service
@RequiredArgsConstructor
public class RefreshTokenUseCase {
    private final JwtUtils jwtUtils;
    private final RefreshTokenPort refreshTokenPort;

    public LoginResponse execute(TokenRefreshCommand command) {
        String requestRefreshToken = command.getRefreshToken();
        if (!jwtUtils.validateJwtToken(requestRefreshToken)) {
            throw new RuntimeException("Invalid refresh token!");
        }
        RefreshToken refreshToken = refreshTokenPort.findByTokenHash(requestRefreshToken)
                .orElseThrow(() -> new RuntimeException("Session not found!"));
        if (refreshToken.getIsRevoked()) {
            throw new RuntimeException("Session has been revoked (Replay Attack check)!");
        }
        if (refreshToken.getExpiresAt() < System.currentTimeMillis()) {
            throw new RuntimeException("Session has expired. Please login again!");
        }
        refreshToken.setIsRevoked(true);
        refreshTokenPort.save(refreshToken);
        Account account = refreshToken.getAccount();
        String newAccessToken = jwtUtils.generateAccessToken(account);
        String newRefreshTokenString = jwtUtils.generateRefreshToken(account);
        RefreshToken newRefreshToken = RefreshToken.builder()
                .account(account)
                .tokenHash(newRefreshTokenString)
                .expiresAt(System.currentTimeMillis() + 7L * 24 * 60 * 60 * 1000L)
                .build();
        refreshTokenPort.save(newRefreshToken);
        return LoginResponse.builder()
                .accessToken(newAccessToken)
                .refreshToken(newRefreshTokenString)
                .accountId(account.getAccountId().toString())
                .fullName(account.getFullName())
                .role(account.getRole())
                .build();
    }
}

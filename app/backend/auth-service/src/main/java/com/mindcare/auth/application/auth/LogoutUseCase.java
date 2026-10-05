package com.mindcare.auth.application.auth;
import com.mindcare.auth.application.common.MessageResponse;
import com.mindcare.auth.application.port.out.RefreshTokenPort;
import com.mindcare.auth.domain.token.RefreshToken;
import com.mindcare.auth.infrastructure.security.JwtUtils;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
public class LogoutUseCase {
    private final RefreshTokenPort refreshTokenPort;
    private final JwtUtils jwtUtils;

    public MessageResponse execute(LogoutCommand command) {
        String hashedToken = jwtUtils.hashToken(command.getRefreshToken());
        RefreshToken refreshToken = refreshTokenPort.findByTokenHash(hashedToken)
                .orElseThrow(() -> new RuntimeException("Session not found!"));
        refreshToken.setIsRevoked(true);
        refreshTokenPort.save(refreshToken);
        return new MessageResponse("Logged out successfully!");
    }
}

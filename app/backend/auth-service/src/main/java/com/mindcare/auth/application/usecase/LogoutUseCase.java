package com.mindcare.auth.application.usecase;
import com.mindcare.auth.application.dto.command.LogoutCommand;
import com.mindcare.auth.application.dto.response.MessageResponse;
import com.mindcare.auth.application.port.out.RefreshTokenPort;
import com.mindcare.auth.domain.entity.RefreshToken;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
public class LogoutUseCase {
    private final RefreshTokenPort refreshTokenPort;

    public MessageResponse execute(LogoutCommand command) {
        RefreshToken refreshToken = refreshTokenPort.findByTokenHash(command.getRefreshToken())
                .orElseThrow(() -> new RuntimeException("Session not found!"));
        refreshToken.setIsRevoked(true);
        refreshTokenPort.save(refreshToken);
        return new MessageResponse("Logged out successfully!");
    }
}

package com.mindcare.auth.application.usecase;

import com.mindcare.auth.application.dto.command.LogoutCommand;
import com.mindcare.auth.application.dto.response.MessageResponse;
import com.mindcare.auth.application.port.out.RefreshTokenPort;
import com.mindcare.auth.domain.entity.RefreshToken;
import com.mindcare.auth.utils.JwtUtils;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
public class LogoutUseCaseTest {

    @Mock
    private RefreshTokenPort refreshTokenPort;

    @Mock
    private JwtUtils jwtUtils;

    @InjectMocks
    private LogoutUseCase logoutUseCase;

    @Test
    public void TC_AUTH_LGT_SVC_01() {
        // Setup Data
        LogoutCommand command = new LogoutCommand();
        command.setRefreshToken("my-active-refresh-token");

        RefreshToken refreshToken = RefreshToken.builder()
                .tokenHash("hashed-token-123")
                .isRevoked(false)
                .build();

        when(jwtUtils.hashToken("my-active-refresh-token")).thenReturn("hashed-token-123");
        when(refreshTokenPort.findByTokenHash("hashed-token-123")).thenReturn(Optional.of(refreshToken));
        when(refreshTokenPort.save(any(RefreshToken.class))).thenAnswer(invocation -> invocation.getArgument(0));

        // Execution
        MessageResponse response = logoutUseCase.execute(command);

        // Verification
        assertNotNull(response);
        assertEquals("Logged out successfully!", response.getMessage());
        assertTrue(refreshToken.getIsRevoked());

        verify(refreshTokenPort, times(1)).save(refreshToken);
    }

    @Test
    public void TC_AUTH_LGT_SVC_02() {
        // Setup Data
        LogoutCommand command = new LogoutCommand();
        command.setRefreshToken("non-existent-token");

        when(jwtUtils.hashToken("non-existent-token")).thenReturn("hashed-non-existent");
        when(refreshTokenPort.findByTokenHash("hashed-non-existent")).thenReturn(Optional.empty());

        // Execution & Verification
        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            logoutUseCase.execute(command);
        });

        assertEquals("Session not found!", exception.getMessage());
        verify(refreshTokenPort, never()).save(any(RefreshToken.class));
    }
}

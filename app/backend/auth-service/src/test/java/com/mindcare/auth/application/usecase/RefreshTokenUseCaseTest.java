package com.mindcare.auth.application.usecase;

import com.mindcare.auth.application.dto.command.TokenRefreshCommand;
import com.mindcare.auth.application.dto.response.LoginResponse;
import com.mindcare.auth.application.port.out.RefreshTokenPort;
import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.domain.entity.RefreshToken;
import com.mindcare.auth.domain.enums.Role;
import com.mindcare.auth.utils.JwtUtils;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Optional;
import java.util.UUID;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
public class RefreshTokenUseCaseTest {

    @Mock
    private JwtUtils jwtUtils;

    @Mock
    private RefreshTokenPort refreshTokenPort;

    @InjectMocks
    private RefreshTokenUseCase refreshTokenUseCase;

    @Test
    public void TC_AUTH_RFT_SVC_01_InvalidJwtToken() {
        TokenRefreshCommand command = new TokenRefreshCommand();
        command.setRefreshToken("invalid-token");

        when(jwtUtils.validateJwtToken("invalid-token")).thenReturn(false);

        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            refreshTokenUseCase.execute(command);
        });

        assertEquals("Invalid refresh token!", exception.getMessage());
        verify(refreshTokenPort, never()).findByTokenHash(anyString());
    }

    @Test
    public void TC_AUTH_RFT_SVC_02_SessionNotFound() {
        TokenRefreshCommand command = new TokenRefreshCommand();
        command.setRefreshToken("valid-format-token");

        when(jwtUtils.validateJwtToken("valid-format-token")).thenReturn(true);
        when(jwtUtils.hashToken("valid-format-token")).thenReturn("hashed-valid-format");
        when(refreshTokenPort.findByTokenHash("hashed-valid-format")).thenReturn(Optional.empty());

        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            refreshTokenUseCase.execute(command);
        });

        assertEquals("Session not found!", exception.getMessage());
    }

    @Test
    public void TC_AUTH_RFT_SVC_03_RevokedSession() {
        TokenRefreshCommand command = new TokenRefreshCommand();
        command.setRefreshToken("revoked-token");

        RefreshToken token = RefreshToken.builder()
                .tokenHash("hashed-revoked")
                .isRevoked(true)
                .build();

        when(jwtUtils.validateJwtToken("revoked-token")).thenReturn(true);
        when(jwtUtils.hashToken("revoked-token")).thenReturn("hashed-revoked");
        when(refreshTokenPort.findByTokenHash("hashed-revoked")).thenReturn(Optional.of(token));

        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            refreshTokenUseCase.execute(command);
        });

        assertEquals("Session has been revoked (Replay Attack check)!", exception.getMessage());
    }

    @Test
    public void TC_AUTH_RFT_SVC_04_ExpiredSession() {
        TokenRefreshCommand command = new TokenRefreshCommand();
        command.setRefreshToken("expired-token");

        RefreshToken token = RefreshToken.builder()
                .tokenHash("hashed-expired")
                .isRevoked(false)
                .expiresAt(System.currentTimeMillis() - 1000L) // Expired 1 sec ago
                .build();

        when(jwtUtils.validateJwtToken("expired-token")).thenReturn(true);
        when(jwtUtils.hashToken("expired-token")).thenReturn("hashed-expired");
        when(refreshTokenPort.findByTokenHash("hashed-expired")).thenReturn(Optional.of(token));

        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            refreshTokenUseCase.execute(command);
        });

        assertEquals("Session has expired. Please login again!", exception.getMessage());
    }

    @Test
    public void TC_AUTH_RFT_SVC_05_Success() {
        TokenRefreshCommand command = new TokenRefreshCommand();
        command.setRefreshToken("active-token");

        Account account = Account.builder()
                .accountId(UUID.randomUUID())
                .fullName("Jane Doe")
                .email("jane@example.com")
                .role(Role.PATIENT)
                .build();

        RefreshToken token = RefreshToken.builder()
                .tokenHash("hashed-active")
                .isRevoked(false)
                .expiresAt(System.currentTimeMillis() + 3600000L) // Valid 1h
                .account(account)
                .build();

        when(jwtUtils.validateJwtToken("active-token")).thenReturn(true);
        when(jwtUtils.hashToken("active-token")).thenReturn("hashed-active");
        when(refreshTokenPort.findByTokenHash("hashed-active")).thenReturn(Optional.of(token));
        when(jwtUtils.generateAccessToken(account)).thenReturn("new-access-token");
        when(jwtUtils.generateRefreshToken(account)).thenReturn("new-refresh-token");
        when(jwtUtils.hashToken("new-refresh-token")).thenReturn("hashed-new-refresh-token");

        LoginResponse response = refreshTokenUseCase.execute(command);

        assertNotNull(response);
        assertEquals("new-access-token", response.getAccessToken());
        assertEquals("new-refresh-token", response.getRefreshToken());
        assertEquals("Jane Doe", response.getFullName());
        assertEquals(Role.PATIENT, response.getRole());

        assertTrue(token.getIsRevoked()); // Old token revoked
        verify(refreshTokenPort, times(2)).save(any(RefreshToken.class)); // 1 for revoking old, 1 for saving new
    }
}

package com.mindcare.auth.application.usecase;

import com.mindcare.auth.application.dto.command.LoginCommand;
import com.mindcare.auth.application.dto.response.LoginResponse;
import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.application.port.out.RefreshTokenPort;
import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.domain.entity.RefreshToken;
import com.mindcare.auth.domain.enums.Role;
import com.mindcare.auth.utils.JwtUtils;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.security.crypto.password.PasswordEncoder;

import java.util.Optional;
import java.util.UUID;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
public class LoginUseCaseTest {

    @Mock
    private AccountPort accountPort;

    @Mock
    private PasswordEncoder passwordEncoder;

    @Mock
    private JwtUtils jwtUtils;

    @Mock
    private RefreshTokenPort refreshTokenPort;

    @InjectMocks
    private LoginUseCase loginUseCase;

    @Test
    public void TC_AUTH_LGN_SVC_01() {
        // Setup Data
        LoginCommand command = new LoginCommand();
        command.setEmail("user@example.com");
        command.setPassword("Secret123!");

        UUID accountId = UUID.randomUUID();
        Account account = Account.builder()
                .accountId(accountId)
                .email("user@example.com")
                .fullName("John Doe")
                .passwordHash("hashed-password-123")
                .role(Role.PATIENT)
                .build();

        when(accountPort.findByEmail("user@example.com")).thenReturn(Optional.of(account));
        when(passwordEncoder.matches("Secret123!", "hashed-password-123")).thenReturn(true);
        when(jwtUtils.generateAccessToken(account)).thenReturn("access-token-xyz");
        when(jwtUtils.generateRefreshToken(account)).thenReturn("refresh-token-xyz");
        when(jwtUtils.hashToken("refresh-token-xyz")).thenReturn("hashed-token-xyz");

        // Execution
        LoginResponse response = loginUseCase.execute(command);

        // Verification
        assertNotNull(response);
        assertEquals("access-token-xyz", response.getAccessToken());
        assertEquals("refresh-token-xyz", response.getRefreshToken());
        assertEquals(accountId.toString(), response.getAccountId());
        assertEquals("John Doe", response.getFullName());
        assertEquals(Role.PATIENT, response.getRole());

        // Verify token saved correctly
        ArgumentCaptor<RefreshToken> tokenCaptor = ArgumentCaptor.forClass(RefreshToken.class);
        verify(refreshTokenPort, times(1)).save(tokenCaptor.capture());
        RefreshToken savedToken = tokenCaptor.getValue();
        assertEquals(account, savedToken.getAccount());
        assertEquals("hashed-token-xyz", savedToken.getTokenHash());
        assertNotNull(savedToken.getExpiresAt());
        assertFalse(savedToken.getIsRevoked());
    }

    @Test
    public void TC_AUTH_LGN_SVC_02() {
        // Setup Data
        LoginCommand command = new LoginCommand();
        command.setEmail("notfound@example.com");
        command.setPassword("Secret123!");

        when(accountPort.findByEmail("notfound@example.com")).thenReturn(Optional.empty());

        // Execution & Verification
        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            loginUseCase.execute(command);
        });

        assertEquals("Account not found!", exception.getMessage());
        verify(passwordEncoder, never()).matches(anyString(), anyString());
        verify(refreshTokenPort, never()).save(any(RefreshToken.class));
    }

    @Test
    public void TC_AUTH_LGN_SVC_03() {
        // Setup Data
        LoginCommand command = new LoginCommand();
        command.setEmail("user@example.com");
        command.setPassword("WrongPassword!");

        Account account = Account.builder()
                .accountId(UUID.randomUUID())
                .email("user@example.com")
                .passwordHash("hashed-password-123")
                .build();

        when(accountPort.findByEmail("user@example.com")).thenReturn(Optional.of(account));
        when(passwordEncoder.matches("WrongPassword!", "hashed-password-123")).thenReturn(false);

        // Execution & Verification
        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            loginUseCase.execute(command);
        });

        assertEquals("Incorrect password!", exception.getMessage());
        verify(jwtUtils, never()).generateAccessToken(any(Account.class));
        verify(refreshTokenPort, never()).save(any(RefreshToken.class));
    }
}

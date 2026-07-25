package com.mindcare.auth.application.usecase;

import com.mindcare.auth.application.dto.command.ResetPasswordCommand;
import com.mindcare.auth.application.dto.response.MessageResponse;
import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.application.port.out.VerificationTokenPort;
import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.domain.entity.VerificationToken;
import com.mindcare.auth.domain.enums.TokenType;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.security.crypto.password.PasswordEncoder;

import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
public class ResetPasswordUseCaseTest {

    @Mock
    private VerificationTokenPort tokenPort;

    @Mock
    private AccountPort accountPort;

    @Mock
    private PasswordEncoder passwordEncoder;

    @InjectMocks
    private ResetPasswordUseCase resetPasswordUseCase;

    @Test
    public void TC_AUTH_PWD_RST_01_InvalidToken() {
        ResetPasswordCommand command = new ResetPasswordCommand();
        command.setToken("invalid-reset-token");
        command.setNewPassword("NewSecret123!");

        when(tokenPort.findByTokenAndTokenType("invalid-reset-token", TokenType.PASSWORD_RESET)).thenReturn(Optional.empty());

        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            resetPasswordUseCase.execute(command);
        });

        assertEquals("Invalid reset token!", exception.getMessage());
        verify(accountPort, never()).save(any());
    }

    @Test
    public void TC_AUTH_PWD_RST_02_TokenAlreadyUsed() {
        ResetPasswordCommand command = new ResetPasswordCommand();
        command.setToken("used-reset-token");
        command.setNewPassword("NewSecret123!");

        VerificationToken token = VerificationToken.builder()
                .token("used-reset-token")
                .isUsed(true)
                .build();

        when(tokenPort.findByTokenAndTokenType("used-reset-token", TokenType.PASSWORD_RESET)).thenReturn(Optional.of(token));

        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            resetPasswordUseCase.execute(command);
        });

        assertEquals("Reset token has already been used!", exception.getMessage());
    }

    @Test
    public void TC_AUTH_PWD_RST_03_TokenExpired() {
        ResetPasswordCommand command = new ResetPasswordCommand();
        command.setToken("expired-reset-token");
        command.setNewPassword("NewSecret123!");

        VerificationToken token = VerificationToken.builder()
                .token("expired-reset-token")
                .isUsed(false)
                .expiresAt(System.currentTimeMillis() - 1000L) // Expired 1 sec ago
                .build();

        when(tokenPort.findByTokenAndTokenType("expired-reset-token", TokenType.PASSWORD_RESET)).thenReturn(Optional.of(token));

        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            resetPasswordUseCase.execute(command);
        });

        assertEquals("Reset token has expired!", exception.getMessage());
    }

    @Test
    public void TC_AUTH_PWD_RST_04_Success() {
        ResetPasswordCommand command = new ResetPasswordCommand();
        command.setToken("valid-reset-token");
        command.setNewPassword("BrandNewSecret123!");

        Account account = Account.builder()
                .email("user@example.com")
                .passwordHash("hashed-old-secret")
                .build();

        VerificationToken token = VerificationToken.builder()
                .token("valid-reset-token")
                .isUsed(false)
                .expiresAt(System.currentTimeMillis() + 900000L)
                .account(account)
                .build();

        when(tokenPort.findByTokenAndTokenType("valid-reset-token", TokenType.PASSWORD_RESET)).thenReturn(Optional.of(token));
        when(passwordEncoder.encode("BrandNewSecret123!")).thenReturn("hashed-new-secret");

        MessageResponse response = resetPasswordUseCase.execute(command);

        assertNotNull(response);
        assertEquals("Password reset successfully! You can now login with your new password.", response.getMessage());
        assertEquals("hashed-new-secret", account.getPasswordHash());
        assertTrue(token.getIsUsed());

        verify(accountPort, times(1)).save(account);
        verify(tokenPort, times(1)).save(token);
    }
}

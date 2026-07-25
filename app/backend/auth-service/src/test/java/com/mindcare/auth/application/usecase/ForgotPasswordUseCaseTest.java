package com.mindcare.auth.application.usecase;

import com.mindcare.auth.application.dto.command.ForgotPasswordCommand;
import com.mindcare.auth.application.dto.response.MessageResponse;
import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.application.port.out.EmailSenderPort;
import com.mindcare.auth.application.port.out.VerificationTokenPort;
import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.domain.entity.VerificationToken;
import com.mindcare.auth.domain.enums.TokenType;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
public class ForgotPasswordUseCaseTest {

    @Mock
    private AccountPort accountPort;

    @Mock
    private VerificationTokenPort tokenPort;

    @Mock
    private EmailSenderPort emailSenderPort;

    @InjectMocks
    private ForgotPasswordUseCase forgotPasswordUseCase;

    @Test
    public void TC_AUTH_PWD_FRG_01_AccountNotFound() {
        ForgotPasswordCommand command = new ForgotPasswordCommand();
        command.setEmail("notfound@example.com");

        when(accountPort.findByEmail("notfound@example.com")).thenReturn(Optional.empty());

        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            forgotPasswordUseCase.execute(command);
        });

        assertEquals("Account not found!", exception.getMessage());
        verify(tokenPort, never()).save(any());
        verify(emailSenderPort, never()).sendEmail(anyString(), anyString(), anyString());
    }

    @Test
    public void TC_AUTH_PWD_FRG_02_Success() {
        ForgotPasswordCommand command = new ForgotPasswordCommand();
        command.setEmail("user@example.com");

        Account account = Account.builder()
                .email("user@example.com")
                .fullName("John Doe")
                .build();

        when(accountPort.findByEmail("user@example.com")).thenReturn(Optional.of(account));

        MessageResponse response = forgotPasswordUseCase.execute(command);

        assertNotNull(response);
        assertEquals("Password reset request sent successfully to your email!", response.getMessage());

        ArgumentCaptor<VerificationToken> tokenCaptor = ArgumentCaptor.forClass(VerificationToken.class);
        verify(tokenPort, times(1)).save(tokenCaptor.capture());
        VerificationToken savedToken = tokenCaptor.getValue();

        assertEquals(account, savedToken.getAccount());
        assertEquals(TokenType.PASSWORD_RESET, savedToken.getTokenType());
        assertNotNull(savedToken.getToken());
        assertTrue(savedToken.getExpiresAt() > System.currentTimeMillis());

        verify(emailSenderPort, times(1)).sendEmail(eq("user@example.com"), eq("Password Reset Request"), contains(savedToken.getToken()));
    }
}

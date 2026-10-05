package com.mindcare.auth.application.password;

import com.mindcare.auth.application.common.MessageResponse;
import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.domain.account.Account;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
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
public class ChangePasswordUseCaseTest {

    @Mock
    private AccountPort accountPort;

    @Mock
    private PasswordEncoder passwordEncoder;

    @InjectMocks
    private ChangePasswordUseCase changePasswordUseCase;

    @Test
    public void TC_AUTH_PWD_CHG_01_AccountNotFound() {
        ChangePasswordCommand command = new ChangePasswordCommand();
        command.setOldPassword("OldSecret123!");
        command.setNewPassword("NewSecret123!");
        command.setConfirmNewPassword("NewSecret123!");

        when(accountPort.findByEmail("unknown@example.com")).thenReturn(Optional.empty());

        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            changePasswordUseCase.execute("unknown@example.com", command);
        });

        assertEquals("Account not found!", exception.getMessage());
        verify(accountPort, never()).save(any());
    }

    @Test
    public void TC_AUTH_PWD_CHG_02_IncorrectOldPassword() {
        ChangePasswordCommand command = new ChangePasswordCommand();
        command.setOldPassword("WrongOldSecret!");
        command.setNewPassword("NewSecret123!");
        command.setConfirmNewPassword("NewSecret123!");

        Account account = Account.builder()
                .email("user@example.com")
                .passwordHash("hashed-old-secret")
                .build();

        when(accountPort.findByEmail("user@example.com")).thenReturn(Optional.of(account));
        when(passwordEncoder.matches("WrongOldSecret!", "hashed-old-secret")).thenReturn(false);

        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            changePasswordUseCase.execute("user@example.com", command);
        });

        assertEquals("Incorrect old password!", exception.getMessage());
        verify(accountPort, never()).save(any());
    }

    @Test
    public void TC_AUTH_PWD_CHG_03_ConfirmPasswordMismatch() {
        ChangePasswordCommand command = new ChangePasswordCommand();
        command.setOldPassword("OldSecret123!");
        command.setNewPassword("NewSecret123!");
        command.setConfirmNewPassword("MismatchNewSecret!");

        Account account = Account.builder()
                .email("user@example.com")
                .passwordHash("hashed-old-secret")
                .build();

        when(accountPort.findByEmail("user@example.com")).thenReturn(Optional.of(account));
        when(passwordEncoder.matches("OldSecret123!", "hashed-old-secret")).thenReturn(true);

        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            changePasswordUseCase.execute("user@example.com", command);
        });

        assertEquals("Confirmation password does not match!", exception.getMessage());
        verify(accountPort, never()).save(any());
    }

    @Test
    public void TC_AUTH_PWD_CHG_04_Success() {
        ChangePasswordCommand command = new ChangePasswordCommand();
        command.setOldPassword("OldSecret123!");
        command.setNewPassword("NewSecret123!");
        command.setConfirmNewPassword("NewSecret123!");

        Account account = Account.builder()
                .accountId(UUID.randomUUID())
                .email("user@example.com")
                .passwordHash("hashed-old-secret")
                .build();

        when(accountPort.findByEmail("user@example.com")).thenReturn(Optional.of(account));
        when(passwordEncoder.matches("OldSecret123!", "hashed-old-secret")).thenReturn(true);
        when(passwordEncoder.encode("NewSecret123!")).thenReturn("hashed-new-secret");

        MessageResponse response = changePasswordUseCase.execute("user@example.com", command);

        assertNotNull(response);
        assertEquals("Password changed successfully! Please use your new password for the next login.", response.getMessage());
        assertEquals("hashed-new-secret", account.getPasswordHash());
        verify(accountPort, times(1)).save(account);
    }
}

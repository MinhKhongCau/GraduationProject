package com.mindcare.auth.application.usecase;

import com.mindcare.auth.application.dto.command.RegisterCommand;
import com.mindcare.auth.application.dto.response.MessageResponse;
import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.domain.enums.Role;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.security.crypto.password.PasswordEncoder;

import java.time.LocalDate;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
public class RegisterUseCaseTest {

    @Mock
    private AccountPort accountPort;

    @Mock
    private PasswordEncoder passwordEncoder;

    @InjectMocks
    private RegisterUseCase registerUseCase;

    @Test
    public void TC_AUTH_REG_SVC_01() {
        // Setup Data
        RegisterCommand command = new RegisterCommand();
        command.setFullName("John Doe");
        command.setEmail("john@example.com");
        command.setPassword("Secret123!");
        command.setConfirmPassword("Secret123!");
        command.setRole(Role.EXPERT);
        command.setDateOfBirth(LocalDate.of(1990, 1, 1));

        when(accountPort.existsByEmail("john@example.com")).thenReturn(false);
        when(passwordEncoder.encode("Secret123!")).thenReturn("encrypted-secret-hash");
        when(accountPort.save(any(Account.class))).thenAnswer(invocation -> invocation.getArgument(0));

        // Execution
        MessageResponse response = registerUseCase.execute(command);

        // Verification
        assertNotNull(response);
        assertEquals("User registered successfully!", response.getMessage());

        ArgumentCaptor<Account> accountCaptor = ArgumentCaptor.forClass(Account.class);
        verify(accountPort, times(1)).save(accountCaptor.capture());
        Account savedAccount = accountCaptor.getValue();

        assertEquals("John Doe", savedAccount.getFullName());
        assertEquals("john@example.com", savedAccount.getEmail());
        assertEquals("encrypted-secret-hash", savedAccount.getPasswordHash());
        assertEquals(Role.EXPERT, savedAccount.getRole());
        assertTrue(savedAccount.getIsEmailVerified());
        assertEquals(LocalDate.of(1990, 1, 1), savedAccount.getDateOfBirth());
    }

    @Test
    public void TC_AUTH_REG_SVC_02() {
        // Setup Data
        RegisterCommand command = new RegisterCommand();
        command.setFullName("Patient One");
        command.setEmail("patient@example.com");
        command.setPassword("Secret123!");
        command.setConfirmPassword("Secret123!");
        command.setRole(null); // Null role -> should default to PATIENT

        when(accountPort.existsByEmail("patient@example.com")).thenReturn(false);
        when(passwordEncoder.encode("Secret123!")).thenReturn("encrypted-patient-hash");
        when(accountPort.save(any(Account.class))).thenAnswer(invocation -> invocation.getArgument(0));

        // Execution
        MessageResponse response = registerUseCase.execute(command);

        // Verification
        assertNotNull(response);
        assertEquals("User registered successfully!", response.getMessage());

        ArgumentCaptor<Account> accountCaptor = ArgumentCaptor.forClass(Account.class);
        verify(accountPort, times(1)).save(accountCaptor.capture());
        Account savedAccount = accountCaptor.getValue();

        assertEquals("Patient One", savedAccount.getFullName());
        assertEquals("patient@example.com", savedAccount.getEmail());
        assertEquals("encrypted-patient-hash", savedAccount.getPasswordHash());
        assertEquals(Role.PATIENT, savedAccount.getRole()); // Defaulted
    }

    @Test
    public void TC_AUTH_REG_SVC_03() {
        // Setup Data
        RegisterCommand command = new RegisterCommand();
        command.setFullName("John Doe");
        command.setEmail("john@example.com");
        command.setPassword("Secret123!");
        command.setConfirmPassword("DifferentPassword123!");

        // Execution & Verification
        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            registerUseCase.execute(command);
        });

        assertEquals("Confirmation password does not match!", exception.getMessage());
        verify(accountPort, never()).existsByEmail(anyString());
        verify(accountPort, never()).save(any(Account.class));
    }

    @Test
    public void TC_AUTH_REG_SVC_04() {
        // Setup Data
        RegisterCommand command = new RegisterCommand();
        command.setFullName("Existing User");
        command.setEmail("existing@example.com");
        command.setPassword("Secret123!");
        command.setConfirmPassword("Secret123!");

        when(accountPort.existsByEmail("existing@example.com")).thenReturn(true);

        // Execution & Verification
        RuntimeException exception = assertThrows(RuntimeException.class, () -> {
            registerUseCase.execute(command);
        });

        assertEquals("Email is already registered!", exception.getMessage());
        verify(passwordEncoder, never()).encode(anyString());
        verify(accountPort, never()).save(any(Account.class));
    }
}

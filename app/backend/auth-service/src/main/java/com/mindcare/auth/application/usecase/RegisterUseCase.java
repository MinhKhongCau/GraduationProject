package com.mindcare.auth.application.usecase;
import com.mindcare.auth.application.dto.command.RegisterCommand;
import com.mindcare.auth.application.dto.response.MessageResponse;
import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.application.service.UserEventPublisher;
import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.domain.enums.Role;
import lombok.RequiredArgsConstructor;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
public class RegisterUseCase {
    private final AccountPort accountPort;
    private final PasswordEncoder passwordEncoder;
    private final UserEventPublisher userEventPublisher;

    public MessageResponse execute(RegisterCommand command) {
        if (!command.getPassword().equals(command.getConfirmPassword())) {
            throw new RuntimeException("Confirmation password does not match!");
        }
        if (accountPort.existsByEmail(command.getEmail())) {
            throw new RuntimeException("Email is already registered!");
        }
        Account newAccount = Account.builder()
                .fullName(command.getFullName())
                .dateOfBirth(command.getDateOfBirth())
                .email(command.getEmail())
                .passwordHash(passwordEncoder.encode(command.getPassword()))
                .role(command.getRole() != null ? command.getRole() : Role.PATIENT)
                .isEmailVerified(true)
                .build();
        Account savedAccount = accountPort.save(newAccount);

        // Publish only after the account has been successfully persisted,
        // per RABBITMQ_CONVENTION.md's "publish after a successful DB transaction" rule.
        userEventPublisher.publishUserCreated(savedAccount);

        return new MessageResponse("User registered successfully!");
    }
}

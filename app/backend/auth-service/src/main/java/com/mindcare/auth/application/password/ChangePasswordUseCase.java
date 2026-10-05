package com.mindcare.auth.application.password;
import com.mindcare.auth.application.common.MessageResponse;
import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.domain.account.Account;
import lombok.RequiredArgsConstructor;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
public class ChangePasswordUseCase {
    private final AccountPort accountPort;
    private final PasswordEncoder passwordEncoder;

    public MessageResponse execute(String email, ChangePasswordCommand command) {
        Account account = accountPort.findByEmail(email)
                .orElseThrow(() -> new RuntimeException("Account not found!"));
        if (!passwordEncoder.matches(command.getOldPassword(), account.getPasswordHash())) {
            throw new RuntimeException("Incorrect old password!");
        }
        if (!command.getNewPassword().equals(command.getConfirmNewPassword())) {
            throw new RuntimeException("Confirmation password does not match!");
        }
        account.setPasswordHash(passwordEncoder.encode(command.getNewPassword()));
        accountPort.save(account);
        return new MessageResponse("Password changed successfully! Please use your new password for the next login.");
    }
}

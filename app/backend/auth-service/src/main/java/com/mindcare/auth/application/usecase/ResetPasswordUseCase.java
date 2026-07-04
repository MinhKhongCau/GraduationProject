package com.mindcare.auth.application.usecase;
import com.mindcare.auth.application.dto.command.ResetPasswordCommand;
import com.mindcare.auth.application.dto.response.MessageResponse;
import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.application.port.out.VerificationTokenPort;
import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.domain.entity.VerificationToken;
import com.mindcare.auth.domain.enums.TokenType;
import lombok.RequiredArgsConstructor;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import java.time.LocalDateTime;

@Service
@RequiredArgsConstructor
public class ResetPasswordUseCase {
    private final VerificationTokenPort tokenPort;
    private final AccountPort accountPort;
    private final PasswordEncoder passwordEncoder;

    public MessageResponse execute(ResetPasswordCommand command) {
        VerificationToken token = tokenPort.findByTokenAndTokenType(command.getToken(), TokenType.PASSWORD_RESET)
                .orElseThrow(() -> new RuntimeException("Invalid reset token!"));
        if (token.getIsUsed()) {
            throw new RuntimeException("Reset token has already been used!");
        }
        if (token.getExpiresAt() < System.currentTimeMillis()) {
            throw new RuntimeException("Reset token has expired!");
        }
        Account account = token.getAccount();
        account.setPasswordHash(passwordEncoder.encode(command.getNewPassword()));
        accountPort.save(account);
        token.setIsUsed(true);
        tokenPort.save(token);
        return new MessageResponse("Password reset successfully! You can now login with your new password.");
    }
}

package com.mindcare.auth.application.password;
import com.mindcare.auth.application.common.MessageResponse;
import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.application.port.out.EmailSenderPort;
import com.mindcare.auth.application.port.out.VerificationTokenPort;
import com.mindcare.auth.domain.account.Account;
import com.mindcare.auth.domain.token.VerificationToken;
import com.mindcare.auth.domain.token.TokenType;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import java.time.LocalDateTime;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class ForgotPasswordUseCase {
    private final AccountPort accountPort;
    private final VerificationTokenPort tokenPort;
    private final EmailSenderPort emailSenderPort;

    public MessageResponse execute(ForgotPasswordCommand command) {
        Account account = accountPort.findByEmail(command.getEmail())
                .orElseThrow(() -> new RuntimeException("Account not found!"));
        String resetToken = UUID.randomUUID().toString();
        VerificationToken token = VerificationToken.builder()
                .token(resetToken)
                .account(account)
                .tokenType(TokenType.PASSWORD_RESET)
                .expiresAt(System.currentTimeMillis() + 15 * 60 * 1000L)
                .build();
        tokenPort.save(token);
        String emailBody = "Your password reset token is: " + resetToken + ". It will expire in 15 minutes.";
        emailSenderPort.sendEmail(account.getEmail(), "Password Reset Request", emailBody);
        return new MessageResponse("Password reset request sent successfully to your email!");
    }
}

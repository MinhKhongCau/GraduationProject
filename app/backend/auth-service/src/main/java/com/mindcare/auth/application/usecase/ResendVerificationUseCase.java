package com.mindcare.auth.application.usecase;
import com.mindcare.auth.application.dto.command.ResendVerificationCommand;
import com.mindcare.auth.application.dto.response.MessageResponse;
import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.application.port.out.EmailSenderPort;
import com.mindcare.auth.application.port.out.VerificationTokenPort;
import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.domain.entity.VerificationToken;
import com.mindcare.auth.domain.enums.TokenType;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import java.time.LocalDateTime;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class ResendVerificationUseCase {
    private final AccountPort accountPort;
    private final VerificationTokenPort tokenPort;
    private final EmailSenderPort emailSenderPort;

    public MessageResponse execute(ResendVerificationCommand command) {
        Account account = accountPort.findByEmail(command.getEmail())
                .orElseThrow(() -> new RuntimeException("Account not found!"));
        if (account.getIsEmailVerified()) {
            throw new RuntimeException("Email is already verified!");
        }
        String code = UUID.randomUUID().toString().substring(0, 6).toUpperCase();
        VerificationToken token = VerificationToken.builder()
                .token(code)
                .account(account)
                .tokenType(TokenType.EMAIL_VERIFICATION)
                .expiresAt(System.currentTimeMillis() + 15 * 60 * 1000L)
                .build();
        tokenPort.save(token);
        String emailBody = "Your email verification code is: " + code + ". It will expire in 15 minutes.";
        emailSenderPort.sendEmail(account.getEmail(), "Verify Your Email", emailBody);
        return new MessageResponse("Verification email resent successfully!");
    }
}

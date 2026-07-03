package com.mindcare.auth.application.usecase;
import com.mindcare.auth.application.dto.command.VerifyEmailCommand;
import com.mindcare.auth.application.dto.response.MessageResponse;
import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.application.port.out.VerificationTokenPort;
import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.domain.entity.VerificationToken;
import com.mindcare.auth.domain.enums.TokenType;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import java.time.LocalDateTime;

@Service
@RequiredArgsConstructor
public class VerifyEmailUseCase {
    private final VerificationTokenPort tokenPort;
    private final AccountPort accountPort;

    public MessageResponse execute(VerifyEmailCommand command) {
        VerificationToken token = tokenPort.findByTokenAndTokenType(command.getCode(), TokenType.EMAIL_VERIFICATION)
                .orElseThrow(() -> new RuntimeException("Invalid verification code!"));
        if (token.getIsUsed()) {
            throw new RuntimeException("Verification code has already been used!");
        }
        if (token.getExpiresAt() < System.currentTimeMillis()) {
            throw new RuntimeException("Verification code has expired!");
        }
        Account account = token.getAccount();
        if (!account.getEmail().equals(command.getUserId()) && !account.getAccountId().toString().equals(command.getUserId())) {
             throw new RuntimeException("Verification code does not match the provided user!");
        }
        account.setIsEmailVerified(true);
        accountPort.save(account);
        token.setIsUsed(true);
        tokenPort.save(token);
        return new MessageResponse("Email verified successfully!");
    }
}

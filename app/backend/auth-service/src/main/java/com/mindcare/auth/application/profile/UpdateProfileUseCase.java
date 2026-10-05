package com.mindcare.auth.application.profile;
import com.mindcare.auth.application.common.MessageResponse;
import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.domain.account.Account;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
public class UpdateProfileUseCase {
    private final AccountPort accountPort;

    public MessageResponse execute(String email, ProfileUpdateCommand command) {
        Account account = accountPort.findByEmail(email)
                .orElseThrow(() -> new RuntimeException("Account not found!"));
        if (command.getFullName() != null && !command.getFullName().trim().isEmpty()) {
            account.setFullName(command.getFullName());
        }
        if (command.getDateOfBirth() != null) {
            account.setDateOfBirth(command.getDateOfBirth());
        }
        accountPort.save(account);
        return new MessageResponse("Profile updated successfully!");
    }
}

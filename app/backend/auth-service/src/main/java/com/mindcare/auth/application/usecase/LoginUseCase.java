package com.mindcare.auth.application.usecase;
import com.mindcare.auth.application.dto.command.LoginCommand;
import com.mindcare.auth.application.dto.response.LoginResponse;
import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.application.port.out.RefreshTokenPort;
import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.domain.entity.RefreshToken;
import com.mindcare.auth.utils.JwtUtils;
import lombok.RequiredArgsConstructor;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import java.time.LocalDateTime;

@Service
@RequiredArgsConstructor
public class LoginUseCase {
    private final AccountPort accountPort;
    private final PasswordEncoder passwordEncoder;
    private final JwtUtils jwtUtils;
    private final RefreshTokenPort refreshTokenPort;

    public LoginResponse execute(LoginCommand command) {
        Account account = accountPort.findByEmail(command.getEmail())
                .orElseThrow(() -> new RuntimeException("Account not found!"));
        if (!passwordEncoder.matches(command.getPassword(), account.getPasswordHash())) {
            throw new RuntimeException("Incorrect password!");
        }
        String accessToken = jwtUtils.generateAccessToken(account);
        String refreshTokenString = jwtUtils.generateRefreshToken(account);
        RefreshToken refreshToken = RefreshToken.builder()
                .account(account)
                .tokenHash(jwtUtils.hashToken(refreshTokenString)) 
                .expiresAt(System.currentTimeMillis() + 7L * 24 * 60 * 60 * 1000L)
                .build();
        refreshTokenPort.save(refreshToken);
        return LoginResponse.builder()
                .accessToken(accessToken)
                .refreshToken(refreshTokenString)
                .accountId(account.getAccountId().toString())
                .fullName(account.getFullName())
                .role(account.getRole())
                .build();
    }
}

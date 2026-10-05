package com.mindcare.auth.application.port.out;

import com.mindcare.auth.domain.token.VerificationToken;
import com.mindcare.auth.domain.token.TokenType;
import java.util.Optional;

public interface VerificationTokenPort {
    Optional<VerificationToken> findByTokenAndTokenType(String token, TokenType tokenType);
    VerificationToken save(VerificationToken token);
}

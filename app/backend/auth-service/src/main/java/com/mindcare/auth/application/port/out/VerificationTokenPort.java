package com.mindcare.auth.application.port.out;

import com.mindcare.auth.domain.entity.VerificationToken;
import com.mindcare.auth.domain.enums.TokenType;
import java.util.Optional;

public interface VerificationTokenPort {
    Optional<VerificationToken> findByTokenAndTokenType(String token, TokenType tokenType);
    VerificationToken save(VerificationToken token);
}

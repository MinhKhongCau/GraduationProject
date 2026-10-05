package com.mindcare.auth.application.port.out;

import com.mindcare.auth.domain.token.RefreshToken;
import java.util.Optional;

public interface RefreshTokenPort {
    Optional<RefreshToken> findByTokenHash(String tokenHash);
    RefreshToken save(RefreshToken refreshToken);
}

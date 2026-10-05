package com.mindcare.auth.infrastructure.persistence.adapter;
import com.mindcare.auth.application.port.out.VerificationTokenPort;
import com.mindcare.auth.domain.token.VerificationToken;
import com.mindcare.auth.domain.token.TokenType;
import com.mindcare.auth.infrastructure.persistence.repository.VerificationTokenJpaRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;
import java.util.Optional;
@Component
@RequiredArgsConstructor
public class VerificationTokenPersistenceAdapter implements VerificationTokenPort {
    private final VerificationTokenJpaRepository repository;
    @Override
    public Optional<VerificationToken> findByTokenAndTokenType(String token, TokenType tokenType) {
        return repository.findByTokenAndTokenType(token, tokenType);
    }
    @Override
    public VerificationToken save(VerificationToken token) {
        return repository.save(token);
    }
}

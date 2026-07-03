package com.mindcare.auth.infrastructure.persistence;
import com.mindcare.auth.domain.entity.VerificationToken;
import com.mindcare.auth.domain.enums.TokenType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import java.util.Optional;
import java.util.UUID;
@Repository
public interface VerificationTokenJpaRepository extends JpaRepository<VerificationToken, UUID> {
    Optional<VerificationToken> findByTokenAndTokenType(String token, TokenType tokenType);
}

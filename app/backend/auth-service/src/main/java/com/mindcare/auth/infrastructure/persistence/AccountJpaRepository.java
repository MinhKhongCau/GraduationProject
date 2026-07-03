package com.mindcare.auth.infrastructure.persistence;

import com.mindcare.auth.domain.entity.Account;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;
import java.util.UUID;

@Repository
public interface AccountJpaRepository extends JpaRepository<Account, UUID> {
    // TÃ¬m tÃ i khoáº£n báº±ng email (DÃ¹ng khi ÄÄƒng nháº­p)
    Optional<Account> findByEmail(String email);

    // Kiá»ƒm tra xem email Ä‘Ã£ cÃ³ ai Ä‘Äƒng kÃ½ chÆ°a (DÃ¹ng khi ÄÄƒng kÃ½)
    boolean existsByEmail(String email);
}

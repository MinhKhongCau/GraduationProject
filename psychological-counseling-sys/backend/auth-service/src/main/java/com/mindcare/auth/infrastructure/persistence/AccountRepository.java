package com.mindcare.auth.infrastructure.persistence;

import com.mindcare.auth.domain.entity.Account;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;
import java.util.UUID;

@Repository
public interface AccountRepository extends JpaRepository<Account, UUID> {
    // Tìm tài khoản bằng email (Dùng khi Đăng nhập)
    Optional<Account> findByEmail(String email);

    // Kiểm tra xem email đã có ai đăng ký chưa (Dùng khi Đăng ký)
    boolean existsByEmail(String email);
}
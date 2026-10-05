package com.mindcare.auth.application.port.out;

import com.mindcare.auth.domain.account.Account;
import java.util.Optional;

public interface AccountPort {
    boolean existsByEmail(String email);
    Optional<Account> findByEmail(String email);
    Account save(Account account);
}

package com.mindcare.auth.infrastructure.persistence.adapter;

import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.domain.account.Account;
import com.mindcare.auth.infrastructure.persistence.repository.AccountJpaRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.Optional;

@Component
@RequiredArgsConstructor
public class AccountPersistenceAdapter implements AccountPort {

    private final AccountJpaRepository accountJpaRepository;

    @Override
    public boolean existsByEmail(String email) {
        return accountJpaRepository.existsByEmail(email);
    }

    @Override
    public Optional<Account> findByEmail(String email) {
        return accountJpaRepository.findByEmail(email);
    }

    @Override
    public Account save(Account account) {
        return accountJpaRepository.save(account);
    }
}

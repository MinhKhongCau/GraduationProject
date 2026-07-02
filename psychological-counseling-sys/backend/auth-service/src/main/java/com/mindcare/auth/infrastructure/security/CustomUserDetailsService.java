package com.mindcare.auth.infrastructure.security;

import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.infrastructure.persistence.AccountRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.userdetails.User;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.core.userdetails.UserDetailsService;
import org.springframework.security.core.userdetails.UsernameNotFoundException;
import org.springframework.stereotype.Service;

import java.util.Collections;

@Service
@RequiredArgsConstructor
public class CustomUserDetailsService implements UserDetailsService {

    private final AccountRepository accountRepository;

    @Override
    public UserDetails loadUserByUsername(String email) throws UsernameNotFoundException {
        // Lấy thông tin user từ DB của chúng ta
        Account account = accountRepository.findByEmail(email)
                .orElseThrow(() -> new UsernameNotFoundException("Không tìm thấy email: " + email));

        // Chuyển đổi sang định dạng mà Spring Security hiểu được
        return new User(
                account.getEmail(),
                account.getPasswordHash(),
                Collections.singletonList(new SimpleGrantedAuthority("ROLE_" + account.getRole()))
        );
    }
}
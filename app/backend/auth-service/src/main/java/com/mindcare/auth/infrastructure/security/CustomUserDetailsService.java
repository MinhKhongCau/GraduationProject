package com.mindcare.auth.infrastructure.security;

import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.application.port.out.AccountPort;
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

    private final AccountPort accountPort;

    @Override
    public UserDetails loadUserByUsername(String email) throws UsernameNotFoundException {
        // Láº¥y thÃ´ng tin user tá»« DB cá»§a chÃºng ta
        Account account = accountPort.findByEmail(email)
                .orElseThrow(() -> new UsernameNotFoundException("KhÃ´ng tÃ¬m tháº¥y email: " + email));

        // Chuyá»ƒn Ä‘á»•i sang Ä‘á»‹nh dáº¡ng mÃ  Spring Security hiá»ƒu Ä‘Æ°á»£c
        return new User(
                account.getEmail(),
                account.getPasswordHash(),
                Collections.singletonList(new SimpleGrantedAuthority("ROLE_" + account.getRole().name()))
        );
    }
}

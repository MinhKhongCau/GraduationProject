package com.mindcare.auth.infrastructure.persistence;

import com.mindcare.auth.application.port.out.AccountPort;
import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.domain.enums.Role;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.CommandLineRunner;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Component;

import java.time.LocalDate;

@Component
@RequiredArgsConstructor
@Slf4j
public class DataSeeder implements CommandLineRunner {

    private final AccountPort accountPort;
    private final PasswordEncoder passwordEncoder;

    @org.springframework.beans.factory.annotation.Value("${server.port:8080}")
    private String serverPort;

    @Override
    public void run(String... args) throws Exception {
        log.info("\n========================================================================\n" +
                "🚀 MINDCARE SYSTEM - PORT ROUTING INFORMATION\n" +
                "========================================================================\n" +
                "[1] AUTH SERVICE (SPRING BOOT BACKEND):\n" +
                "    - Local URL:   http://localhost:{}\n" +
                "    - Swagger UI:  http://localhost:{}/swagger-ui/index.html\n" +
                "    - Test Route:  http://localhost:{}/api/v1/auth/test-gateway\n\n" +
                "[2] API GATEWAY (KONG GATEWAY):\n" +
                "    - Proxy URL:   http://localhost:8000\n" +
                "    - Admin URL:   http://localhost:8001\n" +
                "    - Route Auth:  http://localhost:8000/api/v1/auth/test-gateway\n\n" +
                "[3] DATABASE CONFIG (POSTGRESQL):\n" +
                "    - Local DB Port: 5434  (Host Port for DBeaver)\n" +
                "    - JDBC URL:      jdbc:postgresql://localhost:5434/auth_db\n" +
                "    - Container:     5432  (Internal Docker Port)\n" +
                "========================================================================", 
                serverPort, serverPort, serverPort);

        // Seed 3 tài khoản mặc định nếu DB chưa có tài khoản nào
        if (!accountPort.existsByEmail("admin@mindcare.com")) {
            log.info("========== SEEDING TEST ACCOUNTS ==========");
            
            // 1. ADMIN
            Account admin = Account.builder()
                    .fullName("System Administrator")
                    .email("admin@mindcare.com")
                    .passwordHash(passwordEncoder.encode("admin@mindcare.com"))
                    .role(Role.ADMIN)
                    .isEmailVerified(true)
                    .isActive(true)
                    .dateOfBirth(LocalDate.of(1990, 1, 1))
                    .build();
            accountPort.save(admin);
            log.info("Created Admin Account: admin@mindcare.com / admin@mindcare.com");

            // 2. EXPERT
            Account expert = Account.builder()
                    .fullName("Dr. MindCare Expert")
                    .email("expert@mindcare.com")
                    .passwordHash(passwordEncoder.encode("expert@mindcare.com"))
                    .role(Role.EXPERT)
                    .isEmailVerified(true)
                    .isActive(true)
                    .dateOfBirth(LocalDate.of(1985, 5, 15))
                    .build();
            accountPort.save(expert);
            log.info("Created Expert Account: expert@mindcare.com / expert@mindcare.com");

            // 3. PATIENT
            Account patient = Account.builder()
                    .fullName("John Patient")
                    .email("patient@mindcare.com")
                    .passwordHash(passwordEncoder.encode("patient@mindcare.com"))
                    .role(Role.PATIENT)
                    .isEmailVerified(true)
                    .isActive(true)
                    .dateOfBirth(LocalDate.of(2000, 10, 10))
                    .build();
            accountPort.save(patient);
            log.info("Created Patient Account: patient@mindcare.com / patient@mindcare.com");
            
            log.info("===========================================");
        }
    }
}

package com.mindcare.auth.config;

import com.mindcare.auth.infrastructure.security.JwtAuthFilter; // Import máy quét thẻ
import lombok.RequiredArgsConstructor;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.method.configuration.EnableMethodSecurity;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

@Configuration
@EnableWebSecurity
@EnableMethodSecurity
@RequiredArgsConstructor 
public class SecurityConfig {

    private final JwtAuthFilter jwtAuthFilter; // Gọi máy quét thẻ vào

    @Bean
    public PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder();
    }

    @Bean
    public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
        http
            .csrf(AbstractHttpConfigurer::disable)
            // Chuẩn Microservices: Không lưu trạng thái Session (Stateless)
            .sessionManagement(session -> session.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
            .authorizeHttpRequests(auth -> auth
                // Health check cho Kong Gateway, không cần token
                .requestMatchers("/health").permitAll()
                // Swagger UI / OpenAPI docs, không cần token
                .requestMatchers("/swagger-ui/**", "/swagger-ui.html", "/v3/api-docs/**").permitAll()
                .requestMatchers(
                    "/api/v1/auth/register",
                    "/api/v1/auth/login",
                    "/api/v1/auth/refresh",
                    "/api/v1/auth/logout",
                    "/api/v1/auth/verify-email",
                    "/api/v1/auth/resend-verification",
                    "/api/v1/auth/forgot-password",
                    "/api/v1/auth/reset-password",
                    "/api/v1/auth/test-gateway",
                    "/api/v1/auth/public-key",
                    "/swagger-ui/**",
                    "/v3/api-docs/**",
                    "/swagger-ui.html"
                ).permitAll()
                // ─── Internal M2M endpoint ────────────────────────────────────────────
                // Spring bỏ qua auth filter → chỉ Kong mới chặn từ external.
                // Đây là "double lock": Kong block ngoài, Spring chỉ check client_secret bên trong.
                .requestMatchers("/internal/auth/token").permitAll()
                // Tất cả các cửa khác phải quét thẻ
                .anyRequest().authenticated()
            );

        // Lắp máy quét thẻ vào ngay trước cửa kiểm tra mặc định của Spring
        http.addFilterBefore(jwtAuthFilter, UsernamePasswordAuthenticationFilter.class);

        return http.build();
    }
}
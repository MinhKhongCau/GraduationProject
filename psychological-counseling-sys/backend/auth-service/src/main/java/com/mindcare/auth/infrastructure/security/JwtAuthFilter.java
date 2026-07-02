package com.mindcare.auth.infrastructure.security;

import com.mindcare.auth.utils.JwtUtils;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.web.authentication.WebAuthenticationDetailsSource;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;

@Component
@RequiredArgsConstructor
public class JwtAuthFilter extends OncePerRequestFilter {

    private final JwtUtils jwtUtils;
    private final CustomUserDetailsService userDetailsService;

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain filterChain)
            throws ServletException, IOException {
        try {
            // 1. Tách lấy Token từ Header gửi lên
            String jwt = parseJwt(request);

            // 2. Nếu có Token và Token hợp lệ (chưa hết hạn, đúng chữ ký)
            if (jwt != null && jwtUtils.validateJwtToken(jwt)) {
                // 3. Lấy email từ token và kéo thông tin từ DB lên
                String email = jwtUtils.getEmailFromToken(jwt);
                UserDetails userDetails = userDetailsService.loadUserByUsername(email);

                // 4. Đóng mộc "Hợp lệ", cho phép đi vào hệ thống
                UsernamePasswordAuthenticationToken authentication = new UsernamePasswordAuthenticationToken(
                        userDetails, null, userDetails.getAuthorities());
                authentication.setDetails(new WebAuthenticationDetailsSource().buildDetails(request));

                SecurityContextHolder.getContext().setAuthentication(authentication);
            }
        } catch (Exception e) {
            System.out.println("Không thể xác thực người dùng: " + e.getMessage());
        }

        // Bỏ qua và cho request đi tiếp đến Controller
        filterChain.doFilter(request, response);
    }

    // Hàm phụ: Cắt bỏ chữ "Bearer " ở đầu để lấy đúng mã thẻ Token
    private String parseJwt(HttpServletRequest request) {
        String headerAuth = request.getHeader("Authorization");
        if (StringUtils.hasText(headerAuth) && headerAuth.startsWith("Bearer ")) {
            return headerAuth.substring(7);
        }
        return null;
    }
}
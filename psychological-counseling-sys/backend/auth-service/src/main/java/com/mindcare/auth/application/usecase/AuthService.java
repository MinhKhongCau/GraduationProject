package com.mindcare.auth.application.usecase;

import com.mindcare.auth.domain.entity.Account;
import com.mindcare.auth.domain.entity.RefreshToken;
import com.mindcare.auth.application.dto.request.ChangePasswordRequest;
import com.mindcare.auth.application.dto.request.LoginRequest;
import com.mindcare.auth.application.dto.request.LogoutRequest;
import com.mindcare.auth.application.dto.request.ProfileUpdateRequest;
import com.mindcare.auth.application.dto.request.RegisterRequest;
import com.mindcare.auth.application.dto.request.TokenRefreshRequest;
import com.mindcare.auth.application.dto.response.LoginResponse;
import com.mindcare.auth.application.dto.response.MessageResponse;
import com.mindcare.auth.infrastructure.persistence.AccountRepository;
import com.mindcare.auth.infrastructure.persistence.RefreshTokenRepository;
import com.mindcare.auth.utils.JwtUtils;

import lombok.RequiredArgsConstructor;

import java.time.LocalDateTime;

import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor // Lombok tự động tạo Constructor để nhúng (inject) Repository và PasswordEncoder
public class AuthService {

    private final AccountRepository accountRepository;
    private final PasswordEncoder passwordEncoder;

    private final JwtUtils jwtUtils;
    private final RefreshTokenRepository refreshTokenRepository;

    public MessageResponse register(RegisterRequest request) {
        // 1. Kiểm tra 2 mật khẩu có khớp nhau không
        if (!request.getPassword().equals(request.getConfirmPassword())) {
            throw new RuntimeException("Lỗi: Mật khẩu xác nhận không khớp!");
        }

        // 2. Kiểm tra xem Email đã tồn tại chưa
        if (accountRepository.existsByEmail(request.getEmail())) {
            throw new RuntimeException("Lỗi: Email này đã được đăng ký!");
        }

        // 3. Tạo đối tượng Account mới với đầy đủ thông tin
        Account newAccount = Account.builder()
                .fullName(request.getFullName())       // THÊM MỚI
                .dateOfBirth(request.getDateOfBirth()) // THÊM MỚI
                .email(request.getEmail())
                .passwordHash(passwordEncoder.encode(request.getPassword()))
                .role(request.getRole() != null ? request.getRole().toUpperCase() : "PATIENT")
                .build();

        // 4. Lưu xuống Database
        accountRepository.save(newAccount);

        return new MessageResponse("Đăng ký tài khoản thành công!");
    }

    public LoginResponse login(LoginRequest request) {
        // 1. Tìm tài khoản trong DB theo Email
        Account account = accountRepository.findByEmail(request.getEmail())
                .orElseThrow(() -> new RuntimeException("Lỗi: Tài khoản không tồn tại!"));

        // 2. Đối chiếu mật khẩu (Bản rõ người dùng nhập vs Bản Hash trong DB)
        if (!passwordEncoder.matches(request.getPassword(), account.getPasswordHash())) {
            throw new RuntimeException("Lỗi: Mật khẩu không chính xác!");
        }

        // 3. In Thẻ (Tokens)
        String accessToken = jwtUtils.generateAccessToken(account);
        String refreshTokenString = jwtUtils.generateRefreshToken(account);

        // 4. Lưu Refresh Token vào Database để quản lý phiên
        RefreshToken refreshToken = RefreshToken.builder()
                .account(account)
                .tokenHash(refreshTokenString) 
                .expiresAt(LocalDateTime.now().plusDays(7)) // Trùng với hạn 7 ngày cấu hình trong yml
                .build();
        refreshTokenRepository.save(refreshToken);

        // 5. Đóng gói gửi về cho Frontend
        return LoginResponse.builder()
                .accessToken(accessToken)
                .refreshToken(refreshTokenString)
                .accountId(account.getAccountId().toString())
                .fullName(account.getFullName())
                .role(account.getRole())
                .build();
    }

    public LoginResponse refreshToken(TokenRefreshRequest request) {
        String requestRefreshToken = request.getRefreshToken();

        // 1. Kiểm tra tính hợp lệ của Token về mặt kỹ thuật
        if (!jwtUtils.validateJwtToken(requestRefreshToken)) {
            throw new RuntimeException("Lỗi: Refresh Token không hợp lệ!");
        }

        // 2. Tìm token trong Database xem có tồn tại và chưa bị thu hồi không
        RefreshToken refreshToken = refreshTokenRepository.findByTokenHash(requestRefreshToken)
                .orElseThrow(() -> new RuntimeException("Lỗi: Phiên đăng nhập không tồn tại!"));

        if (refreshToken.getIsRevoked()) {
            throw new RuntimeException("Lỗi: Phiên đăng nhập đã bị hủy!");
        }

        if (refreshToken.getExpiresAt().isBefore(LocalDateTime.now())) {
            throw new RuntimeException("Lỗi: Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại!");
        }

        // 3. Nếu mọi thứ OK, lấy thông tin Account và in Access Token mới
        Account account = refreshToken.getAccount();
        String newAccessToken = jwtUtils.generateAccessToken(account);

        // Trả về Access Token mới (giữ nguyên Refresh Token cũ hoặc in mới tùy bạn, ở đây ta giữ nguyên)
        return LoginResponse.builder()
                .accessToken(newAccessToken)
                .refreshToken(requestRefreshToken)
                .accountId(account.getAccountId().toString())
                .fullName(account.getFullName())
                .role(account.getRole())
                .build();
    }

    public MessageResponse logout(LogoutRequest request) {
        // 1. Tìm Refresh Token trong Database
        RefreshToken refreshToken = refreshTokenRepository.findByTokenHash(request.getRefreshToken())
                .orElseThrow(() -> new RuntimeException("Lỗi: Không tìm thấy phiên đăng nhập!"));

        // 2. Cập nhật trạng thái thành "Đã thu hồi"
        refreshToken.setIsRevoked(true);

        // 3. Lưu lại vào Database
        refreshTokenRepository.save(refreshToken);

        return new MessageResponse("Đăng xuất thành công!");
    }

    public MessageResponse updateProfile(String email, ProfileUpdateRequest request) {
        // 1. Tìm người dùng trong Database dựa trên Email lấy từ Token
        Account account = accountRepository.findByEmail(email)
                .orElseThrow(() -> new RuntimeException("Lỗi: Không tìm thấy tài khoản!"));

        // 2. Cập nhật thông tin mới (nếu có gửi lên)
        if (request.getFullName() != null && !request.getFullName().trim().isEmpty()) {
            account.setFullName(request.getFullName());
        }
        
        if (request.getDateOfBirth() != null) {
            account.setDateOfBirth(request.getDateOfBirth());
        }

        // 3. Lưu đè xuống Database
        accountRepository.save(account);

        return new MessageResponse("Cập nhật thông tin cá nhân thành công!");
    }

    public MessageResponse changePassword(String email, ChangePasswordRequest request) {
        // 1. Tìm tài khoản trong DB
        Account account = accountRepository.findByEmail(email)
                .orElseThrow(() -> new RuntimeException("Lỗi: Không tìm thấy tài khoản!"));

        // 2. Kiểm tra mật khẩu cũ có đúng không
        if (!passwordEncoder.matches(request.getOldPassword(), account.getPasswordHash())) {
            throw new RuntimeException("Lỗi: Mật khẩu cũ không chính xác!");
        }

        // 3. Kiểm tra mật khẩu mới và xác nhận mật khẩu mới có khớp nhau không
        if (!request.getNewPassword().equals(request.getConfirmNewPassword())) {
            throw new RuntimeException("Lỗi: Mật khẩu xác nhận không khớp!");
        }

        // 4. Băm mật khẩu mới và lưu xuống DB
        account.setPasswordHash(passwordEncoder.encode(request.getNewPassword()));
        accountRepository.save(account);

        return new MessageResponse("Thay đổi mật khẩu thành công! Vui lòng dùng mật khẩu mới cho lần đăng nhập sau.");
    }
}
package com.mindcare.auth.presentation.rest;

import com.mindcare.auth.application.dto.request.ChangePasswordRequest;
import com.mindcare.auth.application.dto.request.LoginRequest;
import com.mindcare.auth.application.dto.request.LogoutRequest;
import com.mindcare.auth.application.dto.request.ProfileUpdateRequest;
import com.mindcare.auth.application.dto.request.RegisterRequest;
import com.mindcare.auth.application.dto.request.TokenRefreshRequest;
import com.mindcare.auth.application.dto.response.ApiResponse;
import com.mindcare.auth.application.dto.response.MessageResponse;
import com.mindcare.auth.application.usecase.AuthService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;
import java.security.Principal;

@RestController
@RequestMapping("/api/v1/auth") // Prefix chuẩn cho API Auth
@RequiredArgsConstructor
public class AuthController {

    private final AuthService authService;

    @PostMapping("/register")
    public ResponseEntity<ApiResponse<Object>> registerUser(@RequestBody RegisterRequest request) {
        try {
            Object data = authService.register(request);
            return ResponseEntity.ok(ApiResponse.success("Đăng ký tài khoản thành công!", data));
        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error("Đăng ký thất bại", e.getMessage()));
        }
    }

    @PostMapping("/login")
    public ResponseEntity<ApiResponse<Object>> loginUser(@RequestBody LoginRequest request) {
        try {
            Object data = authService.login(request);
            return ResponseEntity.ok(ApiResponse.success("Đăng nhập thành công!", data));
        } catch (RuntimeException e) {
            return ResponseEntity.status(401).body(ApiResponse.error("Đăng nhập thất bại", e.getMessage()));
        }
    }

    @PostMapping("/refresh")
    public ResponseEntity<ApiResponse<Object>> refreshToken(@RequestBody TokenRefreshRequest request) {
        try {
            Object data = authService.refreshToken(request);
            return ResponseEntity.ok(ApiResponse.success("Làm mới token thành công", data));
        } catch (RuntimeException e) {
            return ResponseEntity.status(403).body(ApiResponse.error("Làm mới token thất bại", e.getMessage()));
        }
    }

    @PostMapping("/logout")
    public ResponseEntity<ApiResponse<Object>> logoutUser(@RequestBody LogoutRequest request) {
        try {
            Object data = authService.logout(request);
            return ResponseEntity.ok(ApiResponse.success("Đăng xuất thành công", data));
        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error("Đăng xuất thất bại", e.getMessage()));
        }
    }

    @GetMapping("/me")
    public ResponseEntity<ApiResponse<String>> getMyProfile(Principal principal) {
        String email = principal.getName(); 
        return ResponseEntity.ok(ApiResponse.success("Thẻ Access Token hợp lệ!", "Xin chào user: " + email));
    }

    @PutMapping("/profile")
    public ResponseEntity<ApiResponse<Object>> updateMyProfile(
            @RequestBody ProfileUpdateRequest request, 
            Principal principal) {
        try {
            String email = principal.getName(); 
            Object data = authService.updateProfile(email, request);
            return ResponseEntity.ok(ApiResponse.success("Cập nhật hồ sơ thành công!", data));
        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error("Cập nhật hồ sơ thất bại", e.getMessage()));
        }
    }

    @PreAuthorize("hasRole('EXPERT')") 
    @GetMapping("/expert-only")
    public ResponseEntity<ApiResponse<String>> getExpertDashboard() {
        return ResponseEntity.ok(ApiResponse.success("Đã vào được khu vực giới hạn", "Chào mừng Bác sĩ! Bạn đã vào được khu vực VIP."));
    }

    @PutMapping("/change-password")
    public ResponseEntity<ApiResponse<Object>> changePassword(
            @RequestBody ChangePasswordRequest request,
            Principal principal) {
        try {
            String email = principal.getName();
            Object data = authService.changePassword(email, request);
            return ResponseEntity.ok(ApiResponse.success("Đổi mật khẩu thành công!", data));
        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error("Đổi mật khẩu thất bại", e.getMessage()));
        }
    }
}
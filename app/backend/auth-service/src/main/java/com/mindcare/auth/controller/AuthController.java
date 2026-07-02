package com.mindcare.auth.controller;

import com.mindcare.auth.dto.request.ChangePasswordRequest;
import com.mindcare.auth.dto.request.LoginRequest;
import com.mindcare.auth.dto.request.LogoutRequest;
import com.mindcare.auth.dto.request.ProfileUpdateRequest;
import com.mindcare.auth.dto.request.RegisterRequest;
import com.mindcare.auth.dto.request.TokenRefreshRequest;
import com.mindcare.auth.dto.response.MessageResponse;
import com.mindcare.auth.service.AuthService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;
import java.security.Principal;
import org.springframework.security.access.prepost.PreAuthorize;

@RestController
@RequestMapping("/api/v1/auth") // Prefix chuẩn cho API Auth
@RequiredArgsConstructor
public class AuthController {

    private final AuthService authService;

    @PostMapping("/register")
    public ResponseEntity<?> registerUser(@RequestBody RegisterRequest request) {
        try {
            // Gọi hàm xử lý từ AuthService
            return ResponseEntity.ok(authService.register(request));
        } catch (RuntimeException e) {
            // Nếu có lỗi (ví dụ trùng email), trả về mã 400 Bad Request
            return ResponseEntity.badRequest().body(e.getMessage());
        }
    }

    @PostMapping("/login")
    public ResponseEntity<?> loginUser(@RequestBody LoginRequest request) {
        try {
            return ResponseEntity.ok(authService.login(request));
        } catch (RuntimeException e) {
            // Lỗi sai mật khẩu/tài khoản trả về mã 401 Unauthorized
            return ResponseEntity.status(401).body(new MessageResponse(e.getMessage()));
        }
    }

    @PostMapping("/refresh")
    public ResponseEntity<?> refreshToken(@RequestBody TokenRefreshRequest request) {
        try {
            return ResponseEntity.ok(authService.refreshToken(request));
        } catch (RuntimeException e) {
            return ResponseEntity.status(403).body(new MessageResponse(e.getMessage()));
        }
    }

    @PostMapping("/logout")
    public ResponseEntity<?> logoutUser(@RequestBody LogoutRequest request) {
        try {
            return ResponseEntity.ok(authService.logout(request));
        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(new MessageResponse(e.getMessage()));
        }
    }

    @GetMapping("/me")
    public ResponseEntity<?> getMyProfile(Principal principal) {
        // Tham số 'principal' được Spring tự động gán vào nhờ cái JwtAuthFilter lúc nãy
        String email = principal.getName(); 
        return ResponseEntity.ok(new MessageResponse("Thẻ Access Token hợp lệ! Xin chào user: " + email));
    }

    @PutMapping("/profile")
    public ResponseEntity<?> updateMyProfile(
            @RequestBody ProfileUpdateRequest request, 
            Principal principal) {
        try {
            // Lấy email từ thẻ Token đã quét thành công
            String email = principal.getName(); 
            
            // Đẩy xuống Service xử lý
            return ResponseEntity.ok(authService.updateProfile(email, request));
        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(new MessageResponse(e.getMessage()));
        }
    }

    @PreAuthorize("hasRole('EXPERT')") 
    @GetMapping("/expert-only")
    public ResponseEntity<?> getExpertDashboard() {
        return ResponseEntity.ok(new MessageResponse("Chào mừng Bác sĩ! Bạn đã vào được khu vực VIP."));
    }

    @PutMapping("/change-password")
    public ResponseEntity<?> changePassword(
            @RequestBody ChangePasswordRequest request,
            Principal principal) {
        try {
            // Lấy email từ token hợp lệ
            String email = principal.getName();
            return ResponseEntity.ok(authService.changePassword(email, request));
        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(new MessageResponse(e.getMessage()));
        }
    }
}
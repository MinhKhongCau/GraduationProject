package com.mindcare.auth.presentation.rest;
import com.mindcare.auth.application.dto.command.ChangePasswordCommand;
import com.mindcare.auth.application.dto.command.LoginCommand;
import com.mindcare.auth.application.dto.command.LogoutCommand;
import com.mindcare.auth.application.dto.command.ProfileUpdateCommand;
import com.mindcare.auth.application.dto.command.RegisterCommand;
import com.mindcare.auth.application.dto.command.TokenRefreshCommand;
import com.mindcare.auth.application.dto.command.VerifyEmailCommand;
import com.mindcare.auth.application.dto.command.ResendVerificationCommand;
import com.mindcare.auth.application.dto.command.ForgotPasswordCommand;
import com.mindcare.auth.application.dto.command.ResetPasswordCommand;
import com.mindcare.auth.shared.dto.ApiResponse;
import com.mindcare.auth.application.usecase.*;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;
import java.security.Principal;

@RestController
@RequestMapping("/api/v1/auth")
@RequiredArgsConstructor
@Tag(name = "Authentication", description = "Endpoints for user authentication, registration, password reset and email verification.")
public class AuthController {
    private final RegisterUseCase registerUseCase;
    private final LoginUseCase loginUseCase;
    private final RefreshTokenUseCase refreshTokenUseCase;
    private final LogoutUseCase logoutUseCase;
    private final UpdateProfileUseCase updateProfileUseCase;
    private final ChangePasswordUseCase changePasswordUseCase;
    private final VerifyEmailUseCase verifyEmailUseCase;
    private final ResendVerificationUseCase resendVerificationUseCase;
    private final ForgotPasswordUseCase forgotPasswordUseCase;
    private final ResetPasswordUseCase resetPasswordUseCase;

    @Operation(summary = "Health Check", description = "Checks if the authentication service is running.")
    @GetMapping("/test-gateway")
    public ResponseEntity<String> testGateway() {
        return ResponseEntity.ok("NestJS (Java) is Alive!");
    }

    @Operation(summary = "Register User", description = "Registers a new user in the system.")
    @PostMapping("/register")
    public ResponseEntity<ApiResponse<Object>> registerUser(@Valid @RequestBody RegisterCommand command) {
        Object data = registerUseCase.execute(command);
        return ResponseEntity.ok(ApiResponse.success("User registered successfully!", data));
    }

    @Operation(summary = "Login User", description = "Authenticates a user and returns access/refresh tokens.")
    @PostMapping("/login")
    public ResponseEntity<ApiResponse<Object>> loginUser(@Valid @RequestBody LoginCommand command) {
        Object data = loginUseCase.execute(command);
        return ResponseEntity.ok(ApiResponse.success("Login successful!", data));
    }

    @Operation(summary = "Refresh Token", description = "Exchanges a valid refresh token for a new set of tokens.")
    @PostMapping("/refresh")
    public ResponseEntity<ApiResponse<Object>> refreshToken(@Valid @RequestBody TokenRefreshCommand command) {
        Object data = refreshTokenUseCase.execute(command);
        return ResponseEntity.ok(ApiResponse.success("Token refreshed successfully!", data));
    }

    @Operation(summary = "Logout User", description = "Revokes the provided refresh token.")
    @PostMapping("/logout")
    public ResponseEntity<ApiResponse<Object>> logoutUser(@Valid @RequestBody LogoutCommand command) {
        Object data = logoutUseCase.execute(command);
        return ResponseEntity.ok(ApiResponse.success("Logout successful!", data));
    }

    @Operation(summary = "Get Profile", description = "Returns basic information of the currently authenticated user.")
    @GetMapping("/me")
    public ResponseEntity<ApiResponse<String>> getMyProfile(Principal principal) {
        String email = principal.getName(); 
        return ResponseEntity.ok(ApiResponse.success("Valid Access Token!", "Hello user: " + email));
    }

    @Operation(summary = "Update Profile", description = "Updates the authenticated user's profile information.")
    @PutMapping("/profile")
    public ResponseEntity<ApiResponse<Object>> updateMyProfile(
            @Valid @RequestBody ProfileUpdateCommand command, 
            Principal principal) {
        String email = principal.getName(); 
        Object data = updateProfileUseCase.execute(email, command);
        return ResponseEntity.ok(ApiResponse.success("Profile updated successfully!", data));
    }

    @Operation(summary = "Expert Area", description = "Restricted endpoint for EXPERT role only.")
    @PreAuthorize("hasRole('EXPERT')") 
    @GetMapping("/expert-only")
    public ResponseEntity<ApiResponse<String>> getExpertDashboard() {
        return ResponseEntity.ok(ApiResponse.success("Entered restricted area", "Welcome Doctor! You are in the VIP zone."));
    }

    @Operation(summary = "Change Password", description = "Allows authenticated users to change their password.")
    @PostMapping("/change-password")
    public ResponseEntity<ApiResponse<Object>> changePassword(
            @Valid @RequestBody ChangePasswordCommand command,
            Principal principal) {
        String email = principal.getName();
        Object data = changePasswordUseCase.execute(email, command);
        return ResponseEntity.ok(ApiResponse.success("Password changed successfully!", data));
    }

    @Operation(summary = "Verify Email", description = "Verifies user email using the verification code.")
    @PostMapping("/verify-email")
    public ResponseEntity<ApiResponse<Object>> verifyEmail(@Valid @RequestBody VerifyEmailCommand command) {
        Object data = verifyEmailUseCase.execute(command);
        return ResponseEntity.ok(ApiResponse.success("Email verified successfully!", data));
    }

    @Operation(summary = "Resend Verification Email", description = "Resends a new email verification code to the user's email.")
    @PostMapping("/resend-verification")
    public ResponseEntity<ApiResponse<Object>> resendVerification(@Valid @RequestBody ResendVerificationCommand command) {
        Object data = resendVerificationUseCase.execute(command);
        return ResponseEntity.ok(ApiResponse.success("Verification email resent successfully!", data));
    }

    @Operation(summary = "Forgot Password", description = "Generates a password reset token and sends it via email.")
    @PostMapping("/forgot-password")
    public ResponseEntity<ApiResponse<Object>> forgotPassword(@Valid @RequestBody ForgotPasswordCommand command) {
        Object data = forgotPasswordUseCase.execute(command);
        return ResponseEntity.ok(ApiResponse.success("Password reset request sent successfully!", data));
    }

    @Operation(summary = "Reset Password", description = "Resets user password using the token sent to their email.")
    @PostMapping("/reset-password")
    public ResponseEntity<ApiResponse<Object>> resetPassword(@Valid @RequestBody ResetPasswordCommand command) {
        Object data = resetPasswordUseCase.execute(command);
        return ResponseEntity.ok(ApiResponse.success("Password reset successfully!", data));
    }
}

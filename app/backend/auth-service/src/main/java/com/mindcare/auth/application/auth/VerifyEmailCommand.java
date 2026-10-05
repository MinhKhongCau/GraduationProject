package com.mindcare.auth.application.auth;
import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;
@Data
public class VerifyEmailCommand {
    @Schema(description = "User ID or Email", example = "user@example.com")
    @NotBlank(message = "User ID/Email is required")
    private String userId;
    @Schema(description = "Verification code", example = "123456")
    @NotBlank(message = "Verification code is required")
    private String code;
}

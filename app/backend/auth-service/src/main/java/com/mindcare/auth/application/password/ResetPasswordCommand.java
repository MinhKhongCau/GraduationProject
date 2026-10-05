package com.mindcare.auth.application.password;
import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;
@Data
public class ResetPasswordCommand {
    @Schema(description = "Reset token sent via email", example = "xyz123")
    @NotBlank(message = "Token is required")
    private String token;
    @Schema(description = "New password", example = "NewPass123!")
    @NotBlank(message = "New password is required")
    private String newPassword;
}

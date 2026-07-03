package com.mindcare.auth.application.dto.command;
import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;
@Data
public class ForgotPasswordCommand {
    @Schema(description = "Email of the user requesting password reset", example = "user@example.com")
    @NotBlank(message = "Email is required")
    @Email(message = "Invalid email format")
    private String email;
}

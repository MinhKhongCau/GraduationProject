package com.mindcare.auth.application.auth;
import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;
@Data
public class LoginCommand {
    @Schema(description = "Email of the user", example = "user@example.com")
    @NotBlank(message = "Email is required")
    @Email(message = "Invalid email format")
    private String email;
    @Schema(description = "Password of the user", example = "Secret123!")
    @NotBlank(message = "Password is required")
    private String password;
}

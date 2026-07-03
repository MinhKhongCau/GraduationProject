package com.mindcare.auth.application.dto.command;
import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import com.mindcare.auth.domain.enums.Role;
import lombok.Data;
import java.time.LocalDate;
@Data
public class RegisterCommand {
    @Schema(description = "Full name of the user", example = "John Doe")
    @NotBlank(message = "Full name is required")
    private String fullName;
    @Schema(description = "Email of the user", example = "user@example.com")
    @NotBlank(message = "Email is required")
    @Email(message = "Invalid email format")
    private String email;
    @Schema(description = "Password of the user", example = "Secret123!")
    @NotBlank(message = "Password is required")
    @Size(min = 6, message = "Password must be at least 6 characters")
    private String password;
    @Schema(description = "Confirmation of password", example = "Secret123!")
    @NotBlank(message = "Confirmation password is required")
    private String confirmPassword;
    @Schema(description = "Date of birth", example = "1990-01-01")
    private LocalDate dateOfBirth;
    @Schema(description = "User role", example = "PATIENT")
    private Role role; 
}

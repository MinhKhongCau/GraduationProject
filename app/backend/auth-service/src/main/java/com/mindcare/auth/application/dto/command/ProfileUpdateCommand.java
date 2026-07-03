package com.mindcare.auth.application.dto.command;
import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;
import java.time.LocalDate;
@Data
public class ProfileUpdateCommand {
    @Schema(description = "Full name of the user", example = "John Doe")
    @NotBlank(message = "Full name is required")
    private String fullName;
    @Schema(description = "Date of birth", example = "1990-01-01")
    private LocalDate dateOfBirth;
}

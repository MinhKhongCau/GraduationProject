package com.mindcare.auth.application.dto.command;
import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;
@Data
public class LogoutCommand {
    @Schema(description = "The refresh token to be revoked", example = "eyJhbGciOiJIUzI1NiIsIn...")
    @NotBlank(message = "Refresh token is required")
    private String refreshToken;
}

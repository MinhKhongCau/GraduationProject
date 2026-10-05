package com.mindcare.auth.application.token;
import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;
@Data
public class TokenRefreshCommand {
    @Schema(description = "The old refresh token", example = "eyJhbGciOiJIUzI1NiIsIn...")
    @NotBlank(message = "Refresh token is required")
    private String refreshToken;
}

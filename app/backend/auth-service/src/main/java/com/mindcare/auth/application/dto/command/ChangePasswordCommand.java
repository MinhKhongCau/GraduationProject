package com.mindcare.auth.application.dto.command;
import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;
@Data
public class ChangePasswordCommand {
    @Schema(description = "Old password of the user", example = "OldPass123")
    @NotBlank(message = "Old password is required")
    private String oldPassword;
    @Schema(description = "New password of the user", example = "NewPass123!")
    @NotBlank(message = "New password is required")
    private String newPassword;
    @Schema(description = "Confirmation of the new password", example = "NewPass123!")
    @NotBlank(message = "Confirmation password is required")
    private String confirmNewPassword;
}

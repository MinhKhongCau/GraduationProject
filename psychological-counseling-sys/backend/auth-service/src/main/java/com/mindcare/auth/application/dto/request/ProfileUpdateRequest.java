package com.mindcare.auth.application.dto.request;

import lombok.Data;
import java.time.LocalDate;

@Data
public class ProfileUpdateRequest {
    private String fullName;
    private LocalDate dateOfBirth; // Định dạng yyyy-MM-dd
}
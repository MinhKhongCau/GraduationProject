package com.mindcare.auth.application.dto.request;

import lombok.Data;
import java.time.LocalDate;

@Data
public class RegisterRequest {
    private String fullName;
    private String email;
    private String password;
    private String confirmPassword;
    private LocalDate dateOfBirth; // Định dạng chuẩn: yyyy-MM-dd
    private String role; 
}
package com.mindcare.auth.infrastructure.http.request;

import lombok.Data;

@Data
public class LoginRequest {
    private String email;
    private String password;
}
package com.mindcare.auth.application.dto.request;

import lombok.Data;

@Data
public class TokenRefreshRequest {
    private String refreshToken;
}
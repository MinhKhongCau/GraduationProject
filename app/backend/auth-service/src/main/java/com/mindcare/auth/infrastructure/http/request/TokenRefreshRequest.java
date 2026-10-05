package com.mindcare.auth.infrastructure.http.request;

import lombok.Data;

@Data
public class TokenRefreshRequest {
    private String refreshToken;
}
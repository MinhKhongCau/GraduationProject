package com.mindcare.auth.application.auth;

import com.mindcare.auth.domain.account.Role;
import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class LoginResponse {
    private String accessToken;
    private String refreshToken;
    private String accountId;
    private String fullName;
    private Role role;
}
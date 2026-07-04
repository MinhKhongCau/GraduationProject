package com.mindcare.auth.application.dto.response;

import com.mindcare.auth.domain.enums.Role;
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
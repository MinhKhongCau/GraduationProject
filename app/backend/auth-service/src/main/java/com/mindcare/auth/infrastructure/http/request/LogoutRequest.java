package com.mindcare.auth.infrastructure.http.request;

import lombok.Data;

@Data
public class LogoutRequest {
    private String refreshToken; // Chỉ cần gửi thẻ dự phòng lên để Server tiêu diệt
}
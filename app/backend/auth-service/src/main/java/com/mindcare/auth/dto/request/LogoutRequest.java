package com.mindcare.auth.dto.request;

import lombok.Data;

@Data
public class LogoutRequest {
    private String refreshToken; // Chỉ cần gửi thẻ dự phòng lên để Server tiêu diệt
}
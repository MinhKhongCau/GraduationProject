package com.mindcare.auth.infrastructure.http.controller;

import com.mindcare.auth.config.RsaKeyConfig;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/api/v1/auth")
@Tag(name = "Authentication", description = "API cho đăng nhập, đăng ký và xác thực")
public class PublicKeyController {

    private final RsaKeyConfig rsaKeyConfig;

    public PublicKeyController(RsaKeyConfig rsaKeyConfig) {
        this.rsaKeyConfig = rsaKeyConfig;
    }

    @Operation(summary = "Lấy Public Key", description = "Trả về Public Key (Base64) dùng để xác thực JWT (RS256)")
    @GetMapping("/public-key")
    public ResponseEntity<Map<String, String>> getPublicKey() {
        Map<String, String> response = new HashMap<>();
        response.put("publicKey", rsaKeyConfig.getPublicKeyBase64());
        return ResponseEntity.ok(response);
    }
}

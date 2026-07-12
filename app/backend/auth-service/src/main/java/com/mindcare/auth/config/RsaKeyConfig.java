package com.mindcare.auth.config;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Configuration;

import jakarta.annotation.PostConstruct;
import java.security.KeyFactory;
import java.security.KeyPair;
import java.security.KeyPairGenerator;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.security.spec.PKCS8EncodedKeySpec;
import java.security.spec.X509EncodedKeySpec;
import java.util.Base64;

@Configuration
public class RsaKeyConfig {

    private static final Logger log = LoggerFactory.getLogger(RsaKeyConfig.class);

    @Value("${jwt.private-key:}")
    private String privateKeyString;

    @Value("${jwt.public-key:}")
    private String publicKeyString;

    private PrivateKey privateKey;
    private PublicKey publicKey;
    private String publicKeyBase64;

    @PostConstruct
    public void init() throws Exception {
        if (privateKeyString != null && !privateKeyString.trim().isEmpty() 
            && publicKeyString != null && !publicKeyString.trim().isEmpty()) {
            
            log.info("✅ Loading RSA keys from environment configuration...");
            KeyFactory keyFactory = KeyFactory.getInstance("RSA");
            
            // Xóa các khoảng trắng, dấu \n nếu có (tránh lỗi parse Base64 do copy/paste)
            String cleanPriv = privateKeyString.replaceAll("\\s+", "");
            byte[] privateKeyBytes = Base64.getDecoder().decode(cleanPriv);
            PKCS8EncodedKeySpec privateKeySpec = new PKCS8EncodedKeySpec(privateKeyBytes);
            this.privateKey = keyFactory.generatePrivate(privateKeySpec);
            
            String cleanPub = publicKeyString.replaceAll("\\s+", "");
            byte[] publicKeyBytes = Base64.getDecoder().decode(cleanPub);
            X509EncodedKeySpec publicKeySpec = new X509EncodedKeySpec(publicKeyBytes);
            this.publicKey = keyFactory.generatePublic(publicKeySpec);
            
            this.publicKeyBase64 = cleanPub;
        } else {
            log.warn("⚠️ WARNING: JWT keys are not configured in environment! Generating random RSA keys on RAM. Tokens will invalidate on restart.");
            KeyPairGenerator keyPairGenerator = KeyPairGenerator.getInstance("RSA");
            keyPairGenerator.initialize(2048);
            KeyPair keyPair = keyPairGenerator.generateKeyPair();
            this.privateKey = keyPair.getPrivate();
            this.publicKey = keyPair.getPublic();
            this.publicKeyBase64 = Base64.getEncoder().encodeToString(this.publicKey.getEncoded());
        }
    }

    public PrivateKey getPrivateKey() {
        return privateKey;
    }

    public PublicKey getPublicKey() {
        return publicKey;
    }

    public String getPublicKeyBase64() {
        return publicKeyBase64;
    }
}

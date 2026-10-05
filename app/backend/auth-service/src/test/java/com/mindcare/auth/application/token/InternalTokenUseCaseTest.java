package com.mindcare.auth.application.token;

import com.mindcare.auth.config.InternalClientsConfig;
import com.mindcare.auth.config.RsaKeyConfig;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.test.util.ReflectionTestUtils;

import java.security.KeyPair;
import java.security.KeyPairGenerator;
import java.security.interfaces.RSAPrivateKey;
import java.security.interfaces.RSAPublicKey;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
public class InternalTokenUseCaseTest {

    @Mock
    private RsaKeyConfig rsaKeyConfig;

    @Mock
    private PasswordEncoder passwordEncoder;

    @Mock
    private InternalClientsConfig internalClientsConfig;

    @InjectMocks
    private InternalTokenUseCase internalTokenUseCase;

    private RSAPrivateKey privateKey;

    @BeforeEach
    public void setup() throws Exception {
        KeyPairGenerator keyPairGenerator = KeyPairGenerator.getInstance("RSA");
        keyPairGenerator.initialize(2048);
        KeyPair keyPair = keyPairGenerator.generateKeyPair();
        privateKey = (RSAPrivateKey) keyPair.getPrivate();

        ReflectionTestUtils.setField(internalTokenUseCase, "internalExpirationMs", 900000L);
    }

    @Test
    public void TC_AUTH_INT_01_UnknownClientId() {
        when(internalClientsConfig.getHashedSecret("unknown-service")).thenReturn(null);

        IllegalArgumentException exception = assertThrows(IllegalArgumentException.class, () -> {
            internalTokenUseCase.issueInternalToken("unknown-service", "secret");
        });

        assertEquals("Unknown client_id: unknown-service", exception.getMessage());
    }

    @Test
    public void TC_AUTH_INT_02_InvalidClientSecret() {
        when(internalClientsConfig.getHashedSecret("booking-service")).thenReturn("hashed-booking-secret");
        when(passwordEncoder.matches("wrong-secret", "hashed-booking-secret")).thenReturn(false);

        IllegalArgumentException exception = assertThrows(IllegalArgumentException.class, () -> {
            internalTokenUseCase.issueInternalToken("booking-service", "wrong-secret");
        });

        assertEquals("Invalid client_secret for client_id: booking-service", exception.getMessage());
    }

    @Test
    public void TC_AUTH_INT_03_Success() {
        when(internalClientsConfig.getHashedSecret("booking-service")).thenReturn("hashed-booking-secret");
        when(passwordEncoder.matches("valid-secret", "hashed-booking-secret")).thenReturn(true);
        when(rsaKeyConfig.getPrivateKey()).thenReturn(privateKey);

        String token = internalTokenUseCase.issueInternalToken("booking-service", "valid-secret");

        assertNotNull(token);
        assertFalse(token.isEmpty());
        assertTrue(token.split("\\.").length == 3); // Valid JWT structure header.payload.signature
    }
}

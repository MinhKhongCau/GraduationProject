package com.mindcare.auth.infrastructure.http.controller;

import com.mindcare.auth.config.RsaKeyConfig;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;

import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@ExtendWith(MockitoExtension.class)
public class PublicKeyControllerTest {

    private MockMvc mockMvc;

    @Mock
    private RsaKeyConfig rsaKeyConfig;

    @InjectMocks
    private PublicKeyController publicKeyController;

    @BeforeEach
    public void setup() {
        mockMvc = MockMvcBuilders.standaloneSetup(publicKeyController).build();
    }

    @Test
    public void TC_AUTH_PUB_KEY_01_GetPublicKeySuccess() throws Exception {
        when(rsaKeyConfig.getPublicKeyBase64()).thenReturn("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...");

        mockMvc.perform(get("/api/v1/auth/public-key"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.publicKey").value("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA..."));
    }
}

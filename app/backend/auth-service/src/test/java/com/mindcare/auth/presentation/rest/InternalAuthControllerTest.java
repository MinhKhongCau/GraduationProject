package com.mindcare.auth.presentation.rest;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mindcare.auth.application.usecase.InternalTokenUseCase;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;

import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@ExtendWith(MockitoExtension.class)
public class InternalAuthControllerTest {

    private MockMvc mockMvc;
    private ObjectMapper objectMapper = new ObjectMapper();

    @Mock
    private InternalTokenUseCase internalTokenUseCase;

    @InjectMocks
    private InternalAuthController internalAuthController;

    @BeforeEach
    public void setup() {
        mockMvc = MockMvcBuilders.standaloneSetup(internalAuthController).build();
    }

    @Test
    public void TC_AUTH_INT_CON_01_Success() throws Exception {
        InternalAuthController.TokenRequest request = new InternalAuthController.TokenRequest("booking-service", "valid-secret");

        when(internalTokenUseCase.issueInternalToken("booking-service", "valid-secret")).thenReturn("m2m-jwt-token-xyz");

        mockMvc.perform(post("/internal/auth/token")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.access_token").value("m2m-jwt-token-xyz"))
                .andExpect(jsonPath("$.token_type").value("Bearer"))
                .andExpect(jsonPath("$.expires_in").value(900));
    }

    @Test
    public void TC_AUTH_INT_CON_02_Unauthorized() throws Exception {
        InternalAuthController.TokenRequest request = new InternalAuthController.TokenRequest("unknown-service", "invalid-secret");

        when(internalTokenUseCase.issueInternalToken("unknown-service", "invalid-secret"))
                .thenThrow(new IllegalArgumentException("Unknown client_id: unknown-service"));

        mockMvc.perform(post("/internal/auth/token")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isUnauthorized())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.error").value("Unauthorized"))
                .andExpect(jsonPath("$.message").value("Unknown client_id: unknown-service"));
    }
}

package com.mindcare.auth.infrastructure.http.controller;

import com.mindcare.auth.application.auth.LoginUseCase;
import com.mindcare.auth.application.auth.LogoutUseCase;
import com.mindcare.auth.application.auth.RegisterUseCase;
import com.mindcare.auth.application.auth.ResendVerificationUseCase;
import com.mindcare.auth.application.auth.VerifyEmailUseCase;
import com.mindcare.auth.application.password.ChangePasswordUseCase;
import com.mindcare.auth.application.password.ForgotPasswordUseCase;
import com.mindcare.auth.application.password.ResetPasswordUseCase;
import com.mindcare.auth.application.profile.UpdateProfileUseCase;
import com.mindcare.auth.application.token.RefreshTokenUseCase;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.mindcare.auth.application.auth.LoginCommand;
import com.mindcare.auth.application.auth.LogoutCommand;
import com.mindcare.auth.application.auth.RegisterCommand;
import com.mindcare.auth.application.auth.LoginResponse;
import com.mindcare.auth.application.common.MessageResponse;
import com.mindcare.auth.domain.account.Role;
import com.mindcare.auth.infrastructure.http.exception.GlobalExceptionHandler;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.validation.beanvalidation.LocalValidatorFactoryBean;

import java.time.LocalDate;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@ExtendWith(MockitoExtension.class)
public class AuthControllerTest {

    private MockMvc mockMvc;
    private ObjectMapper objectMapper = new ObjectMapper().registerModule(new JavaTimeModule());

    @Mock
    private RegisterUseCase registerUseCase;

    @Mock
    private LoginUseCase loginUseCase;

    @Mock
    private RefreshTokenUseCase refreshTokenUseCase;

    @Mock
    private LogoutUseCase logoutUseCase;

    @Mock
    private UpdateProfileUseCase updateProfileUseCase;

    @Mock
    private ChangePasswordUseCase changePasswordUseCase;

    @Mock
    private VerifyEmailUseCase verifyEmailUseCase;

    @Mock
    private ResendVerificationUseCase resendVerificationUseCase;

    @Mock
    private ForgotPasswordUseCase forgotPasswordUseCase;

    @Mock
    private ResetPasswordUseCase resetPasswordUseCase;

    @InjectMocks
    private AuthController authController;

    @BeforeEach
    public void setup() {
        LocalValidatorFactoryBean validator = new LocalValidatorFactoryBean();
        validator.afterPropertiesSet();

        mockMvc = MockMvcBuilders.standaloneSetup(authController)
                .setControllerAdvice(new GlobalExceptionHandler())
                .setValidator(validator)
                .build();
    }

    // ==========================================
    // UC-01: LOGIN CONTROLLER TESTS
    // ==========================================

    @Test
    public void TC_AUTH_LGN_CON_01() throws Exception {
        LoginCommand command = new LoginCommand();
        command.setEmail("user@example.com");
        command.setPassword("Secret123!");

        LoginResponse response = LoginResponse.builder()
                .accessToken("access-token-xyz")
                .refreshToken("refresh-token-xyz")
                .accountId("123e4567-e89b-12d3-a456-426614174000")
                .fullName("John Doe")
                .role(Role.PATIENT)
                .build();

        when(loginUseCase.execute(any(LoginCommand.class))).thenReturn(response);

        mockMvc.perform(post("/api/v1/auth/login")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.success").value(true))
                .andExpect(jsonPath("$.message").value("Login successful!"))
                .andExpect(jsonPath("$.data.accessToken").value("access-token-xyz"))
                .andExpect(jsonPath("$.data.refreshToken").value("refresh-token-xyz"))
                .andExpect(jsonPath("$.data.accountId").value("123e4567-e89b-12d3-a456-426614174000"))
                .andExpect(jsonPath("$.data.fullName").value("John Doe"))
                .andExpect(jsonPath("$.data.role").value("PATIENT"));
    }

    @Test
    public void TC_AUTH_LGN_CON_02() throws Exception {
        LoginCommand command = new LoginCommand();
        command.setEmail(""); // Empty email
        command.setPassword("Secret123!");

        mockMvc.perform(post("/api/v1/auth/login")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value("Invalid input data"))
                .andExpect(jsonPath("$.error.email").value("Email is required"));
    }

    @Test
    public void TC_AUTH_LGN_CON_03() throws Exception {
        LoginCommand command = new LoginCommand();
        command.setEmail("invalid-email-format"); // Invalid format
        command.setPassword("Secret123!");

        mockMvc.perform(post("/api/v1/auth/login")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value("Invalid input data"))
                .andExpect(jsonPath("$.error.email").value("Invalid email format"));
    }

    @Test
    public void TC_AUTH_LGN_CON_04() throws Exception {
        LoginCommand command = new LoginCommand();
        command.setEmail("user@example.com");
        command.setPassword(""); // Empty password

        mockMvc.perform(post("/api/v1/auth/login")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value("Invalid input data"))
                .andExpect(jsonPath("$.error.password").value("Password is required"));
    }

    @Test
    public void TC_AUTH_LGN_CON_05() throws Exception {
        LoginCommand command = new LoginCommand();
        command.setEmail("user@example.com");
        command.setPassword("WrongPassword");

        when(loginUseCase.execute(any(LoginCommand.class)))
                .thenThrow(new RuntimeException("Incorrect password!"));

        mockMvc.perform(post("/api/v1/auth/login")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value("Business validation failed"))
                .andExpect(jsonPath("$.error").value("Incorrect password!"));
    }

    @Test
    public void TC_AUTH_LGN_CON_06() throws Exception {
        LoginCommand command = new LoginCommand();
        command.setEmail("user@example.com");
        command.setPassword("Secret123!");

        when(loginUseCase.execute(any(LoginCommand.class)))
                .thenAnswer(invocation -> {
                    throw new Exception("Checked exception simulation");
                });

        mockMvc.perform(post("/api/v1/auth/login")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isInternalServerError())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value("System error"))
                .andExpect(jsonPath("$.error").value("INTERNAL_SERVER_ERROR"));
    }

    // ==========================================
    // UC-01: LOGOUT CONTROLLER TESTS
    // ==========================================

    @Test
    public void TC_AUTH_LGT_CON_01() throws Exception {
        LogoutCommand command = new LogoutCommand();
        command.setRefreshToken("valid-refresh-token");

        MessageResponse response = new MessageResponse("Logged out successfully!");
        when(logoutUseCase.execute(any(LogoutCommand.class))).thenReturn(response);

        mockMvc.perform(post("/api/v1/auth/logout")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.success").value(true))
                .andExpect(jsonPath("$.message").value("Logout successful!"))
                .andExpect(jsonPath("$.data.message").value("Logged out successfully!"));
    }

    @Test
    public void TC_AUTH_LGT_CON_02() throws Exception {
        LogoutCommand command = new LogoutCommand();
        command.setRefreshToken(""); // Empty token

        mockMvc.perform(post("/api/v1/auth/logout")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value("Invalid input data"))
                .andExpect(jsonPath("$.error.refreshToken").value("Refresh token is required"));
    }

    @Test
    public void TC_AUTH_LGT_CON_03() throws Exception {
        LogoutCommand command = new LogoutCommand();
        command.setRefreshToken("invalid-refresh-token");

        when(logoutUseCase.execute(any(LogoutCommand.class)))
                .thenThrow(new RuntimeException("Session not found!"));

        mockMvc.perform(post("/api/v1/auth/logout")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value("Business validation failed"))
                .andExpect(jsonPath("$.error").value("Session not found!"));
    }

    // ==========================================
    // UC-02: REGISTER CONTROLLER TESTS
    // ==========================================

    @Test
    public void TC_AUTH_REG_CON_01() throws Exception {
        RegisterCommand command = new RegisterCommand();
        command.setFullName("John Doe");
        command.setEmail("john@example.com");
        command.setPassword("Secret123!");
        command.setConfirmPassword("Secret123!");
        command.setRole(Role.PATIENT);
        command.setDateOfBirth(LocalDate.of(1990, 1, 1));

        MessageResponse response = new MessageResponse("User registered successfully!");
        when(registerUseCase.execute(any(RegisterCommand.class))).thenReturn(response);

        mockMvc.perform(post("/api/v1/auth/register")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.success").value(true))
                .andExpect(jsonPath("$.message").value("User registered successfully!"))
                .andExpect(jsonPath("$.data.message").value("User registered successfully!"));
    }

    @Test
    public void TC_AUTH_REG_CON_02() throws Exception {
        RegisterCommand command = new RegisterCommand();
        command.setFullName(""); // Empty name
        command.setEmail("john@example.com");
        command.setPassword("Secret123!");
        command.setConfirmPassword("Secret123!");

        mockMvc.perform(post("/api/v1/auth/register")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value("Invalid input data"))
                .andExpect(jsonPath("$.error.fullName").value("Full name is required"));
    }

    @Test
    public void TC_AUTH_REG_CON_03() throws Exception {
        RegisterCommand command = new RegisterCommand();
        command.setFullName("John Doe");
        command.setEmail(""); // Empty email
        command.setPassword("Secret123!");
        command.setConfirmPassword("Secret123!");

        mockMvc.perform(post("/api/v1/auth/register")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value("Invalid input data"))
                .andExpect(jsonPath("$.error.email").value("Email is required"));
    }

    @Test
    public void TC_AUTH_REG_CON_04() throws Exception {
        RegisterCommand command = new RegisterCommand();
        command.setFullName("John Doe");
        command.setEmail("wrong-format"); // Invalid email
        command.setPassword("Secret123!");
        command.setConfirmPassword("Secret123!");

        mockMvc.perform(post("/api/v1/auth/register")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value("Invalid input data"))
                .andExpect(jsonPath("$.error.email").value("Invalid email format"));
    }

    @Test
    public void TC_AUTH_REG_CON_05() throws Exception {
        RegisterCommand command = new RegisterCommand();
        command.setFullName("John Doe");
        command.setEmail("john@example.com");
        command.setPassword(null); // Null password to trigger @NotBlank
        command.setConfirmPassword("Secret123!");

        mockMvc.perform(post("/api/v1/auth/register")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value("Invalid input data"))
                .andExpect(jsonPath("$.error.password").value("Password is required"));
    }

    @Test
    public void TC_AUTH_REG_CON_06() throws Exception {
        RegisterCommand command = new RegisterCommand();
        command.setFullName("John Doe");
        command.setEmail("john@example.com");
        command.setPassword("12345"); // Less than 6 chars
        command.setConfirmPassword("12345");

        mockMvc.perform(post("/api/v1/auth/register")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value("Invalid input data"))
                .andExpect(jsonPath("$.error.password").value("Password must be at least 6 characters"));
    }

    @Test
    public void TC_AUTH_REG_CON_07() throws Exception {
        RegisterCommand command = new RegisterCommand();
        command.setFullName("John Doe");
        command.setEmail("john@example.com");
        command.setPassword("Secret123!");
        command.setConfirmPassword(""); // Empty confirm password

        mockMvc.perform(post("/api/v1/auth/register")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(command)))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value("Invalid input data"))
                .andExpect(jsonPath("$.error.confirmPassword").value("Confirmation password is required"));
    }
}

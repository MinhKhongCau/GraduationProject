package com.mindcare.auth;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

import io.swagger.v3.oas.annotations.OpenAPIDefinition;
import io.swagger.v3.oas.annotations.servers.Server;

@SpringBootApplication
@OpenAPIDefinition(
		servers = {
				@Server(url = "http://localhost:8000", description = "Local API Gateway (Docker)"),
				@Server(url = "http://localhost:8080", description = "Local Auth Service (Direct)"),
				@Server(url = "https://api.qmcloud.io.vn", description = "Cloud API Gateway (Production)")
		}
)
public class AuthServiceApplication {

	public static void main(String[] args) {
		java.util.TimeZone.setDefault(java.util.TimeZone.getTimeZone("Asia/Ho_Chi_Minh"));
		SpringApplication.run(AuthServiceApplication.class, args);
	}

}

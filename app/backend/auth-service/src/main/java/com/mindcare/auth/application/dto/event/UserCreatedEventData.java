package com.mindcare.auth.application.dto.event;

import com.mindcare.auth.domain.enums.Role;

import java.time.LocalDate;
import java.util.UUID;

public record UserCreatedEventData(
        UUID accountId,
        String email,
        String fullName,
        Role role,
        LocalDate dateOfBirth
) {
}

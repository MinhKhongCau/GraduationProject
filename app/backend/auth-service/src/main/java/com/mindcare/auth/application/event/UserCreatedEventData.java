package com.mindcare.auth.application.event;

import com.mindcare.auth.domain.account.Role;

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

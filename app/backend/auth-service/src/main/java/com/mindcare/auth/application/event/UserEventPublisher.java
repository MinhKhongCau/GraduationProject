package com.mindcare.auth.application.event;

import com.mindcare.auth.application.port.out.EventPublisherPort;
import com.mindcare.auth.domain.account.Account;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

/**
 * Publishes the user.created domain event (RABBITMQ_CONVENTION.md) for a
 * persisted Account. Shared by every code path that creates an account
 * (self-service registration, admin seeding) so the envelope/routing-key
 * are built in exactly one place.
 */
@Component
@RequiredArgsConstructor
public class UserEventPublisher {

    private static final String USER_CREATED_ROUTING_KEY = "user.created";

    private final EventPublisherPort eventPublisherPort;

    @Value("${app.rabbitmq.user-exchange}")
    private String userExchange;

    public void publishUserCreated(Account account) {
        UserCreatedEventData data = new UserCreatedEventData(
                account.getAccountId(),
                account.getEmail(),
                account.getFullName(),
                account.getRole(),
                account.getDateOfBirth()
        );
        eventPublisherPort.publish(userExchange, USER_CREATED_ROUTING_KEY, USER_CREATED_ROUTING_KEY, data);
    }
}

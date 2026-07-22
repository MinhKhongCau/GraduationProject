package com.mindcare.auth.application.port.out;

public interface EventPublisherPort {
    void publish(String exchange, String routingKey, String eventType, Object data);
}

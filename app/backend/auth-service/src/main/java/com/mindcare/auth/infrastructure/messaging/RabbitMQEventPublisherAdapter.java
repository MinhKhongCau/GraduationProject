package com.mindcare.auth.infrastructure.messaging;

import com.mindcare.auth.application.port.out.EventPublisherPort;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.amqp.core.MessageDeliveryMode;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.UUID;

/**
 * Publishes domain events onto RabbitMQ following the envelope defined in
 * RABBITMQ_CONVENTION.md (eventId/eventType/occurredAt/source/data).
 * Publishing is best-effort: a broker outage must not fail the use case that
 * triggered the event, so failures are logged rather than propagated.
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class RabbitMQEventPublisherAdapter implements EventPublisherPort {

    private static final String SOURCE = "auth-service";

    private final RabbitTemplate rabbitTemplate;

    @Override
    public void publish(String exchange, String routingKey, String eventType, Object data) {
        Map<String, Object> envelope = new LinkedHashMap<>();
        envelope.put("eventId", UUID.randomUUID().toString());
        envelope.put("eventType", eventType);
        envelope.put("occurredAt", Instant.now().toString());
        envelope.put("source", SOURCE);
        envelope.put("data", data);

        try {
            rabbitTemplate.convertAndSend(exchange, routingKey, envelope, message -> {
                message.getMessageProperties().setDeliveryMode(MessageDeliveryMode.PERSISTENT);
                return message;
            });
            log.info("Published event type={} routingKey={} exchange={}", eventType, routingKey, exchange);
        } catch (Exception ex) {
            log.error("Failed to publish event type={} routingKey={} exchange={}: {}",
                    eventType, routingKey, exchange, ex.getMessage(), ex);
        }
    }
}

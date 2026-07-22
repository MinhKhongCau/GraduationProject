# RabbitMQ Convention

## Purpose

RabbitMQ is used for asynchronous communication between microservices using an **Event-Driven Architecture (EDA)**.

## Architecture

```text
Producer → Exchange → Queue → Consumer
```

* **Producer** publishes events.
* **Exchange** routes messages.
* **Queue** stores messages.
* **Consumer** processes messages.

## Naming Convention

**Exchange**

```text
<domain>.exchange
```

Examples:

```text
user.exchange
booking.exchange
payment.exchange
```

**Routing Key**

```text
<entity>.<action>
```

Examples:

```text
user.created
booking.cancelled
payment.success
```

**Queue**

```text
<service>.queue
```

Examples:

```text
profile.queue
payment.queue
notification.queue
```

## Event Format

```json
{
  "eventId": "uuid",
  "eventType": "user.created",
  "occurredAt": "2026-07-22T10:00:00Z",
  "source": "auth-service",
  "data": {}
}
```

## Guidelines

* Publish events **only after** a successful database transaction.
* Use **Topic Exchange** as the default exchange type.
* All queues must be **Durable**.
* All messages must be **Persistent**.
* Consumers must use **Manual Acknowledgement (ACK)**.
* Consumers must be **Idempotent** to handle duplicate messages.
* Configure a **Dead Letter Queue (DLQ)** for each queue.
* Retry failed messages before sending them to the DLQ.
* Never include sensitive data (passwords, JWTs, secrets) in event payloads.
* Events should represent **business facts**, not commands.

## Example Flow

```text
Auth Service
      │
Publish: user.created
      │
 user.exchange
      │
 ┌───────────────┐
 │               │
profile.queue  notification.queue
 │               │
Create Profile  Send Welcome Email
```

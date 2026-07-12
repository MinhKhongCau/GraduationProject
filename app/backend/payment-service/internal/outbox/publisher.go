package outbox

import (
	"context"
	"log"
	"payment-service/pkg/rabbitmq"
	"time"
)

type Publisher struct {
	repo Repository
}

func NewPublisher(repo Repository) *Publisher {
	return &Publisher{repo: repo}
}

func (p *Publisher) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	log.Println("⏳ Outbox Publisher Worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Outbox Publisher Worker")
			return
		case <-ticker.C:
			p.publishPendingEvents()
		}
	}
}

func (p *Publisher) publishPendingEvents() {
	events, err := p.repo.GetUnpublishedEvents(20)
	if err != nil {
		log.Printf("Outbox Publisher: Error fetching unpublished events: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	var publishedIDs []string
	for _, event := range events {
		log.Printf("[OUTBOX PUBLISH] Event ID: %s, EventType: %s, Payload: %s", event.ID, event.EventType, event.Payload)

		// TODO: rabbitmq publish
		err := rabbitmq.PublishEvent(event.EventType, event.Payload)
		if err != nil {
			log.Printf("Outbox Publisher: Failed to publish event %s: %v", event.ID, err)
			continue
		}

		publishedIDs = append(publishedIDs, event.ID.String())
	}

	if len(publishedIDs) > 0 {
		err := p.repo.MarkAsPublished(publishedIDs)
		if err != nil {
			log.Printf("Outbox Publisher: Failed to mark events as published in database: %v", err)
		}
	}
}

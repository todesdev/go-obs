package wrappers

import (
	"time"

	"github.com/nats-io/nats.go"
	natscollector "github.com/todesdev/go-obs/internal/metrics/nats_collector"
)

func HandleStreamMessage(next func(msg *nats.Msg)) func(msg *nats.Msg) {
	return func(msg *nats.Msg) {
		startTime := time.Now()
		next(msg)

		elapsedTime := time.Since(startTime)
		natsCollector := natscollector.GetNATSCollector()

		natsCollector.ProcessingDurationObserve(msg.Subject, natscollector.NatsJetStreamMessageType, elapsedTime)
		natsCollector.ProcessedMessagesInc(msg.Subject, natscollector.NatsJetStreamMessageType)
	}
}

func PublishToJetStream(js nats.JetStreamContext, subject string, data []byte) error {
	_, err := js.Publish(subject, data)
	if err != nil {
		return err
	}

	natscollector.GetNATSCollector().PublishedMessagesInc(subject, natscollector.NatsJetStreamMessageType)
	return nil
}

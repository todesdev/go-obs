package wrappers

import (
	"go.opentelemetry.io/otel/propagation"
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

func header(h nats.Header) propagation.HeaderCarrier {
	if h == nil {
		return nil
	}

	// Find total number of values.
	nv := 0
	for _, vv := range h {
		nv += len(vv)
	}

	sv := make([]string, nv) // shared backing array for headers' values
	h2 := make(propagation.HeaderCarrier, len(h))

	for k, vv := range h {
		if vv == nil {
			// Preserve nil values. ReverseProxy distinguishes
			// between nil and zero-length header values.
			h2[k] = nil
			continue
		}

		n := copy(sv, vv)
		h2[k] = sv[:n:n]
		sv = sv[n:]
	}

	return h2
}

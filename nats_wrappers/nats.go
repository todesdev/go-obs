package nats_wrappers

import (
	"context"
	"github.com/todesdev/go-obs/internal/observer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"time"

	"github.com/nats-io/nats.go"
	natscollector "github.com/todesdev/go-obs/internal/metrics/nats_collector"
)

type SubscribeHandler func(msg *nats.Msg, ctxOpts ...context.Context) error

func SubscribeWithObservability(ctx context.Context, stream nats.JetStream, subject, queue string, handler SubscribeHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	sub, err := stream.QueueSubscribeSync(subject, queue, opts...)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := handleSubscription(ctx, sub, handler); err != nil {
			panic(err)
		}
	}()

	return sub, nil
}

func handleSubscription(ctx context.Context, sub *nats.Subscription, handler SubscribeHandler) error {
	for sub.IsValid() {
		natsCollector := natscollector.GetNATSCollector()

		msg, err := sub.NextMsgWithContext(ctx)
		if err != nil {
			return err
		}

		startTime := time.Now()
		subject := msg.Subject

		prop := otel.GetTextMapPropagator()
		headers := propHeader(msg.Header)
		ctx = prop.Extract(ctx, headers)

		obs := observer.ConsumerObserver(ctx, "NATS Consumer:"+subject)
		obs.LogInfo("NATS Consumer: Received new message", zap.String("subject", subject))

		err = handler(msg, obs.Ctx())
		if err != nil {
			elapsedTime := time.Since(startTime)
			natsCollector.ProcessingDurationObserve(subject, natscollector.NatsJetStreamMessageType, elapsedTime)

			obs.RecordErrorWithLogging("Error handling the message", err)

			return err
		}

		elapsedTime := time.Since(startTime)
		natsCollector.ProcessingDurationObserve(subject, natscollector.NatsJetStreamMessageType, elapsedTime)
		natsCollector.PublishedMessagesInc(subject, natscollector.NatsJetStreamMessageType)

		obs.RecordInfoWithLogging("Successfully processed message")

		obs.End()
	}
	return nil
}

func PublishTracedMessage(ctx context.Context, js nats.JetStreamContext, subject string, data []byte) error {
	obs := observer.ProducerObserver(ctx, "NATS Producer:"+subject)
	defer obs.End()

	obs.LogInfo("NATS Producer: Sending message to JetStream", zap.String("subject", subject))

	_, err := js.PublishMsg(newMsg(obs.Ctx(), subject, data))
	if err != nil {
		obs.RecordErrorWithLogging("Error sending message to JetStream", err)
		return err
	}

	obs.RecordInfoWithLogging("Sent message to JetStream")

	natscollector.GetNATSCollector().PublishedMessagesInc(subject, natscollector.NatsJetStreamMessageType)
	return nil
}

func newMsg(ctx context.Context, subject string, data []byte) *nats.Msg {
	prop := otel.GetTextMapPropagator()
	headers := make(propagation.HeaderCarrier)
	prop.Inject(ctx, headers)

	return &nats.Msg{
		Subject: subject,
		Header:  natsHeader(headers),
		Data:    data,
	}
}

func natsHeader(h propagation.HeaderCarrier) nats.Header {
	if h == nil {
		return nil
	}

	// Find total number of values.
	nv := 0
	for _, vv := range h {
		nv += len(vv)
	}

	sv := make([]string, nv) // shared backing array for headers' values
	h2 := make(nats.Header, len(h))

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

func propHeader(h nats.Header) propagation.HeaderCarrier {
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
			// between nil and zero-length propHeader values.
			h2[k] = nil
			continue
		}

		n := copy(sv, vv)
		h2[k] = sv[:n:n]
		sv = sv[n:]
	}

	return h2
}

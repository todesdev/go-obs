package nats_wrappers

import (
	"context"
	goobs "github.com/todesdev/go-obs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"time"

	"github.com/nats-io/nats.go"
	natscollector "github.com/todesdev/go-obs/internal/metrics/nats_collector"
)

type SubscribeHandler func(msg *nats.Msg, ctxOpts ...context.Context) error

func SubscribeWithObservability(ctx context.Context, stream nats.JetStream, subject, consumer string, handler SubscribeHandler, opts ...nats.SubOpt) {
	sub, err := stream.QueueSubscribeSync(subject, consumer, opts...)
	if err != nil {
		panic(err)
	}

	err = handleSubscription(ctx, sub, handler)
	if err != nil {
		panic(err)
	}
}

func handleSubscription(ctx context.Context, sub *nats.Subscription, handler SubscribeHandler) error {
	for {
		select {
		case <-ctx.Done():
			return sub.Unsubscribe()
		default:
		}

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

		processName := "NATS:" + subject

		c, span := goobs.NewConsumerTrace(ctx, processName)
		logger := goobs.TracedLoggerWithProcess(span, processName)

		logger.Info("NATS Consumer: Received new message")

		err = handler(msg, c)
		if err != nil {
			elapsedTime := time.Since(startTime)
			natsCollector.ProcessingDurationObserve(subject, natscollector.NatsJetStreamMessageType, elapsedTime)

			errMsg := "Error handling the message"
			span.RecordError(err)
			logger.Error(errMsg, zap.Error(err))

			return err
		}

		elapsedTime := time.Since(startTime)
		natsCollector.ProcessingDurationObserve(subject, natscollector.NatsJetStreamMessageType, elapsedTime)
		natsCollector.PublishedMessagesInc(subject, natscollector.NatsJetStreamMessageType)

		span.SetStatus(codes.Ok, "Successfully processed message")
		logger.Info("Successfully processed message")
	}
}

func PublishTracedMessage(ctx context.Context, js nats.JetStreamContext, subject string, data []byte) error {
	c, span := goobs.NewProducerTrace(ctx, "NATS Producer: Sending message to JetStream")
	_, err := js.PublishMsg(newMsg(c, subject, data))
	if err != nil {
		return err
	}

	span.SetStatus(codes.Ok, "Sent message to JetStream")

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

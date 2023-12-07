package natscollector

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/todesdev/go-obs/internal/logging"
)

const (
	NatsSubsystem = "nats"

	NatsProcessedMessagesTotal    = "processed_messages_total"
	NatsMessageProcessingDuration = "message_processing_duration_seconds"
	NatsPublishedMessagesTotal    = "published_messages_total"

	NatsMessagesTotalHelp             = "Total number of NATS messages processed."
	NatsMessageProcessingDurationHelp = "Duration of NATS message processing."
	NatsPublishedMessagesHelp         = "Total number of NATS messages published."

	NatsSubjectLabel = "subject"
	NatsTypeLabel    = "type"

	NatsSimpleMessageType    = "simple"
	NatsJetStreamMessageType = "jetstream"
)

var natsCollector *NATSCollector

type NATSCollector struct {
	processedMessages  *prometheus.CounterVec
	processingDuration *prometheus.HistogramVec
	publishedMessages  *prometheus.CounterVec
}

func newNATSCollector(serviceName string) *NATSCollector {
	processedMessages := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: prometheus.BuildFQName(serviceName, NatsSubsystem, NatsProcessedMessagesTotal),
			Help: NatsMessagesTotalHelp,
		},
		[]string{NatsTypeLabel, NatsSubjectLabel},
	)

	processingDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    prometheus.BuildFQName(serviceName, NatsSubsystem, NatsMessageProcessingDuration),
			Help:    NatsMessageProcessingDurationHelp,
			Buckets: prometheus.DefBuckets,
		},
		[]string{NatsTypeLabel, NatsSubjectLabel},
	)

	publishedMessages := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: prometheus.BuildFQName(serviceName, NatsSubsystem, NatsPublishedMessagesTotal),
			Help: NatsPublishedMessagesHelp,
		},
		[]string{NatsTypeLabel, NatsSubjectLabel},
	)

	natsCollector = &NATSCollector{
		processedMessages:  processedMessages,
		processingDuration: processingDuration,
		publishedMessages:  publishedMessages,
	}

	return natsCollector
}

func (collector *NATSCollector) Register(registry *prometheus.Registry) {
	registry.MustRegister(collector.processedMessages)
	registry.MustRegister(collector.processingDuration)
	registry.MustRegister(collector.publishedMessages)
}

func Setup(registry *prometheus.Registry, serviceName string) {
	logger := logging.LoggerWithProcess("NatsCollectorSetup")
	logger.Info("Setting up NATS collector")

	newNATSCollector(serviceName).Register(registry)

	logger.Info("NATS collector setup complete")
}

func GetNATSCollector() *NATSCollector {
	return natsCollector
}

func (collector *NATSCollector) ProcessedMessagesInc(subject string, messageType string) {
	collector.processedMessages.WithLabelValues(messageType, subject).Inc()
}

func (collector *NATSCollector) ProcessingDurationObserve(subject string, messageType string, duration time.Duration) {
	collector.processingDuration.WithLabelValues(messageType, subject).Observe(float64(duration) / float64(time.Second))
}

func (collector *NATSCollector) PublishedMessagesInc(subject string, messageType string) {
	collector.publishedMessages.WithLabelValues(messageType, subject).Inc()
}

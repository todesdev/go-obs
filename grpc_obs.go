package goobs

import (
	"github.com/todesdev/go-obs/internal/logging"
	"github.com/todesdev/go-obs/internal/observer"
	"github.com/todesdev/go-obs/internal/tracing"
	"go.uber.org/zap"
)

type GRPCObserverConfig struct {
	ServiceName      string
	ServiceVersion   string
	Region           string
	LogLevel         string
	OTLPGRPCEndpoint string
	TracingEnabled   bool
}

func InitializeGRPCObserver(cfg *GRPCObserverConfig) error {
	validatedConfig, err := validateGRPCObserverConfig(cfg)
	if err != nil {
		return err
	}

	logging.Setup(validatedConfig.Region, validatedConfig.ServiceName, validatedConfig.ServiceVersion, validatedConfig.LogLevel)
	logger := logging.LoggerWithProcess("observability:initialize")
	logger.Info("Logger setup complete")

	if validatedConfig.TracingEnabled {
		observer.SetTracingEnabled(true)
		res, err := registerResource(validatedConfig.ServiceName, validatedConfig.ServiceVersion, validatedConfig.Region)
		if err != nil {
			logger.Error("Failed to register resource", zap.Error(err))
			return err
		}

		if validatedConfig.OTLPGRPCEndpoint != "" {
			err := tracing.SetupOtlpGrpcTracer(validatedConfig.OTLPGRPCEndpoint, validatedConfig.ServiceName, res)
			if err != nil {
				logger.Error("Failed to setup OTLP GRPC tracer", zap.Error(err))
				logger.Info("Switching to STDOUT tracer")
				err = tracing.SetupStdOutTracer(validatedConfig.ServiceName, res)
				if err != nil {
					logger.Error("Failed to setup STDOUT tracer", zap.Error(err))
					return err
				}

				logger.Info("Tracing exporter set to STDOUT")
			}

			logger.Info("Tracing exporter set to OTLP GRPC")
		} else {
			err := tracing.SetupStdOutTracer(validatedConfig.ServiceName, res)
			if err != nil {
				logger.Error("Failed to setup STDOUT tracer", zap.Error(err))
				return err
			}

			logger.Info("Tracing exporter set to STDOUT")
		}

		logger.Info("Tracing setup complete")
	} else {
		logger.Warn("Tracing is disabled")
	}

	return nil
}

func validateGRPCObserverConfig(cfg *GRPCObserverConfig) (*GRPCObserverConfig, error) {
	var validatedConfig GRPCObserverConfig

	validatedConfig.ServiceName = cfg.ServiceName
	validatedConfig.ServiceVersion = cfg.ServiceVersion
	validatedConfig.Region = cfg.Region
	validatedConfig.LogLevel = cfg.LogLevel
	validatedConfig.TracingEnabled = cfg.TracingEnabled

	validatedConfig.OTLPGRPCEndpoint = cfg.OTLPGRPCEndpoint

	return &validatedConfig, nil
}

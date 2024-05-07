package goobs

import (
	"github.com/todesdev/go-obs/internal/logging"
	"github.com/todesdev/go-obs/internal/tracing"
)

type GRPCObserverConfig struct {
	ServiceName      string
	ServiceVersion   string
	Region           string
	OTLPGRPCEndpoint string
	TracingEnabled   bool
}

func InitializeGRPCObserver(cfg *GRPCObserverConfig) error {
	validatedConfig, err := validateGRPCObserverConfig(cfg)
	if err != nil {
		return err
	}

	logging.Setup(validatedConfig.Region, validatedConfig.ServiceName, validatedConfig.ServiceVersion)

	if validatedConfig.TracingEnabled {
		res, err := registerResource(validatedConfig.ServiceName, validatedConfig.ServiceVersion, validatedConfig.Region)
		if err != nil {
			return err
		}

		if validatedConfig.OTLPGRPCEndpoint != "" {
			err := tracing.SetupOtlpGrpcTracer(validatedConfig.OTLPGRPCEndpoint, validatedConfig.ServiceName, res)
			if err != nil {
				return err
			}
		} else {
			err := tracing.SetupStdOutTracer(validatedConfig.ServiceName, res)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func validateGRPCObserverConfig(cfg *GRPCObserverConfig) (*GRPCObserverConfig, error) {
	var validatedConfig GRPCObserverConfig

	validatedConfig.ServiceName = cfg.ServiceName
	validatedConfig.ServiceVersion = cfg.ServiceVersion
	validatedConfig.Region = cfg.Region
	validatedConfig.TracingEnabled = cfg.TracingEnabled

	validatedConfig.OTLPGRPCEndpoint = cfg.OTLPGRPCEndpoint

	return &validatedConfig, nil
}

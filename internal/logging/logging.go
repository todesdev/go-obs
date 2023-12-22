package logging

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var logger *zap.Logger

func Setup(region, serviceName, serviceVersion string) {
	rawJSON := []byte(`{
		  "level": "info",
		  "encoding": "json",
		  "outputPaths": ["stdout"],
		  "errorOutputPaths": ["stderr"],
		  "encoderConfig": {
		    "levelKey": "level",
		    "timeKey": "timestamp",
			"regionKey": "region",
		    "serviceKey": "service",
			"versionKey": "version",
		    "instanceIdKey": "instanceID",
		    "messageKey": "message",
		    "levelEncoder": "lowercase",
		    "timeEncoder": "epoch"
		  }
		}`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	var err error

	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}

	instanceID, err := getInstanceIP()
	if err != nil {
		instanceID = generateRandomInstanceID()
	}

	logger = logger.With(
		zap.String("region", region),
		zap.String("service", serviceName),
		zap.String("version", serviceVersion),
		zap.String("instanceID", instanceID),
	)
}

func getLogger() *zap.Logger {
	return logger
}

func getInstanceIP() (string, error) {
	instanceIP := os.Getenv("INSTANCE_IP")
	if instanceIP == "" {
		return "", errors.New("INSTANCE_IP not found")
	}

	return instanceIP, nil
}

func generateRandomInstanceID() string {
	return uuid.New().String()[:7]
}

package logging

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var logger *zap.Logger

func Setup(serviceName string) {
	rawJSON := []byte(`{
		  "level": "info",
		  "encoding": "json",
		  "outputPaths": ["stdout"],
		  "errorOutputPaths": ["stderr"],
		  "encoderConfig": {
		    "levelKey": "level",
		    "timeKey": "timestamp",
		    "serviceKey": "service",
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

	logger = logger.With(zap.String("service", serviceName), zap.String("instanceID", instanceID))
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

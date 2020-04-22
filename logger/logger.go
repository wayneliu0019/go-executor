package logger

import (
	"log"
	"sync"

	"github.com/spf13/viper"

	"go.uber.org/zap"
)

var instance *zap.Logger
var once sync.Once

// GetInstance initializes a logger instance (if needed) and returns it
func GetInstance() *zap.Logger {
	once.Do(func() {
		prodConfig := zap.NewProductionConfig()
		prodConfig.DisableStacktrace = true

		level:=viper.GetString("logging_level")
		switch level {
		case "info":
			prodConfig.Level.SetLevel(zap.InfoLevel)
		case "debug":
			prodConfig.Level.SetLevel(zap.DebugLevel)
		case "warn":
			prodConfig.Level.SetLevel(zap.WarnLevel)
		case "error":
			prodConfig.Level.SetLevel(zap.ErrorLevel)
		default:
			prodConfig.Level.SetLevel(zap.InfoLevel)
		}

		prodConfig.OutputPaths = []string{"/root/out.log"}
		prodConfig.ErrorOutputPaths = []string{"/root/err.log"}
		prod, err := prodConfig.Build()
		if err != nil {
			log.Fatalf("Error while initializing production logger: %v", err)
		}

		instance = prod
	})

	return instance
}

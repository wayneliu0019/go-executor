package logger

import (
	"fmt"
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

		log_dir := viper.GetString("log_dir")
		if len(log_dir) <=0 {
			log_dir = "/root"
		}
		prodConfig.OutputPaths = []string{log_dir+"/out.log"}
		prodConfig.ErrorOutputPaths = []string{log_dir+"/err.log"}

		prod, err := prodConfig.Build()
		if err != nil {
			fmt.Println("Error while initializing production logger: %v", err)
		}

		instance = prod
	})

	return instance
}

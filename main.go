package main

import (
	"fmt"
	"time"

	"go-mesos-executor/container"
	"go-mesos-executor/executor"
	"go-mesos-executor/hook"
	"go-mesos-executor/logger"

	"github.com/mesos/mesos-go/api/v1/lib/executor/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	agentEndpoint           string
	dockerSocket            string
	executorID              string
	frameworkID             string
	logDir                  string
	loggingLevel            string
)

var rootCmd = &cobra.Command{
	Use:   "mesos-docker-executor",
	Short: "Custom Mesos Docker executor",
	Run: func(cmd *cobra.Command, args []string) {
		logger.GetInstance().Info("Initializing the executor",
			zap.String("executorID", executorID),
			zap.String("frameworkID", frameworkID),
                        zap.String("agentEndpoint", agentEndpoint),
		)

		// Prepare docker containerizer
		c, err := container.NewDockerContainerizer(dockerSocket)
		if err != nil {
			logger.GetInstance().Fatal("An error occured while initializing the containerizer",
				zap.Error(err),
			)
		}

		// Create hook manager
		hooks := viper.GetStringSlice("hooks")
		logger.GetInstance().Info("Creating hook manager",
			zap.Reflect("hooks", hooks),
		)
		m := hook.NewManager(hooks)
		m.RegisterHooks(&hook.ACLHook)
		m.RegisterHooks(&hook.IptablesHook)
		m.RegisterHooks(&hook.NetnsHook)
		m.RegisterHooks(&hook.RemoveContainerHook)
		m.RegisterHooks(&hook.NetworkHook)

		// Create and run the executor
		config := config.Config{AgentEndpoint: agentEndpoint, ExecutorID: executorID, FrameworkID: frameworkID}
		e := executor.NewExecutor(config, c, m)

		if err := e.Execute(); err != nil {
			fmt.Println("An error occured while running the executor, %v", err)
			logger.GetInstance().Fatal("An error occured while running the executor",
				zap.Error(err),
			)
		}
	},
}

func init() {
	cobra.OnInitialize(readConfig)

	// Flags given by the agent when running th executor
	rootCmd.PersistentFlags().StringVar(&dockerSocket, "docker_socket", "/var/run/docker.sock", "Docker socket path")
	rootCmd.PersistentFlags().StringVar(&logDir, "log_dir", "", "Location to put log files")
	rootCmd.PersistentFlags().StringVar(&loggingLevel, "logging_level", "", "Logging level")

	// Custom flags
	rootCmd.PersistentFlags().Bool("debug", true, "Enable debug mode")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	rootCmd.PersistentFlags().StringSlice("hooks", []string{"ddfd"}, "Enabled hooks")
	viper.BindPFlag("hooks", rootCmd.PersistentFlags().Lookup("hooks"))

	rootCmd.PersistentFlags().String("proc_path", "/proc", "Proc mount path")
	viper.BindPFlag("proc_path", rootCmd.PersistentFlags().Lookup("proc_path"))

	rootCmd.PersistentFlags().Duration("registering_retry", 100*time.Millisecond, "Executor registering delay in duration")
	viper.BindPFlag("registering_retry", rootCmd.PersistentFlags().Lookup("registering_retry"))


	// Iptables hook
	viper.SetDefault("iptables.ip_forwarding", true)
	viper.SetDefault("iptables.ip_masquerading", true)

	hooks := viper.GetStringSlice("hooks")
	logger.GetInstance().Info(fmt.Sprintf("at the end of init, hooks are %v", hooks))

}

func readConfig() {
	viper.SetEnvPrefix("mesos")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	viper.BindEnv("agent_endpoint")
	agentEndpoint = viper.GetString("agent_endpoint")

	viper.BindEnv("executor_id")
	executorID = viper.GetString("executor_id")

	viper.BindEnv("framework_id")
	frameworkID = viper.GetString("framework_id")

	if err := viper.ReadInConfig(); err != nil {
		logger.GetInstance().Fatal("An error occured while reading the configuration file",
			zap.Error(err),
		)
	}

	hooks := viper.GetStringSlice("hooks")
	logger.GetInstance().Info(fmt.Sprintf("readInConfig, hooks are %v", hooks))

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.GetInstance().Fatal("An error occured while running the root command",
			zap.Error(err),
		)
	}
}

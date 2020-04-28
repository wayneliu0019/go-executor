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
	containerdSocket        string
	namespace               string
	image                   string
	command                 string

	executorID              string
	frameworkID             string
	agentEndpoint           string
)

var rootCmd = &cobra.Command{
	Use:   "mesos-containerd-executor",
	Short: "Custom Mesos Containerd executor",
	Run: func(cmd *cobra.Command, args []string) {
		logger.GetInstance().Info("Initializing the executor",
			zap.String("executorID", executorID),
			zap.String("frameworkID", frameworkID),
			zap.String("agentEndpoint", agentEndpoint),
		)

		// Prepare containerd containerizer
		c, err := container.NewContainerdContainerizer(containerdSocket, image, namespace, command)
		if err != nil {
			logger.GetInstance().Fatal("An error occurred while initializing the containerizer",
				zap.Error(err),
			)
		}

		// Create hook manager
		hooks := viper.GetStringSlice("hooks")
		logger.GetInstance().Info("Creating hook manager",
			zap.Reflect("hooks", hooks),
		)
		m := hook.NewManager(hooks)
		//m.RegisterHooks(&hook.ACLHook)
		//m.RegisterHooks(&hook.IptablesHook)
		//m.RegisterHooks(&hook.NetnsHook)
		m.RegisterHooks(&hook.RemoveContainerHook)
		//m.RegisterHooks(&hook.NetworkHook)

		// Create and run the executor
		config := config.Config{AgentEndpoint: agentEndpoint, ExecutorID: executorID, FrameworkID: frameworkID}
		e := executor.NewExecutor(config, c, m)

		if err := e.Execute(); err != nil {
			logger.GetInstance().Fatal("An error occurred while running the executor",
				zap.Error(err),
			)
		}
	},
}

func init() {
	//the readConfig function will not be ran directly, it will be invoked when execute "cmd.Excute"
	cobra.OnInitialize(readConfig)

	// Flags given by the agent when running th executor
	rootCmd.PersistentFlags().StringVar(&containerdSocket, "containerd_socket", "/run/containerd/containerd.sock", "Containerd socket path")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "default", "Containerd namespace that will be used ")
	rootCmd.PersistentFlags().StringVar(&image, "image", "", "Image that will be used to create container")
	rootCmd.PersistentFlags().StringVar(&command, "command", "", "The command that will be executed after container startup")

	// Custom flags
	rootCmd.PersistentFlags().StringSlice("hooks", []string{}, "Enabled hooks")
	viper.BindPFlag("hooks", rootCmd.PersistentFlags().Lookup("hooks"))

	rootCmd.PersistentFlags().Duration("registering_retry", 100*time.Millisecond, "Executor registering delay in duration")
	viper.BindPFlag("registering_retry", rootCmd.PersistentFlags().Lookup("registering_retry"))


	// Iptables hook
	viper.SetDefault("iptables.ip_forwarding", true)
	viper.SetDefault("iptables.ip_masquerading", true)

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
		fmt.Println("An error occured while reading the configuration file %v", err)
	}

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.GetInstance().Fatal("An error occurred while running the root command",
			zap.Error(err),
		)
	}
}

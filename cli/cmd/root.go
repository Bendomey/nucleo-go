package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/Bendomey/nucleo-go/nucleo"
	"github.com/Bendomey/nucleo-go/nucleo/broker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "nucli",
	Short: "Nucleo Go CLI",
	Long:  `Nucleo CLI allows you to control the lifecycle of your services.`,
}

type RunOpts struct {
	Config       *nucleo.Config
	StartHandler func(*broker.ServiceBroker, *cobra.Command)
}

var UserOpts *RunOpts

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(runOpts RunOpts) {
	UserOpts = &runOpts
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is <app>/.nucleo-config.yaml)")

	// RootCmd.PersistentFlags().StringVar(&UserOpts.Config.LogLevel, "log", "l", "Log Level - fatal, error, debug, trace")
	// viper.BindPFlag("log", RootCmd.PersistentFlags().Lookup("log"))

	// RootCmd.PersistentFlags().StringVar(&UserOpts.Config.LogFormat, "logFormat", "lf", "Log Format - Options: JSON or TEXT")
	// viper.BindPFlag("logFormat", RootCmd.PersistentFlags().Lookup("logFormat"))

	// RootCmd.PersistentFlags().StringVar(&UserOpts.Config.Transporter, "transporter", "t", "Transporter")
	// viper.BindPFlag("transporter", RootCmd.PersistentFlags().Lookup("transporter"))

	// RootCmd.PersistentFlags().StringVar(&UserOpts.Config.Namespace, "namespace", "n", "Namespace")
	// viper.BindPFlag("namespace", RootCmd.PersistentFlags().Lookup("namespace"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find executable directory.
		basePath, err := os.Executable()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(path.Dir(basePath))
		viper.AddConfigPath(".")
		viper.SetConfigName("nucleo-config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("Error loading config - Error: ", err)
	}
}

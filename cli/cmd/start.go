package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bendomey/nucleo-go/nucleo"
	"github.com/Bendomey/nucleo-go/nucleo/broker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var environment string

// startCmd starts the service broker
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the service broker.",
	Long:  `starts the service broker and publish all microservices added.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start called - UserOpts.Config -> ", UserOpts.Config)
		if UserOpts == nil {
			panic("No options set!")
		}

		argsConfig := argsToConfig(cmd)
		services := viper.GetStringMap("services")
		if services != nil {
			UserOpts.Config.Services = services
		}
		broker := broker.New(UserOpts.Config, argsConfig)

		signalC := make(chan os.Signal)
		signal.Notify(signalC, os.Interrupt, syscall.SIGTERM)

		UserOpts.StartHandler(broker, cmd)
		<-signalC
		broker.Stop()
	},
}

// argsToConfig read args sent ot the CLI and populate a molecule config.
func argsToConfig(cmd *cobra.Command) *nucleo.Config {
	return &nucleo.Config{}
}

func init() {
	RootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	startCmd.PersistentFlags().StringVarP(&environment, "env", "e", "ENV", "Environment name.")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// helloCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

package cli

import (
	"github.com/Bendomey/nucleo-go/nucleo"
	"github.com/Bendomey/nucleo-go/nucleo/broker"
	"github.com/Bendomey/nucleo-go/nucleo/cli/cmd"
	"github.com/spf13/cobra"
)

// Start parse the config from the cli args. creates a service broker and pass down to the startHandler.
func Start(config *nucleo.Config, startHandler func(*broker.ServiceBroker, *cobra.Command)) {
	cmd.Execute(cmd.RunOpts{Config: config, StartHandler: startHandler})
}

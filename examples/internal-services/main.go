package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bendomey/nucleo-go"
	"github.com/Bendomey/nucleo-go/broker"
)

func main() {
	nucleoBroker := broker.New(&nucleo.Config{
		Namespace: "basic-example",
		LogLevel:  nucleo.LogLevelInfo,
	})

	// list all services here
	nucleoBroker.PublishServices()
	nucleoBroker.Start()

	additionResult := <-nucleoBroker.Call("$node.services", map[string]interface{}{
		"onlyAvailable": true,
		"withActions":   true,
	})
	fmt.Print("additionResult", additionResult)

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt, syscall.SIGTERM)

	<-signalC

	nucleoBroker.Stop()
}

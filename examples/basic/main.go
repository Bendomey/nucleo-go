package main

import (
	"github.com/Bendomey/nucleo-go/nucleo"
	"github.com/Bendomey/nucleo-go/nucleo/broker"
)

func main() {
	nucleoBroker := broker.New(nucleo.Config{
		Namespace: "basic-example",
		LogFormat: nucleo.LogFormatJSON,
	})

	// list all services here
	nucleoBroker.PublishServices()

	nucleoBroker.Start()

	nucleoBroker.Stop()
}

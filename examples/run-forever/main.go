package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bendomey/nucleo-go"
	"github.com/Bendomey/nucleo-go/broker"
)

var Calculator = nucleo.ServiceSchema{
	Name:     "calculator",
	Settings: map[string]interface{}{},
	Actions: []nucleo.Action{
		{
			Name:        "add",
			Description: "add two numbers",
			Params: map[string]interface{}{
				"a": "number",
				"b": "number",
			},
			Handler: func(ctx nucleo.Context, params nucleo.Payload) interface{} {
				ctx.Logger().Infoln("add action called")

				return params.Get("a").Int() + params.Get("b").Int()
			},
		},
	},
}

func main() {
	nucleoBroker := broker.New(&nucleo.Config{
		Namespace: "basic-example",
		LogLevel:  nucleo.LogLevelInfo,
	})

	// list all services here
	nucleoBroker.PublishServices(Calculator)
	nucleoBroker.Start()

	additionResult := <-nucleoBroker.Call("calculator.add", map[string]interface{}{
		"a": 1,
		"b": "dkjvbsduivf",
	})

	signalC := make(chan os.Signal)
	signal.Notify(signalC, os.Interrupt, syscall.SIGTERM)

	fmt.Print("additionResult", additionResult)
	<-signalC

	nucleoBroker.Stop()
}

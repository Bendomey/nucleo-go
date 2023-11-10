package main

import (
	"fmt"

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
			Handler: func(ctx nucleo.Context, params nucleo.Payload) interface{} {
				ctx.Logger().Info("add action called")

				return params.Get("a").Int() + params.Get("b").Int()
			},
		},
	},
}

func main() {
	nucleoBroker := broker.New(&nucleo.Config{
		Namespace: "basic-example",
		LogFormat: nucleo.LogFormatJSON,
		LogLevel:  nucleo.LogLevelDebug,
	})

	// list all services here
	nucleoBroker.PublishServices(Calculator)
	nucleoBroker.Start()

	additionResult := <-nucleoBroker.Call("calculator.add", map[string]interface{}{
		"a": 1,
		"b": 2,
	})

	fmt.Print("additionResult", additionResult)

	nucleoBroker.Stop()
}

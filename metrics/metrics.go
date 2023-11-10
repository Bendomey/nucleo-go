package metrics

import (
	"fmt"
	"time"

	"github.com/Bendomey/nucleo-go"
	"github.com/Bendomey/nucleo-go/context"
	"github.com/Bendomey/nucleo-go/middleware"
)

func metricEnd(brokerContext nucleo.BrokerContext, result nucleo.Payload) {
	ctx := brokerContext.(*context.Context)
	if !ctx.Meta().Get("startTime").Exists() {
		return
	}

	startTime := ctx.Meta().Get("startTime").Time()
	payload := metricsPayload(brokerContext)

	mlseconds := float64(time.Since(startTime).Nanoseconds()) / 1000000
	payload["duration"] = mlseconds
	payload["endTime"] = time.Now().Format(time.RFC3339)
	if result.IsError() {
		payload["error"] = map[string]string{
			"message": fmt.Sprintf("%s", result.Error()),
		}
	}
	ctx.Emit("metrics.trace.span.finish", payload)
}

func metricStart(context nucleo.BrokerContext) {
	meta := context.Meta().Add("startTime", time.Now()).Add("duration", 0)
	context.UpdateMeta(meta)
	context.Emit("metrics.trace.span.start", metricsPayload(context))
}

// metricsPayload generate the payload for the metrics event
func metricsPayload(brokerContext nucleo.BrokerContext) map[string]interface{} {
	rawContext := brokerContext.(*context.Context)
	contextMap := brokerContext.AsMap()
	if rawContext.Meta().Get("startTime").Exists() {
		contextMap["startTime"] = rawContext.Meta().Get("startTime").Time().Format(time.RFC3339)
	}
	nodeID := rawContext.BrokerDelegates().LocalNode().GetID()
	contextMap["nodeID"] = nodeID
	if rawContext.SourceNodeID() == nodeID {
		contextMap["remoteCall"] = false
	} else {
		contextMap["remoteCall"] = true
		contextMap["callerNodeID"] = rawContext.SourceNodeID()
	}
	_, isAction := contextMap["action"]
	if isAction {
		action := contextMap["action"].(string)
		svcs := rawContext.BrokerDelegates().ServiceForAction(action)
		contextMap["action"] = map[string]string{"name": action}
		contextMap["service"] = map[string]string{"name": svcs[0].Name, "version": svcs[0].Version}
	}
	return contextMap
}

// shouldMetric check if it should metric for this context.
func createShouldMetric(Config nucleo.Config) func(context nucleo.BrokerContext) bool {
	var callsCount float32 = 0
	return func(context nucleo.BrokerContext) bool {
		if context.Meta().Get("tracing").Bool() {
			callsCount++
			if callsCount*Config.MetricsRate >= 1.0 {

				callsCount = 0
				return true
			}
		}
		return false
	}
}

// Middleware create a metrics middleware
func Middlewares() nucleo.Middlewares {
	var Config = nucleo.DefaultConfig
	shouldMetric := createShouldMetric(Config)
	return map[string]nucleo.MiddlewareHandler{
		// store the broker config
		"Config": func(params interface{}, next func(...interface{})) {
			Config = params.(nucleo.Config)
			shouldMetric = createShouldMetric(Config)
			next()
		},
		"afterLocalAction": func(params interface{}, next func(...interface{})) {
			payload := params.(middleware.AfterActionParams)
			context := payload.BrokerContext
			result := payload.Result
			if shouldMetric(context) {
				metricEnd(context, result)
			}
			next()
		},
		"beforeLocalAction": func(params interface{}, next func(...interface{})) {
			context := params.(nucleo.BrokerContext)
			if shouldMetric(context) {
				metricStart(context)
			}
			next()
		},
	}
}

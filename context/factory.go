package context

import (
	"errors"
	"fmt"

	"github.com/Bendomey/nucleo-go"
	"github.com/Bendomey/nucleo-go/payload"
	"github.com/Bendomey/nucleo-go/utils"
	log "github.com/sirupsen/logrus"
)

type Context struct {
	id           string
	requestID    string
	broker       *nucleo.BrokerDelegates
	targetNodeID string
	sourceNodeID string
	parentID     string
	actionName   string
	eventName    string
	groups       []string
	broadcast    bool
	params       nucleo.Payload
	meta         nucleo.Payload
	timeout      int
	level        int
	caller       string
}

func BrokerContext(broker *nucleo.BrokerDelegates) nucleo.BrokerContext {
	localNodeID := broker.LocalNode().GetID()
	id := fmt.Sprint("rootContext-broker-", localNodeID, "-", utils.RandomString(12))
	context := Context{
		id:       id,
		broker:   broker,
		level:    1,
		parentID: "ImGroot;)",
		meta:     payload.Empty(),
	}
	return &context
}

// ChildEventContext : create a child context for a specific event call.
func (context *Context) ChildEventContext(eventName string, params nucleo.Payload, groups []string, broadcast bool) nucleo.BrokerContext {
	parentContext := context
	meta := parentContext.meta
	if context.broker.Config.Metrics {
		meta = meta.Add("tracing", true)
	}
	id := utils.RandomString(12)
	var requestID string
	if parentContext.requestID != "" {
		requestID = parentContext.requestID
	} else {
		requestID = id
	}
	caller := parentContext.actionName
	if parentContext.eventName != "" {
		caller = parentContext.eventName
	}
	eventContext := Context{
		id:        id,
		requestID: requestID,
		broker:    parentContext.broker,
		eventName: eventName,
		groups:    groups,
		params:    params,
		broadcast: broadcast,
		level:     parentContext.level + 1,
		meta:      meta,
		parentID:  parentContext.id,
		caller:    caller,
	}
	return &eventContext
}

// Config return the broker config attached to this context.
func (context *Context) BrokerDelegates() *nucleo.BrokerDelegates {
	return context.broker
}

// ChildActionContext : create a child context for a specific action call.
func (context *Context) ChildActionContext(actionName string, params nucleo.Payload, opts ...nucleo.Options) nucleo.BrokerContext {
	parentContext := context
	meta := parentContext.meta
	if context.broker.Config.Metrics {
		meta = meta.Add("tracing", true)
	}
	if len(opts) > 0 && opts[0].Meta != nil && opts[0].Meta.Len() > 0 {
		meta = meta.AddMany(opts[0].Meta.RawMap())
	}
	id := utils.RandomString(12)
	var requestID string
	if parentContext.requestID != "" {
		requestID = parentContext.requestID
	} else {
		requestID = id
	}
	caller := parentContext.actionName
	if parentContext.eventName != "" {
		caller = parentContext.eventName
	}
	actionContext := Context{
		id:         id,
		requestID:  requestID,
		broker:     parentContext.broker,
		actionName: actionName,
		params:     params,
		level:      parentContext.level + 1,
		meta:       meta,
		parentID:   parentContext.id,
		caller:     caller,
	}
	return &actionContext
}

// ActionContext create an action context for remote call.
func ActionContext(broker *nucleo.BrokerDelegates, values map[string]interface{}) nucleo.BrokerContext {
	var level int
	var timeout int
	var meta nucleo.Payload

	sourceNodeID := values["sender"].(string)
	id := values["id"].(string)
	actionName, isAction := values["action"]
	if !isAction {
		panic(errors.New("Can't create an action context, you need a action field!"))
	}
	level = values["level"].(int)

	parentID := ""
	if p, ok := values["parentID"]; ok {
		if s, ok := p.(string); ok {
			parentID = s
		}
	}
	// params := payload.Empty()
	// if values["params"] != nil {
	// 	params = payload.New(values["params"])
	// }
	params := payload.New(values["params"])

	if values["timeout"] != nil {
		timeout = values["timeout"].(int)
	}
	if values["meta"] != nil {
		meta = payload.New(values["meta"])
	} else {
		meta = payload.Empty()
	}

	newContext := Context{
		broker:       broker,
		sourceNodeID: sourceNodeID,
		targetNodeID: sourceNodeID,
		id:           id,
		actionName:   actionName.(string),
		parentID:     parentID,
		params:       params,
		meta:         meta,
		timeout:      timeout,
		level:        level,
	}

	return &newContext
}

// EventContext create an event context for a remote call.
func EventContext(broker *nucleo.BrokerDelegates, values map[string]interface{}) nucleo.BrokerContext {
	var meta nucleo.Payload
	sourceNodeID := values["sender"].(string)
	id := ""
	if t, ok := values["id"]; ok {
		id = t.(string)
	}
	eventName, isEvent := values["event"]
	if !isEvent {
		panic(errors.New("Can't create an event context, you need an event field!"))
	}
	if values["meta"] != nil {
		meta = payload.New(values["meta"])
	} else {
		meta = payload.Empty()
	}
	newContext := Context{
		broker:       broker,
		sourceNodeID: sourceNodeID,
		id:           id,
		eventName:    eventName.(string),
		broadcast:    values["broadcast"].(bool),
		params:       payload.New(values["data"]),
		meta:         meta,
	}
	if values["groups"] != nil {
		temp := values["groups"]
		aTransformer := payload.ArrayTransformer(&temp)
		if aTransformer != nil {
			iArray := aTransformer.InterfaceArray(&temp)
			sGroups := make([]string, len(iArray))
			for index, item := range iArray {
				sGroups[index] = item.(string)
			}
			newContext.groups = sGroups
		}
	}
	return &newContext
}

func (context *Context) IsBroadcast() bool {
	return context.broadcast
}

func (context *Context) RequestID() string {
	return context.requestID
}

// AsMap : export context info in a map[string]
func (context *Context) AsMap() map[string]interface{} {
	mapResult := make(map[string]interface{})

	var tracing bool
	if context.meta.Get("tracing").Exists() {
		tracing = context.meta.Get("tracing").Bool()
	}

	mapResult["id"] = context.id
	mapResult["requestID"] = context.requestID

	mapResult["level"] = context.level
	mapResult["meta"] = context.meta.RawMap()
	mapResult["caller"] = context.caller
	mapResult["tracing"] = tracing
	mapResult["parentID"] = context.parentID

	if context.actionName != "" {
		mapResult["action"] = context.actionName
		mapResult["timeout"] = context.timeout
		mapResult["params"] = context.params.Value()
	}
	if context.eventName != "" {
		mapResult["event"] = context.eventName
		mapResult["groups"] = context.groups
		mapResult["broadcast"] = context.broadcast
		mapResult["data"] = context.params.Value()
		mapResult["level"] = context.level
	}

	//streaming not supported yet
	mapResult["stream"] = false
	//mapResult["seq"] = 0 // for stream payloads

	return mapResult
}

func (context *Context) MCall(callMaps map[string]map[string]interface{}) chan map[string]nucleo.Payload {
	return context.broker.MultActionDelegate(callMaps)
}

// Call : main entry point to call actions.
// chained action invocation
func (context *Context) Call(actionName string, params interface{}, opts ...nucleo.Options) chan nucleo.Payload {
	actionContext := context.ChildActionContext(actionName, payload.New(params), opts...)
	return context.broker.ActionDelegate(actionContext, opts...)
}

// Emit : Emit an event (grouped & balanced global event)
func (context *Context) Emit(eventName string, params interface{}, groups ...string) {
	context.Logger().Debug("Context Emit() eventName: ", eventName)
	newContext := context.ChildEventContext(eventName, payload.New(params), groups, false)
	context.broker.EmitEvent(newContext)
}

// Broadcast : Broadcast an event for all local & remote services
func (context *Context) Broadcast(eventName string, params interface{}, groups ...string) {
	newContext := context.ChildEventContext(eventName, payload.New(params), groups, true)
	context.broker.BroadcastEvent(newContext)
}

func (context *Context) WaitFor(services ...string) error {
	return context.broker.WaitFor(services...)
}

func (context *Context) Publish(services ...interface{}) {
	context.broker.PublishServices(services...)
}

func (context *Context) ActionName() string {
	return context.actionName
}

func (context *Context) EventName() string {
	return context.eventName
}

func (context *Context) Groups() []string {
	return context.groups
}

func (context *Context) Payload() nucleo.Payload {
	return context.params
}

func (context *Context) SetTargetNodeID(targetNodeID string) {
	context.Logger().Debug("context factory SetTargetNodeID() targetNodeID: ", targetNodeID)
	context.targetNodeID = targetNodeID
}

func (context *Context) TargetNodeID() string {
	return context.targetNodeID
}

func (context *Context) SourceNodeID() string {
	return context.sourceNodeID
}

func (context *Context) ID() string {
	return context.id
}

func (context *Context) Meta() nucleo.Payload {
	return context.meta
}

func (context *Context) UpdateMeta(meta nucleo.Payload) {
	context.meta = meta
}

func (context *Context) Logger() *log.Entry {
	if context.actionName != "" {
		return context.broker.Logger("action", context.actionName)
	}
	if context.eventName != "" {
		return context.broker.Logger("event", context.eventName)
	}
	return context.broker.Logger("context", "<root>")
}

func (context *Context) Caller() string {
	return context.caller
}

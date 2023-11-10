package context

import (
	"fmt"

	"github.com/Bendomey/nucleo-go/nucleo"
	"github.com/Bendomey/nucleo-go/nucleo/utils"
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
	params       interface{}
	meta         interface{}
	timeout      int
	level        int
	caller       string
}

func BrokerContext(broker *nucleo.BrokerDelegates) nucleo.BrokerContext {
	// localNodeID := broker.LocalNode().GetID()
	id := fmt.Sprint("rootContext-broker-", "-", utils.RandomString(12))
	// id := fmt.Sprint("rootContext-broker-", localNodeID, "-", utils.RandomString(12))
	context := Context{
		id:       id,
		broker:   broker,
		level:    1,
		parentID: "sudo",
		meta:     map[string]interface{}{},
	}
	return &context
}

func (context *Context) RequestID() string {
	return context.requestID
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

func (context *Context) PublishServices(services ...interface{}) {
	context.broker.PublishServices(services...)
}

func (context *Context) Call(actionName string, params interface{}, opts nucleo.Options) chan interface{} {
	return nil
}

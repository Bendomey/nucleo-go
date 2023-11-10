package broker

import (
	"github.com/Bendomey/nucleo-go/nucleo"
	"github.com/Bendomey/nucleo-go/nucleo/serializer"
	bus "github.com/moleculer-go/goemitter"
	log "github.com/sirupsen/logrus"
)

type ServiceBroker struct {
	namespace string

	logger *log.Entry

	localBus *bus.Emitter

	// registry *registry.ServiceRegistry

	// middlewares *middleware.Dispatch

	serializer *serializer.Serializer

	// services []*service.Service

	started  bool
	starting bool

	context nucleo.BrokerContext

	config nucleo.Config

	delegates *nucleo.BrokerDelegates

	id string

	instanceID string

	node nucleo.Node
}

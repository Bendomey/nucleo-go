package broker

import (
	"time"

	"github.com/Bendomey/nucleo-go/nucleo"
	"github.com/Bendomey/nucleo-go/nucleo/context"
	"github.com/hashicorp/go-uuid"
	log "github.com/sirupsen/logrus"
)

type ServiceBroker struct {
	namespace string
	logger    *log.Entry

	started    bool
	startedAt  *time.Time
	starting   bool
	startingAt *time.Time

	rootContext nucleo.BrokerContext
	config      nucleo.Config

	instanceID string
	id         string

	services map[string]interface{}
}

func New(userConfig nucleo.Config) *ServiceBroker {
	config := mergeConfigs(nucleo.DefaultConfig, userConfig)

	broker := ServiceBroker{
		config: config,
	}

	broker.init()

	return &broker
}

func (broker *ServiceBroker) init() {
	broker.id = broker.config.DiscoverNodeID()
	broker.logger = broker.createBrokerLogger()

	broker.instanceID = broker.generateUUID()

	broker.rootContext = context.BrokerContext(broker.brokerDelegates())
}

func (broker *ServiceBroker) IsStarted() bool {
	return broker.started
}

func (broker *ServiceBroker) newLogger(name string, value string) *log.Entry {
	return broker.logger.WithField(name, value)
}

func (broker *ServiceBroker) brokerDelegates() *nucleo.BrokerDelegates {
	return &nucleo.BrokerDelegates{
		Logger:    broker.newLogger,
		IsStarted: broker.IsStarted,
		Config:    broker.config,
		InstanceID: func() string {
			return broker.instanceID
		},
		BrokerContext: func() nucleo.BrokerContext {
			return broker.rootContext
		},
		PublishServices: broker.PublishServices,
	}
}

func (broker *ServiceBroker) Start() {
	if broker.IsStarted() {
		broker.logger.Warn("broker.Start() called on a broker that already started!")
		return
	}

	broker.starting = true
	startingAtNow := time.Now()
	broker.startingAt = &startingAtNow
	broker.logger.Info("Moleculer is starting...")
	// broker.logger.Info("Node ID: ", broker.localNode.GetID())

	//TODO: start registry here.

	//TODO: start services here.

	broker.started = true
	startedAtNow := time.Now()
	broker.startedAt = &startedAtNow

	broker.startingAt = nil
	broker.starting = false
	broker.logger.Info("Service Broker with ", len(broker.services), " service(s) started successfully.")
}

func (broker *ServiceBroker) Stop() {
	if !broker.started {
		broker.logger.Info("Broker is not started!")
		return
	}

	broker.logger.Info("Service Broker is stopping...")

	//TODO: stop services here.
	//TODO: stop registry here.

	broker.started = false
	broker.startedAt = nil
	broker.logger.Info("Service Broker with ", len(broker.services), " service(s) stopped successfully.")
}

func (broker *ServiceBroker) PublishServices(services ...interface{}) {
}

func (broker *ServiceBroker) generateUUID() string {
	instanceID, err := uuid.GenerateUUID()
	if err != nil {
		broker.logger.Error("Could not create an instance id -  error ", err)
		instanceID = "error creating instance id"
	}

	return instanceID
}

func (broker *ServiceBroker) createBrokerLogger() *log.Entry {
	if broker.config.LogFormat == nucleo.LogFormatJSON {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{})
	}

	if broker.config.LogLevel == nucleo.LogLevelWarn {
		log.SetLevel(log.WarnLevel)
	} else if broker.config.LogLevel == nucleo.LogLevelDebug {
		log.SetLevel(log.DebugLevel)
	} else if broker.config.LogLevel == nucleo.LogLevelTrace {
		log.SetLevel(log.TraceLevel)
	} else if broker.config.LogLevel == nucleo.LogLevelError {
		log.SetLevel(log.ErrorLevel)
	} else if broker.config.LogLevel == nucleo.LogLevelFatal {
		log.SetLevel(log.FatalLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	brokerLogger := log.WithFields(log.Fields{
		"broker":    broker.id,
		"namespace": broker.config.Namespace,
	})
	return brokerLogger
}

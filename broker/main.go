package broker

import (
	"errors"
	"time"

	"github.com/Bendomey/nucleo-go"
	"github.com/Bendomey/nucleo-go/cache"
	"github.com/Bendomey/nucleo-go/context"
	bus "github.com/Bendomey/nucleo-go/emitter"
	"github.com/Bendomey/nucleo-go/metrics"
	"github.com/Bendomey/nucleo-go/middleware"
	"github.com/Bendomey/nucleo-go/payload"
	"github.com/Bendomey/nucleo-go/registry"
	"github.com/Bendomey/nucleo-go/serializer"
	"github.com/Bendomey/nucleo-go/service"
	"github.com/hashicorp/go-uuid"
	log "github.com/sirupsen/logrus"
)

type ServiceBroker struct {
	namespace string

	logger *log.Entry

	localBus *bus.Emitter

	registry *registry.ServiceRegistry

	middlewares *middleware.Dispatch

	cache cache.Cache

	serializer *serializer.Serializer

	services []*service.Service

	started  bool
	starting bool

	rootContext nucleo.BrokerContext

	config nucleo.Config

	delegates *nucleo.BrokerDelegates

	id string

	instanceID string

	localNode nucleo.Node
}

// GetLocalBus : return the service broker local bus (Event Emitter)
func (broker *ServiceBroker) LocalBus() *bus.Emitter {
	return broker.localBus
}

// stopService stop the service.
func (broker *ServiceBroker) stopService(svc *service.Service) {
	broker.middlewares.CallHandlers("serviceStopping", svc)
	svc.Stop(broker.rootContext.ChildActionContext("service.stop", payload.Empty()))
	broker.middlewares.CallHandlers("serviceStopped", svc)
}

// applyServiceConfig apply broker config to the service configuration
// settings is an import config copy from broker to the service.
func (broker *ServiceBroker) applyServiceConfig(svc *service.Service) {
	if bkrConfig, exists := broker.config.Services[svc.Name()]; exists {
		svcConfig, ok := bkrConfig.(map[string]interface{})
		if ok {
			_, ok := svcConfig["settings"]
			if ok {
				settings, ok := svcConfig["settings"].(map[string]interface{})
				if ok {
					svc.AddSettings(settings)
				} else {
					broker.logger.Errorln("Could not add service settings - Error converting the input settings to map[string]interface{} - Invalid format! Service Config : ", svcConfig)
				}
			}

			_, ok = svcConfig["metadata"]
			if ok {
				metadata, ok := svcConfig["metadata"].(map[string]interface{})
				if ok {
					svc.AddSettings(metadata)
				} else {
					broker.logger.Errorln("Could not add service metadata - Error converting the input metadata to map[string]interface{} - Invalid format! Service Config : ", svcConfig)
				}
			}
		} else {
			broker.logger.Errorln("Could not apply service configuration - Error converting the service config to map[string]interface{} - Invalid format! Broker Config : ", bkrConfig)
		}

	}
}

// startService start a service.
func (broker *ServiceBroker) startService(svc *service.Service) {

	broker.logger.Debugln("Broker start service: ", svc.FullName())

	broker.applyServiceConfig(svc)

	broker.middlewares.CallHandlers("serviceStarting", svc)

	broker.waitForDependencies(svc)

	broker.registry.AddLocalService(svc)

	broker.middlewares.CallHandlers("serviceStarted", svc)

	svc.Start(broker.rootContext.ChildActionContext("service.start", payload.Empty()))
}

// waitForDependencies wait for all services listed in the service dependencies to be discovered.
func (broker *ServiceBroker) waitForDependencies(service *service.Service) {
	if len(service.Dependencies()) == 0 {
		return
	}
	start := time.Now()
	for {
		if !broker.started {
			break
		}
		found := true
		for _, dependency := range service.Dependencies() {
			known := broker.registry.KnowService(dependency)
			if !known {
				found = false
				break
			}
		}
		if found {
			broker.logger.Debugln("waitForDependencies() - All dependencies were found :) -> service: ", service.Name(), " wait For Dependencies: ", service.Dependencies())
			break
		}
		if time.Since(start) > broker.config.WaitForDependenciesTimeout {
			broker.logger.Warnln("waitForDependencies() - Time out ! service: ", service.Name(), " wait For Dependencies: ", service.Dependencies())
			break
		}
		time.Sleep(time.Microsecond)
	}
}

func (broker *ServiceBroker) broadcastLocal(eventName string, params ...interface{}) {
	broker.LocalBus().EmitAsync(eventName, params)
}

func (broker *ServiceBroker) createBrokerLogger() *log.Entry {
	if broker.config.LogFormat == nucleo.LogFormatJSON {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
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
		"broker": broker.id,
	})
	//broker.logger.Debugln("Broker Log Setup -> Level", log.GetLevel(), " nodeID: ", nodeID)
	return brokerLogger
}

// addService internal addService .. adds one service.Service instance to broker.services list.
func (broker *ServiceBroker) addService(svc *service.Service) {
	svc.SetNodeID(broker.localNode.GetID())
	broker.services = append(broker.services, svc)
	if broker.started || broker.starting {
		broker.startService(svc)
	}
	broker.logger.Debugln("Broker - addService() - fullname: ", svc.FullName(), " # actions: ", len(svc.Actions()), " # events: ", len(svc.Events()))
}

// resolveSchema getting schema from interface
func (broker *ServiceBroker) resolveSchema(svc interface{}) (nucleo.ServiceSchema, bool) {
	schema, isSchema := svc.(nucleo.ServiceSchema)
	if !isSchema {
		s, ok := svc.(*nucleo.ServiceSchema)
		if ok {
			schema = *s
			isSchema = ok
		}
	}

	return schema, isSchema
}

// createService create a new service instance, from a struct or a schema :)
func (broker *ServiceBroker) createService(svc interface{}) (*service.Service, error) {

	schema, isSchema := broker.resolveSchema(svc)

	if !isSchema {
		svc, err := service.FromObject(svc, broker.delegates)
		if err != nil {
			return nil, err
		}
		return svc, nil
	}
	return service.FromSchema(schema, broker.delegates), nil
}

// WaitFor : wait for all services to be available
func (broker *ServiceBroker) WaitFor(services ...string) error {
	for _, svc := range services {
		if err := broker.waitForService(svc); err != nil {
			return err
		}
	}
	return nil
}

// WaitForNodes : wait for all nodes to be available
func (broker *ServiceBroker) WaitForNodes(nodes ...string) error {
	for _, nodeID := range nodes {
		if err := broker.waitForNode(nodeID); err != nil {
			return err
		}
	}
	return nil
}

func (broker *ServiceBroker) KnowAction(action string) bool {
	return broker.registry.KnowAction(action)
}

// WaitForActions : wait for all actions to be available
func (broker *ServiceBroker) WaitForActions(actions ...string) error {
	for _, action := range actions {
		if err := broker.waitAction(action); err != nil {
			return err
		}
	}
	return nil
}

// waitForService wait for a service to be available
func (broker *ServiceBroker) waitForService(service string) error {
	start := time.Now()
	for {
		if broker.registry.KnowService(service) {
			break
		}
		if time.Since(start) > broker.config.WaitForDependenciesTimeout {
			err := errors.New("waitForService() - Timeout ! service: " + service)
			broker.logger.Errorln(err)
			return err
		}
		time.Sleep(time.Microsecond)
	}
	return nil
}

// waitAction wait for an action to be available
func (broker *ServiceBroker) waitAction(action string) error {
	start := time.Now()
	for {
		if broker.registry.KnowAction(action) {
			break
		}
		if time.Since(start) > broker.config.WaitForDependenciesTimeout {
			err := errors.New("waitAction() - Timeout ! action: " + action)
			broker.logger.Errorln(err)
			return err
		}
		time.Sleep(time.Microsecond)
	}
	return nil
}

// waitForNode wait for a node to be available
func (broker *ServiceBroker) waitForNode(nodeID string) error {
	start := time.Now()
	for {
		if broker.registry.KnowNode(nodeID) {
			break
		}
		if time.Since(start) > broker.config.WaitForDependenciesTimeout {
			err := errors.New("waitForNode() - Timeout ! nodeID: " + nodeID)
			broker.logger.Errorln(err)
			return err
		}
		time.Sleep(time.Microsecond)
	}
	return nil
}

// Publish : for each service schema it will validate and create
// a service instance in the broker.
func (broker *ServiceBroker) PublishServices(services ...interface{}) {
	for _, item := range services {
		svc, err := broker.createService(item)
		if err != nil {
			panic(errors.New("Could not publish service - error: " + err.Error()))
		}
		broker.addService(svc)
	}
}

func (broker *ServiceBroker) Start() {
	if broker.IsStarted() {
		broker.logger.Warnln("broker.Start() called on a broker that already started!")
		return
	}
	broker.starting = true
	broker.logger.Infoln("nucleo is starting...")
	broker.logger.Infoln("Node ID: ", broker.localNode.GetID())

	broker.middlewares.CallHandlers("brokerStarting", broker.delegates)

	broker.registry.Start()

	internalServices := broker.registry.LocalServices()
	for _, service := range internalServices {
		service.SetNodeID(broker.localNode.GetID())
		broker.startService(service)
	}

	for _, service := range broker.services {
		broker.startService(service)
	}

	for _, service := range internalServices {
		broker.addService(service)
	}

	broker.logger.Debugln("Broker -> registry started!")

	defer broker.broadcastLocal("$broker.started")
	defer broker.middlewares.CallHandlers("brokerStarted", broker.delegates)

	broker.started = true
	broker.starting = false
	broker.logger.Infoln("Service Broker with ", len(broker.services), " service(s) started successfully.")
}

func (broker *ServiceBroker) Stop() {
	if !broker.started {
		broker.logger.Infoln("Broker is not started!")
		return
	}
	broker.logger.Infoln("Service Broker is stopping...")

	broker.middlewares.CallHandlers("brokerStopping", broker.delegates)

	for _, service := range broker.services {
		broker.stopService(service)
	}

	broker.registry.Stop()

	broker.started = false
	broker.broadcastLocal("$broker.stopped")

	broker.middlewares.CallHandlers("brokerStopped", broker.delegates)
}

type callPair struct {
	label  string
	result nucleo.Payload
}

func (broker *ServiceBroker) invokeMCalls(callMaps map[string]map[string]interface{}, result chan map[string]nucleo.Payload) {
	if len(callMaps) == 0 {
		result <- make(map[string]nucleo.Payload)
		return
	}

	resultChan := make(chan callPair)
	for label, content := range callMaps {
		go func(label, actionName string, params interface{}, results chan callPair) {
			result := <-broker.Call(actionName, params)
			results <- callPair{label, result}
		}(label, content["action"].(string), content["params"], resultChan)
	}

	timeoutChan := make(chan bool, 1)
	go func(timeout time.Duration) {
		time.Sleep(timeout)
		timeoutChan <- true
	}(broker.config.MCallTimeout)

	results := make(map[string]nucleo.Payload)
	for {
		select {
		case pair := <-resultChan:
			results[pair.label] = pair.result
			if len(results) == len(callMaps) {
				result <- results
				return
			}
		case <-timeoutChan:
			timeoutError := errors.New("MCall timeout error.")
			broker.logger.Errorln(timeoutError)
			for label, _ := range callMaps {
				if _, exists := results[label]; !exists {
					results[label] = payload.New(timeoutError)
				}
			}
			result <- results
			return
		}
	}
}

// MCall perform multiple calls and return all results together in a nice map indexed by name.
func (broker *ServiceBroker) MCall(callMaps map[string]map[string]interface{}) chan map[string]nucleo.Payload {
	result := make(chan map[string]nucleo.Payload, 1)
	go broker.invokeMCalls(callMaps, result)
	return result
}

// Call :  invoke a service action and return a channel which will eventualy deliver the results ;)
func (broker *ServiceBroker) Call(actionName string, params interface{}, opts ...nucleo.Options) chan nucleo.Payload {
	broker.logger.Traceln("Broker - Call() actionName: ", actionName, " params: ", params, " opts: ", opts)
	if !broker.IsStarted() {
		panic(errors.New("Broker must be started before making calls :("))
	}
	actionContext := broker.rootContext.ChildActionContext(actionName, payload.New(params), opts...)
	return broker.registry.LoadBalanceCall(actionContext, opts...)
}

func (broker *ServiceBroker) Emit(event string, params interface{}, groups ...string) {
	broker.logger.Traceln("Broker - Emit() event: ", event, " params: ", params, " groups: ", groups)
	if !broker.IsStarted() {
		panic(errors.New("Broker must be started before emiting events :("))
	}
	newContext := broker.rootContext.ChildEventContext(event, payload.New(params), groups, false)
	broker.registry.LoadBalanceEvent(newContext)
}

func (broker *ServiceBroker) Broadcast(event string, params interface{}, groups ...string) {
	broker.logger.Traceln("Broker - Broadcast() event: ", event, " params: ", params, " groups: ", groups)
	if !broker.IsStarted() {
		panic(errors.New("Broker must be started before broadcasting events :("))
	}
	newContext := broker.rootContext.ChildEventContext(event, payload.New(params), groups, true)
	broker.registry.BroadcastEvent(newContext)
}

func (broker *ServiceBroker) IsStarted() bool {
	return broker.started
}

func (broker *ServiceBroker) GetLogger(name string, value string) *log.Entry {
	return broker.logger.WithField(name, value)
}

func (broker *ServiceBroker) LocalNode() nucleo.Node {
	return broker.localNode
}

func (broker *ServiceBroker) newLogger(name string, value string) *log.Entry {
	return broker.logger.WithField(name, value)
}

func (broker *ServiceBroker) setupLocalBus() {
	broker.localBus = bus.Construct()

	broker.localBus.On("$registry.service.added", func(args ...interface{}) {
		//TODO check code from -> this.broker.servicesChanged(true)
	})
}

func (broker *ServiceBroker) registerMiddlewares() {
	broker.middlewares = middleware.Dispatcher(broker.logger.WithField("middleware", "dispatcher"))
	for _, mware := range broker.config.Middlewares {
		broker.middlewares.Add(mware)
	}
	if !broker.config.DisableInternalMiddlewares {
		broker.registerInternalMiddlewares()
	}
}

func (broker *ServiceBroker) registerInternalMiddlewares() {
	broker.middlewares.Add(metrics.Middlewares())
}

func (broker *ServiceBroker) init() {
	broker.id = broker.config.DiscoverNodeID()
	broker.logger = broker.createBrokerLogger()
	broker.setupLocalBus()

	broker.registerMiddlewares()

	broker.config = broker.middlewares.CallHandlers("Config", broker.config).(nucleo.Config)

	instanceID, err := uuid.GenerateUUID()
	if err != nil {
		broker.logger.Errorln("Could not create an instance id -  error ", err)
		instanceID = "error creating instance id"
	}
	broker.instanceID = instanceID

	broker.delegates = broker.createDelegates()
	broker.registry = registry.CreateRegistry(broker.id, broker.delegates)
	broker.localNode = broker.registry.LocalNode()
	broker.rootContext = context.BrokerContext(broker.delegates)

}

func (broker *ServiceBroker) createDelegates() *nucleo.BrokerDelegates {
	return &nucleo.BrokerDelegates{
		LocalNode: broker.LocalNode,
		Logger:    broker.newLogger,
		Bus:       broker.LocalBus,
		IsStarted: broker.IsStarted,
		Config:    broker.config,
		InstanceID: func() string {
			return broker.instanceID
		},
		ActionDelegate: func(context nucleo.BrokerContext, opts ...nucleo.Options) chan nucleo.Payload {
			return broker.registry.LoadBalanceCall(context, opts...)
		},
		EmitEvent: func(context nucleo.BrokerContext) {
			broker.registry.LoadBalanceEvent(context)
		},
		BroadcastEvent: func(context nucleo.BrokerContext) {
			broker.registry.BroadcastEvent(context)
		},
		HandleRemoteEvent: func(context nucleo.BrokerContext) {
			broker.registry.HandleRemoteEvent(context)
		},
		ServiceForAction: func(name string) []*nucleo.ServiceSchema {
			svcs := broker.registry.ServiceForAction(name)
			if svcs != nil {
				result := make([]*nucleo.ServiceSchema, len(svcs))
				for i, svc := range svcs {
					result[i] = svc.Schema()
				}
				return result
			}
			return nil
		},
		MultActionDelegate: func(callMaps map[string]map[string]interface{}) chan map[string]nucleo.Payload {
			return broker.MCall(callMaps)
		},
		BrokerContext: func() nucleo.BrokerContext {
			return broker.rootContext
		},
		MiddlewareHandler: broker.middlewares.CallHandlers,
		PublishServices:   broker.PublishServices,
		WaitFor:           broker.WaitFor,
	}
}

// New : returns a valid broker based on environment configuration
// this is usually called when creating a broker to starting the service(s)
func New(userConfig ...*nucleo.Config) *ServiceBroker {
	config := mergeConfigs(nucleo.DefaultConfig, userConfig)
	broker := ServiceBroker{config: config}
	broker.init()
	return &broker
}

package transit

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Bendomey/nucleo-go/nucleo"
	"github.com/Bendomey/nucleo-go/nucleo/payload"
	"github.com/Bendomey/nucleo-go/nucleo/serializer"
	"github.com/Bendomey/nucleo-go/nucleo/transport"
	"github.com/Bendomey/nucleo-go/nucleo/transport/nats"
	memory "github.com/Bendomey/nucleo-go/nucleo/transport/tcp"
	log "github.com/sirupsen/logrus"
)

type pendingRequest struct {
	context    nucleo.BrokerContext
	resultChan *chan nucleo.Payload
	timer      *time.Timer
}

// Transit is a transport implementation.
type Transit struct {
	logger               *log.Entry
	transport            transport.Transport
	broker               *nucleo.BrokerDelegates
	isConnected          bool
	pendingRequests      map[string]pendingRequest
	pendingRequestsMutex *sync.Mutex
	serializer           serializer.Serializer

	knownNeighbours   map[string]int64
	neighboursTimeout time.Duration
	neighboursMutex   *sync.Mutex
	// brokerStarted     bool
}

const DATATYPE_UNDEFINED = 0
const DATATYPE_NULL = 1
const DATATYPE_JSON = 2
const DATATYPE_BUFFER = 3

func Create(broker *nucleo.BrokerDelegates) transport.Transit {
	pendingRequests := make(map[string]pendingRequest)
	knownNeighbours := make(map[string]int64)
	transitImpl := Transit{
		broker:            broker,
		isConnected:       false,
		pendingRequests:   pendingRequests,
		logger:            broker.Logger("Transit", ""),
		serializer:        serializer.New(broker),
		neighboursTimeout: broker.Config.NeighboursCheckTimeout,
		knownNeighbours:   knownNeighbours,
	}
	// TODO: work on events later
	// broker.Bus().On("$node.disconnected", transitImpl.onNodeDisconnected)
	// broker.Bus().On("$node.connected", transitImpl.onNodeConnected)
	// broker.Bus().On("$broker.started", transitImpl.onBrokerStarted)
	// broker.Bus().On("$registry.service.added", transitImpl.onServiceAdded)

	return &transitImpl
}

// func (transit *Transit) onBrokerStarted(values ...interface{}) {
// 	if transit.isConnected {
// 		transit.broadcastNodeInfo("")
// 		transit.brokerStarted = true
// 	}
// }

// // broadcastNodeInfo send the local node info to the target node, if empty to all nodes.
// func (transit *Transit) broadcastNodeInfo(targetNodeID string) {
// 	payload := transit.broker.LocalNode().ExportAsMap()
// 	payload["sender"] = payload["id"]
// 	payload["neighbours"] = transit.neighbours()
// 	payload["version"] = 1
// 	payload["config"] = configToMap(transit.broker.Config)
// 	payload["instanceID"] = transit.broker.InstanceID()

// 	message, _ := transit.serializer.MapToPayload(&payload)
// 	transit.transport.Publish("INFO", targetNodeID, message)
// }

// func configToMap(config nucleo.Config) map[string]string {
// 	m := make(map[string]string)
// 	m["logLevel"] = config.LogLevel
// 	m["transporter"] = config.Transporter
// 	m["namespace"] = config.Namespace
// 	m["requestTimeout"] = config.RequestTimeout.String()
// 	return m
// }

func (transit *Transit) SendHeartbeat() {
	node := transit.broker.LocalNode().ExportAsMap()
	payload := map[string]interface{}{
		"sender":  node["id"],
		"cpu":     node["cpu"],
		"cpuSeq":  node["cpuSeq"],
		"version": 1,
	}
	message, err := transit.serializer.MapToPayload(&payload)
	if err == nil {
		transit.transport.Publish("HEARTBEAT", "", message)
	}
}

func (transit *Transit) requestTimedOut(resultChan *chan nucleo.Payload, context nucleo.BrokerContext) func() {
	pError := payload.New(errors.New("request timeout"))
	return func() {
		transit.logger.Debug("requestTimedOut() nodeID: ", context.TargetNodeID())
		transit.pendingRequestsMutex.Lock()
		defer transit.pendingRequestsMutex.Unlock()

		p, exists := transit.pendingRequests[context.ID()]
		if exists {
			(*p.resultChan) <- pError
			p.timer.Stop()
			delete(transit.pendingRequests, p.context.ID())
		}
	}
}

func (transit *Transit) Request(context nucleo.BrokerContext) chan nucleo.Payload {
	resultChan := make(chan nucleo.Payload)

	targetNodeID := context.TargetNodeID()
	payload := context.AsMap()
	payload["sender"] = transit.broker.LocalNode().GetID()
	payload["version"] = 1
	if context.Payload().Exists() {
		payload["paramsType"] = DATATYPE_JSON
	} else {
		payload["paramsType"] = DATATYPE_NULL
	}

	transit.logger.Trace("Request() targetNodeID: ", targetNodeID, " payload: ", payload)

	message, err := transit.serializer.MapToPayload(&payload)
	if err != nil {
		transit.logger.Error("Request() Error serializing the payload: ", payload, " error: ", err)
		panic(fmt.Errorf("Error trying to serialize the payload. Likely issues with the action params. Error: %s", err))
	}

	transit.pendingRequestsMutex.Lock()
	transit.logger.Debug("Request() pending request id: ", context.ID(), " targetNodeId: ", context.TargetNodeID())
	transit.pendingRequests[context.ID()] = pendingRequest{
		context,
		&resultChan,

		time.AfterFunc(
			transit.broker.Config.RequestTimeout,
			transit.requestTimedOut(&resultChan, context)),
	}
	transit.pendingRequestsMutex.Unlock()

	transit.transport.Publish("REQ", targetNodeID, message)
	return resultChan
}

func isNats(v string) bool {
	return strings.Index(v, "nats://") > -1
}

func resolveNamespace(namespace string) string {
	if namespace != "" {
		return "NUCLEO-" + namespace
	}
	return "NUCLEO"
}

// waitForNeighbours this function will wait for neighbour nodes or timeout if the expected number is not received after a time out.
func (transit *Transit) waitForNeighbours() bool {
	if transit.broker.Config.DontWaitForNeighbours {
		return true
	}
	start := time.Now()
	for {
		expected := transit.expectedNeighbours()
		neighbours := transit.neighbours()
		if expected <= neighbours && (expected > 0 || neighbours > 0) {
			transit.logger.Debug("waitForNeighbours() - received info from all expected neighbours :) -> expected: ", expected)
			return true
		}
		if time.Since(start) > transit.neighboursTimeout {
			transit.logger.Warn("waitForNeighbours() - Time out ! did not receive info from all expected neighbours: ", expected, "  INFOs received: ", neighbours)
			return false
		}
		if !transit.isConnected {
			return false
		}
		time.Sleep(transit.broker.Config.WaitForNeighboursInterval)
	}
}

// expectedNeighbours calculate the expected number of neighbours
func (transit *Transit) expectedNeighbours() int64 {
	neighbours := transit.neighbours()
	if neighbours == 0 {
		return 0
	}

	var total int64
	transit.neighboursMutex.Lock()
	for _, value := range transit.knownNeighbours {
		total = total + value
	}
	transit.neighboursMutex.Unlock()
	return total / neighbours
}

// neighbours return the total number of known neighbours.
func (transit *Transit) neighbours() int64 {
	return int64(len(transit.knownNeighbours))
}

//DiscoverNodes will check if there are neighbours and return true if any are found ;).
func (transit *Transit) DiscoverNodes() chan bool {
	result := make(chan bool)
	go func() {
		transit.DiscoverNode("")
		result <- transit.waitForNeighbours()
	}()
	return result
}

func (transit *Transit) DiscoverNode(nodeID string) {
	payload := map[string]interface{}{
		"sender":  transit.broker.LocalNode().GetID(),
		"version": 1,
	}
	message, err := transit.serializer.MapToPayload(&payload)
	if err == nil {
		transit.transport.Publish("DISCOVER", nodeID, message)
	}
}

// CreateTransport : based on config it will load the transporter
func (transit *Transit) createTransport() transport.Transport {
	var transport transport.Transport
	if isNats(transit.broker.Config.Transporter) {
		transit.logger.Info("Transporter: NatsTransporter")
		transport = transit.createNatsTransporter()
	} else {
		transit.logger.Info("Transporter: Memory")
		transport = transit.createMemoryTransporter()
	}
	transport.SetPrefix(resolveNamespace(transit.broker.Config.Namespace))
	transport.SetNodeID(transit.broker.LocalNode().GetID())
	transport.SetSerializer(transit.serializer)
	return transport
}

func (transit *Transit) createNatsTransporter() transport.Transport {
	transit.logger.Debug("createNatsTransporter()")

	return nats.CreateNatsTransporter(nats.NatsOptions{
		URL:            transit.broker.Config.Transporter,
		Name:           transit.broker.LocalNode().GetID(),
		Logger:         transit.logger.WithField("transport", "nats"),
		Serializer:     transit.serializer,
		AllowReconnect: true,
		ReconnectWait:  time.Second * 2,
		MaxReconnect:   -1,
	})
}

func (transit *Transit) createMemoryTransporter() transport.Transport {
	transit.logger.Debug("createMemoryTransporter() ... ")
	logger := transit.logger.WithField("transport", "memory")
	mem := memory.Create(logger, &memory.SharedMemory{})
	return &mem
}

// Connect : connect the transit with the transporter, subscribe to all events and start publishing its node info
func (transit *Transit) Connect() chan error {
	endChan := make(chan error)
	if transit.isConnected {
		endChan <- nil
		return endChan
	}
	transit.logger.Debug("Transit - Connecting transport...")
	transit.transport = transit.createTransport()
	go func() {
		err := <-transit.transport.Connect()
		if err == nil {
			transit.isConnected = true
			transit.logger.Debug("Transit - Transport Connected!")

			transit.subscribe()

		} else {
			transit.logger.Debug("Transit - Error connecting transport - error: ", err)
		}
		endChan <- err
	}()
	return endChan
}

func (transit *Transit) subscribe() {
	// nodeID := transit.broker.LocalNode().GetID()
	// transit.transport.Subscribe("RES", nodeID, transit.validate(transit.reponseHandler()))

	// transit.transport.Subscribe("REQ", nodeID, transit.validate(transit.requestHandler()))
	// //transit.transport.Subscribe("REQB", nodeID, transit.requestHandler())
	// transit.transport.Subscribe("EVENT", nodeID, transit.validate(transit.eventHandler()))

	// transit.transport.Subscribe("HEARTBEAT", "", transit.validate(transit.emitRegistryEvent("HEARTBEAT")))
	// transit.transport.Subscribe("DISCONNECT", "", transit.validate(transit.emitRegistryEvent("DISCONNECT")))
	// transit.transport.Subscribe("INFO", "", transit.validate(transit.emitRegistryEvent("INFO")))
	// transit.transport.Subscribe("INFO", nodeID, transit.validate(transit.emitRegistryEvent("INFO")))
	// transit.transport.Subscribe("DISCOVER", nodeID, transit.validate(transit.discoverHandler()))
	// transit.transport.Subscribe("DISCOVER", "", transit.validate(transit.discoverHandler()))
	// transit.transport.Subscribe("PING", nodeID, transit.validate(transit.pingHandler()))
	// transit.transport.Subscribe("PONG", nodeID, transit.validate(transit.pongHandler()))

}

// sendDisconnect broadcast a DISCONNECT pkt to all nodes informing this one is stopping.
func (transit *Transit) sendDisconnect() {
	payload := make(map[string]interface{})
	payload["sender"] = transit.broker.LocalNode().GetID()
	payload["version"] = 1
	msg, _ := transit.serializer.MapToPayload(&payload)
	transit.transport.Publish("DISCONNECT", "", msg)
}

// Disconnect : disconnect the transit's  transporter.
func (transit *Transit) Disconnect() chan error {
	endChan := make(chan error)
	if !transit.isConnected {
		endChan <- nil
		return endChan
	}
	transit.logger.Info("Transit - Disconnecting transport...")
	transit.sendDisconnect()
	transit.isConnected = false
	return transit.transport.Disconnect()
}

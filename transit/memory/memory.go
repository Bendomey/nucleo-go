package memory

import (
	"fmt"
	"sync"

	"github.com/Bendomey/nucleo-go"
	"github.com/Bendomey/nucleo-go/serializer"
	"github.com/Bendomey/nucleo-go/transit"
	"github.com/Bendomey/nucleo-go/utils"
	log "github.com/sirupsen/logrus"
)

type Subscription struct {
	id            string
	transporterId string
	handler       transit.TransportHandler
	active        bool
}

type SharedMemory struct {
	handlers map[string][]Subscription
	mutex    *sync.Mutex
}

type MemoryTransporter struct {
	prefix     string
	instanceID string
	logger     *log.Entry
	memory     *SharedMemory
}

func Create(logger *log.Entry, memory *SharedMemory) MemoryTransporter {
	instanceID := utils.RandomString(5)
	if memory.handlers == nil {
		memory.handlers = make(map[string][]Subscription)
	}
	if memory.mutex == nil {
		memory.mutex = &sync.Mutex{}
	}
	return MemoryTransporter{memory: memory, logger: logger, instanceID: instanceID}
}

func (transporter *MemoryTransporter) SetPrefix(prefix string) {
	transporter.prefix = prefix
}

func (transporter *MemoryTransporter) SetNodeID(nodeID string) {
}

func (transporter *MemoryTransporter) SetSerializer(serializer serializer.Serializer) {

}

func (transporter *MemoryTransporter) Connect() chan error {
	transporter.logger.Debugln("[Mem-Trans-", transporter.instanceID, "] -> Connecting() ...")
	endChan := make(chan error)
	go func() {
		endChan <- nil
	}()
	transporter.logger.Infoln("[Mem-Trans-", transporter.instanceID, "] -> Connected() !")
	return endChan
}

func (transporter *MemoryTransporter) Disconnect() chan error {
	endChan := make(chan error)
	transporter.logger.Debugln("[Mem-Trans-", transporter.instanceID, "] -> Disconnecting() ...")

	newHandlers := map[string][]Subscription{}
	for key, subscriptions := range transporter.memory.handlers {
		keep := []Subscription{}
		for _, subscription := range subscriptions {
			if subscription.transporterId != transporter.instanceID {
				keep = append(keep, subscription)
			}
		}
		newHandlers[key] = keep
	}
	transporter.memory.handlers = newHandlers

	go func() {
		endChan <- nil
	}()
	transporter.logger.Infoln("[Mem-Trans-", transporter.instanceID, "] -> Disconnected() !")
	return endChan
}

func topicName(transporter *MemoryTransporter, command string, nodeID string) string {
	if nodeID != "" {
		return fmt.Sprint(transporter.prefix, ".", command, ".", nodeID)
	}
	return fmt.Sprint(transporter.prefix, ".", command)
}

func (transporter *MemoryTransporter) Subscribe(command string, nodeID string, handler transit.TransportHandler) {
	topic := topicName(transporter, command, nodeID)
	transporter.logger.Traceln("[Mem-Trans-", transporter.instanceID, "] Subscribe() listen for command: ", command, " nodeID: ", nodeID, " topic: ", topic)

	subscription := Subscription{utils.RandomString(5) + "_" + command, transporter.instanceID, handler, true}

	transporter.memory.mutex.Lock()
	_, exists := transporter.memory.handlers[topic]
	if exists {
		transporter.memory.handlers[topic] = append(transporter.memory.handlers[topic], subscription)
	} else {
		transporter.memory.handlers[topic] = []Subscription{subscription}
	}
	transporter.memory.mutex.Unlock()
}

func (transporter *MemoryTransporter) Publish(command, nodeID string, message nucleo.Payload) {
	topic := topicName(transporter, command, nodeID)
	transporter.logger.Traceln("[Mem-Trans-", transporter.instanceID, "] Publish() command: ", command, " nodeID: ", nodeID, " message: \n", message, "\n - end")

	transporter.memory.mutex.Lock()
	subscriptions, exists := transporter.memory.handlers[topic]
	transporter.memory.mutex.Unlock()
	if exists {
		for _, subscription := range subscriptions {
			if subscription.active {
				go subscription.handler(message)
			}
		}
	}
}

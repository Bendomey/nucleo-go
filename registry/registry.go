package registry

import (
	"time"

	"github.com/Bendomey/nucleo-go/nucleo"
	log "github.com/sirupsen/logrus"
)

type ServiceRegistry struct {
	logger                *log.Entry
	localNode             nucleo.Node
	broker                *nucleo.BrokerDelegates
	stopping              bool
	heartbeatFrequency    time.Duration
	heartbeatTimeout      time.Duration
	offlineCheckFrequency time.Duration
	offlineTimeout        time.Duration
	namespace             string
}

func CreateRegistry(nodeID string, broker *nucleo.BrokerDelegates) *ServiceRegistry {
	config := broker.Config
	logger := broker.Logger("registry", nodeID)
	localNode := CreateNode(nodeID, true, logger.WithField("Node", nodeID))
	localNode.Unavailable()
	registry := &ServiceRegistry{
		broker:                broker,
		logger:                logger,
		localNode:             localNode,
		heartbeatFrequency:    config.HeartbeatFrequency,
		heartbeatTimeout:      config.HeartbeatTimeout,
		offlineCheckFrequency: config.OfflineCheckFrequency,
		offlineTimeout:        config.OfflineTimeout,
		stopping:              false,
		namespace:             config.Namespace,
	}

	registry.logger.Debug("Service Registry created for broker: ", nodeID)

	return registry
}

func (registry *ServiceRegistry) LocalNode() nucleo.Node {
	return registry.localNode
}

func (registry *ServiceRegistry) Stop() {
	registry.logger.Debug("Registry Stopping...")
	registry.stopping = true
	registry.localNode.Unavailable()
	registry.logger.Debug("Registry Stopped...")
}

// Start : start the registry background processes.
func (registry *ServiceRegistry) Start() {
	registry.logger.Debug("Registry Starting... ")
	registry.stopping = false

	registry.logger.Debug("Registry Started ...")
}

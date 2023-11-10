package registry

import (
	"sync"
	"time"

	"github.com/Bendomey/nucleo-go/nucleo"
	log "github.com/sirupsen/logrus"
)

// NodeCatalog catalog of nodes
type NodeCatalog struct {
	nodes  sync.Map
	logger *log.Entry
}

// CreateNodesCatalog create a node catalog
func CreateNodesCatalog(logger *log.Entry) *NodeCatalog {
	return &NodeCatalog{sync.Map{}, logger}
}

// HeartBeat delegate the heart beat to the node in question payload.sender
func (catalog *NodeCatalog) HeartBeat(heartbeat map[string]interface{}) bool {
	sender := heartbeat["sender"].(string)
	node, nodeExists := catalog.nodes.Load(sender)

	if nodeExists && (node.(nucleo.Node)).IsAvailable() {
		(node.(nucleo.Node)).HeartBeat(heartbeat)
		return true
	}
	return false
}

func (catalog *NodeCatalog) list() []nucleo.Node {
	var result []nucleo.Node
	catalog.nodes.Range(func(key, value interface{}) bool {
		node := value.(nucleo.Node)
		result = append(result, node)
		return true
	})
	return result
}

// expiredNodes check nodes with  heartbeat expired based on the timeout parameter
func (catalog *NodeCatalog) expiredNodes(timeout time.Duration) []nucleo.Node {
	var result []nucleo.Node
	catalog.nodes.Range(func(key, value interface{}) bool {
		node := value.(nucleo.Node)
		if node.IsExpired(timeout) {
			result = append(result, node)
		}
		return true
	})
	return result
}

// findNode : return a Node instance from the catalog
func (catalog *NodeCatalog) findNode(nodeID string) (nucleo.Node, bool) {
	node, exists := catalog.nodes.Load(nodeID)
	if exists {
		return node.(nucleo.Node), true
	} else {
		return nil, false
	}
}

// removeNode : remove a node from the catalog
func (catalog *NodeCatalog) removeNode(nodeID string) {
	catalog.nodes.Delete(nodeID)
}

func (catalog *NodeCatalog) Add(node nucleo.Node) {
	catalog.nodes.Store(node.GetID(), node)
}

func (catalog *NodeCatalog) Info(info map[string]interface{}) (bool, bool) {
	sender := info["sender"].(string)
	node, exists := catalog.findNode(sender)
	var reconnected bool
	if exists {
		reconnected = node.Update(sender, info)
	} else {
		node := CreateNode(sender, false, catalog.logger.WithField("remote-node", sender))
		node.Update(sender, info)
		catalog.Add(node)
	}
	return exists, reconnected
}

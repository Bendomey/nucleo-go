package transit

import (
	"github.com/Bendomey/nucleo-go"
	"github.com/Bendomey/nucleo-go/serializer"
)

type TransportHandler func(nucleo.Payload)

type ValidateMsgFunc func(nucleo.Payload) bool

type Transit interface {
	Emit(nucleo.BrokerContext)
	Request(nucleo.BrokerContext) chan nucleo.Payload
	Connect() chan error
	Disconnect() chan error
	DiscoverNode(nodeID string)

	//DiscoverNodes checks if there are neighbours and return true if any are found ;).
	DiscoverNodes() chan bool
	SendHeartbeat()
}

type Transport interface {
	Connect() chan error
	Disconnect() chan error
	Subscribe(command, nodeID string, handler TransportHandler)
	Publish(command, nodeID string, message nucleo.Payload)

	SetPrefix(prefix string)
	SetNodeID(nodeID string)
	SetSerializer(serializer serializer.Serializer)
}

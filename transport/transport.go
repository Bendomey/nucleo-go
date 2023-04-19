package transport

import "github.com/Bendomey/nucleo-go/nucleo"

type TransportHandler func(nucleo.Payload)

type Transport interface {
	Connect() chan error
	Disconnect() chan error
	Subscribe(command, nodeID string, handler TransportHandler)
	Publish(command, nodeID string, message nucleo.Payload)

	SetPrefix(prefix string)
}

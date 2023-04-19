package serializer

import (
	"io"

	"github.com/Bendomey/nucleo-go/nucleo"
	"github.com/Bendomey/nucleo-go/nucleo/serializer/jsonSerializer"
)

type Serializer interface {
	ReaderToPayload(io.Reader) nucleo.Payload
	BytesToPayload(*[]byte) nucleo.Payload
	PayloadToBytes(nucleo.Payload) []byte
	PayloadToString(nucleo.Payload) string
	MapToString(interface{}) string
	StringToMap(string) map[string]interface{}
	PayloadToContextMap(nucleo.Payload) map[string]interface{}
	MapToPayload(*map[string]interface{}) (nucleo.Payload, error)
}

func New(broker *nucleo.BrokerDelegates) Serializer {
	return jsonSerializer.CreateJSONSerializer(broker.Logger("serializer", "json"))
}

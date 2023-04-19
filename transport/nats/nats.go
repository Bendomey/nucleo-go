package nats

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"

	"github.com/Bendomey/nucleo-go/nucleo"
	"github.com/Bendomey/nucleo-go/nucleo/serializer"
	"github.com/Bendomey/nucleo-go/nucleo/transport"
)

type NatsTransporter struct {
	prefix        string
	opts          *nats.Options
	conn          *nats.Conn
	logger        *log.Entry
	subscriptions []*nats.Subscription
	serializer    serializer.Serializer
}

type NatsOptions struct {
	URL        string
	Name       string
	Logger     *log.Entry
	Serializer serializer.Serializer

	AllowReconnect bool
	ReconnectWait  time.Duration
	MaxReconnect   int
}

// natsOptions - Resolves the nats options.
func natsOptions(options NatsOptions) *nats.Options {
	opts := nats.GetDefaultOptions()
	opts.Name = options.Name
	opts.Url = options.URL
	opts.AllowReconnect = options.AllowReconnect
	if options.ReconnectWait != 0 {
		opts.ReconnectWait = options.ReconnectWait
	}

	if options.MaxReconnect != 0 {
		opts.MaxReconnect = options.MaxReconnect
	}

	return &opts
}

// CreateNatsTransporter - Creates a nats transporter!
func CreateNatsTransporter(options NatsOptions) transport.Transport {
	return &NatsTransporter{
		opts:          natsOptions(options),
		logger:        options.Logger,
		subscriptions: []*nats.Subscription{},
		serializer:    options.Serializer,
	}
}

// Connect - Connects to NATS asynchronously!
func (t *NatsTransporter) Connect() chan error {
	endChan := make(chan error)

	go func() {
		t.logger.Debugf("Nats Connect() - URL: %s, Name: %s", t.opts.Url, t.opts.Name)
		conn, err := t.opts.Connect()

		if err != nil {
			t.logger.Errorf("Nats Connect() - error: %s, URL: %s, Name: %s", err, t.opts.Url, t.opts.Name)
			endChan <- errors.New(fmt.Sprint("Error connecting to NATS. Error: ", err, " URL: ", t.opts.Url))
		}
		t.conn = conn
		t.logger.Infof("Connected to ", t.opts.Url)
		endChan <- nil
	}()

	return endChan
}

// Disconnect - Disconnects from NATS asynchronously!
func (t *NatsTransporter) Disconnect() chan error {
	endChan := make(chan error)

	go func() {
		if t.conn != nil {
			t.logger.Info("Nats not connected")
			endChan <- nil
			return
		}

		for _, sub := range t.subscriptions {
			if err := sub.Unsubscribe(); err != nil {
				t.logger.Errorf("Error occured while unsubscibing. Error: ", err)
			}
		}

		t.conn.Close()
		t.conn = nil
		endChan <- nil
		t.logger.Infof("Disconnected successfully from: ", t.opts.Url)
	}()

	return endChan
}

// SetPrefix - Sets the prefix
func (t *NatsTransporter) SetPrefix(prefix string) {
	t.prefix = prefix
}

// Topic - Resolve topic name by append command to the node's id.
func (t *NatsTransporter) topicName(command string, nodeID string) string {
	parts := []string{t.prefix, command}
	if strings.TrimSpace(nodeID) != "" {
		parts = append(parts, nodeID)
	}

	return strings.Join(parts, ".")
}

func (t *NatsTransporter) Subscribe(command, nodeID string, handler transport.TransportHandler) {
	if t.conn == nil {
		msg := fmt.Sprint("NATS.Subscribe() No connection :( -> command: ", command, " nodeID: ", nodeID)
		t.logger.Warn(msg)
		panic(errors.New(msg))
	}

	topic := t.topicName(command, nodeID)

	sub, err := t.conn.Subscribe(topic, func(msg *nats.Msg) {
		payload := t.serializer.BytesToPayload(&msg.Data)
		t.logger.Debug(fmt.Sprintf("Incoming %s packet from '%s'", topic, payload.Get("sender").String()))
		handler(payload)
	})
	if err != nil {
		t.logger.Error("Cannot subscribe: ", topic, " error: ", err)
		return
	}
	t.subscriptions = append(t.subscriptions, sub)
}

func (t *NatsTransporter) Publish(command, nodeID string, message nucleo.Payload) {
	if t.conn == nil {
		msg := fmt.Sprint("NATS.Publish() No connection :( -> command: ", command, " nodeID: ", nodeID)
		t.logger.Warn(msg)
		panic(errors.New(msg))
	}

	topic := t.topicName(command, nodeID)
	t.logger.Debug("nats.Publish() command: ", command, " topic: ", topic, " nodeID: ", nodeID)
	t.logger.Trace("message: \n", message, "\n - end")
	err := t.conn.Publish(topic, t.serializer.PayloadToBytes(message))
	if err != nil {
		t.logger.Error("Error on publish: error: ", err, " command: ", command, " topic: ", topic)
		panic(err)
	}
}

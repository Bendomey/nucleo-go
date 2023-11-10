package nucleo

import (
	"fmt"
	"os"
	"time"

	"github.com/Bendomey/nucleo-go/nucleo/utils"
	log "github.com/sirupsen/logrus"
)

type LogLevelType string

const (
	LogLevelInfo  LogLevelType = "INFO"
	LogLevelWarn  LogLevelType = "WARN"
	LogLevelError LogLevelType = "ERROR"
	LogLevelFatal LogLevelType = "FATAL"
	LogLevelTrace LogLevelType = "TRACE"
	LogLevelDebug LogLevelType = "DEBUG"
)

type LogFormatType string

const (
	LogFormatJSON LogFormatType = "JSON"
	LogFormatText LogFormatType = "TEXT"
)

type Config struct {
	LogLevel                   LogLevelType
	LogFormat                  LogFormatType
	DiscoverNodeID             func() string
	HeartbeatFrequency         time.Duration
	HeartbeatTimeout           time.Duration
	OfflineCheckFrequency      time.Duration
	OfflineTimeout             time.Duration
	WaitForDependenciesTimeout time.Duration
	Namespace                  string
	RequestTimeout             time.Duration
	Created                    func()
	Started                    func()
	Stopped                    func()

	Services map[string]interface{}
}

// discoverNodeID - should return the node id for this machine
func discoverNodeID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "node-" + utils.RandomString(2)
	}
	return fmt.Sprint(hostname, "-", utils.RandomString(5))
}

var DefaultConfig = Config{
	Namespace:                  "nucleo-example-ns",
	LogLevel:                   LogLevelInfo,
	LogFormat:                  LogFormatText,
	DiscoverNodeID:             discoverNodeID,
	HeartbeatFrequency:         5 * time.Second,
	HeartbeatTimeout:           15 * time.Second,
	OfflineCheckFrequency:      20 * time.Second,
	OfflineTimeout:             10 * time.Minute,
	WaitForDependenciesTimeout: 2 * time.Second,
	Created:                    func() {},
	Started:                    func() {},
	Stopped:                    func() {},
	RequestTimeout:             3 * time.Second,
}

type Options struct {
	Meta interface{}
}

type BrokerContext interface {
	// Generically type response channel
	Call(actionName string, params interface{}, opts Options) chan interface{}

	RequestID() string

	Logger() *log.Entry

	PublishServices(...interface{})
}

type LoggerFunc func(name string, value string) *log.Entry
type isStartedFunc func() bool
type InstanceIDFunc func() string
type BrokerContextFunc func() BrokerContext
type PublishServicesFunc func(...interface{})
type LocalNodeFunc func() Node
type BrokerDelegates struct {
	InstanceID      InstanceIDFunc
	LocalNode       LocalNodeFunc
	Logger          LoggerFunc
	IsStarted       isStartedFunc
	Config          Config
	BrokerContext   BrokerContextFunc
	PublishServices PublishServicesFunc
}

type Node interface {
	GetID() string

	ExportAsMap() map[string]interface{}
	IsAvailable() bool
	Available()
	Unavailable()
	IsExpired(timeout time.Duration) bool
	Update(id string, info map[string]interface{}) bool

	IncreaseSequence()
	HeartBeat(heartbeat map[string]interface{})
	Publish(service map[string]interface{})
}

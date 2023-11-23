package nucleo

import (
	"fmt"
	"os"
	"time"

	bus "github.com/Bendomey/nucleo-go/emitter"
	"github.com/Bendomey/nucleo-go/utils"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type ForEachFunc func(iterator func(key interface{}, value Payload) bool)

// Payload contains the data sent/return to actions.
// I has convinience methods to read action parameters by name with the right type.
type Payload interface {
	First() Payload
	Sort(field string) Payload
	Remove(fields ...string) Payload
	AddItem(value interface{}) Payload
	Add(field string, value interface{}) Payload
	AddMany(map[string]interface{}) Payload
	MapArray() []map[string]interface{}
	RawMap() map[string]interface{}
	Bson() bson.M
	BsonArray() bson.A
	Map() map[string]Payload
	Exists() bool
	IsError() bool
	Error() error
	ErrorPayload() Payload
	Value() interface{}
	ValueArray() []interface{}
	Int() int
	IntArray() []int
	Int64() int64
	Int64Array() []int64
	Uint() uint64
	UintArray() []uint64
	Float32() float32
	Float32Array() []float32
	Float() float64
	FloatArray() []float64
	String() string
	StringArray() []string
	Bool() bool
	BoolArray() []bool
	ByteArray() []byte
	Time() time.Time
	TimeArray() []time.Time
	Array() []Payload
	At(index int) Payload
	Len() int
	Get(path string, defaultValue ...interface{}) Payload
	//Only return a payload containing only the field specified
	Only(path string) Payload
	IsArray() bool
	IsMap() bool
	ForEach(iterator func(key interface{}, value Payload) bool)
	MapOver(tranform func(in Payload) Payload) Payload
}

// ActionSchema is used by the validation engine to check if parameters sent to the action are valid.
type ActionParams = Payload

type ObjectSchema struct {
	Source interface{}
}

type Action struct {
	Name        string
	Handler     ActionHandler
	Params      map[string]interface{}
	Settings    map[string]interface{}
	Description string
}

type Event struct {
	Name    string
	Group   string
	Handler EventHandler
}

type ServiceSchema struct {
	Name         string
	Version      string
	Dependencies []string
	Settings     map[string]interface{}
	Metadata     map[string]interface{}
	Hooks        map[string]interface{}
	Mixins       []Mixin
	Actions      []Action
	Events       []Event
	Created      CreatedFunc
	Started      LifecycleFunc
	Stopped      LifecycleFunc
}

type Mixin struct {
	Name         string
	Dependencies []string
	Settings     map[string]interface{}
	Metadata     map[string]interface{}
	Hooks        map[string]interface{}
	Actions      []Action
	Events       []Event
	Created      CreatedFunc
	Started      LifecycleFunc
	Stopped      LifecycleFunc
}

type TransporterFactoryFunc func() interface{}
type StrategyFactoryFunc func() interface{}

type ValidatorType string

const (
	ValidatorGo  ValidatorType = "GoValidator"
	ValidatorJoi ValidatorType = "Joi"
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

type StrategyType string

const (
	StrategyRoundRobin StrategyType = "RoundRobin"
	StrategyRandom     StrategyType = "Random"
)

type SerializerType string

const (
	SerializerJSON SerializerType = "JSON"
)

type Config struct {
	LogLevel                   LogLevelType
	LogFormat                  LogFormatType
	Serializer                 SerializerType
	Validator                  ValidatorType
	DiscoverNodeID             func() string
	Transporter                string
	TransporterFactory         TransporterFactoryFunc
	Strategy                   StrategyType
	StrategyFactory            StrategyFactoryFunc
	HeartbeatFrequency         time.Duration
	HeartbeatTimeout           time.Duration
	OfflineCheckFrequency      time.Duration
	OfflineTimeout             time.Duration
	NeighboursCheckTimeout     time.Duration
	WaitForDependenciesTimeout time.Duration
	Middlewares                []Middlewares
	Namespace                  string
	RequestTimeout             time.Duration
	MCallTimeout               time.Duration
	RetryPolicy                RetryPolicy
	MaxCallLevel               int
	Metrics                    bool
	MetricsRate                float32
	DisableInternalServices    bool
	DisableInternalMiddlewares bool
	DontWaitForNeighbours      bool
	WaitForNeighboursInterval  time.Duration
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
	Namespace:                  "",
	LogLevel:                   LogLevelInfo,
	LogFormat:                  LogFormatText,
	Serializer:                 SerializerJSON,
	Validator:                  ValidatorGo,
	DiscoverNodeID:             discoverNodeID,
	Transporter:                "MEMORY",
	Strategy:                   StrategyRandom,
	HeartbeatFrequency:         5 * time.Second,
	HeartbeatTimeout:           15 * time.Second,
	OfflineCheckFrequency:      20 * time.Second,
	OfflineTimeout:             10 * time.Minute,
	DontWaitForNeighbours:      true,
	NeighboursCheckTimeout:     2 * time.Second,
	WaitForDependenciesTimeout: 2 * time.Second,
	Metrics:                    false,
	MetricsRate:                1,
	DisableInternalServices:    false,
	DisableInternalMiddlewares: false,
	Created:                    func() {},
	Started:                    func() {},
	Stopped:                    func() {},
	MaxCallLevel:               100,
	RetryPolicy: RetryPolicy{
		Enabled: false,
	},
	RequestTimeout:            3 * time.Second,
	MCallTimeout:              5 * time.Second,
	WaitForNeighboursInterval: 200 * time.Millisecond,
}

type RetryPolicy struct {
	Enabled  bool
	Retries  int
	Delay    int
	MaxDelay int
	Factor   int
	Check    func(error) bool
}

type Options struct {
	Meta   Payload
	NodeID string
}

type Context interface {
	//context methods used by services
	MCall(map[string]map[string]interface{}) chan map[string]Payload
	Call(actionName string, params interface{}, opts ...Options) chan Payload
	Emit(eventName string, params interface{}, groups ...string)
	Broadcast(eventName string, params interface{}, groups ...string)
	Logger() *log.Entry

	Payload() Payload
	Meta() Payload
}

type BrokerContext interface {
	Call(actionName string, params interface{}, opts ...Options) chan Payload
	Emit(eventName string, params interface{}, groups ...string)

	ChildActionContext(actionName string, params Payload, opts ...Options) BrokerContext
	ChildEventContext(eventName string, params Payload, groups []string, broadcast bool) BrokerContext

	ActionName() string
	EventName() string
	Payload() Payload
	SetPayloadSchema(schema map[string]interface{})
	PayloadSchema() map[string]interface{}
	Groups() []string
	IsBroadcast() bool
	Caller() string

	//export context info in a map[string]
	AsMap() map[string]interface{}

	SetTargetNodeID(targetNodeID string)
	TargetNodeID() string

	ID() string
	RequestID() string
	Meta() Payload
	UpdateMeta(Payload)
	Logger() *log.Entry

	Publish(...interface{})
	WaitFor(services ...string) error
}

type ActionHandler func(context Context, params Payload) interface{}
type EventHandler func(context Context, params Payload)
type CreatedFunc func(ServiceSchema, *log.Entry)
type LifecycleFunc func(BrokerContext, ServiceSchema)

type LoggerFunc func(name string, value string) *log.Entry
type BusFunc func() *bus.Emitter
type isStartedFunc func() bool
type LocalNodeFunc func() Node
type InstanceIDFunc func() string
type ActionDelegateFunc func(context BrokerContext, opts ...Options) chan Payload
type EmitEventFunc func(context BrokerContext)
type ServiceForActionFunc func(string) []*ServiceSchema
type MultActionDelegateFunc func(callMaps map[string]map[string]interface{}) chan map[string]Payload
type BrokerContextFunc func() BrokerContext
type MiddlewareHandlerFunc func(name string, params interface{}) interface{}
type PublishServicesFunc func(...interface{})
type WaitForFunc func(...string) error
type MiddlewareHandler func(params interface{}, next func(...interface{}))

type Middlewares map[string]MiddlewareHandler

type BrokerDelegates struct {
	InstanceID         InstanceIDFunc
	LocalNode          LocalNodeFunc
	Logger             LoggerFunc
	Bus                BusFunc
	IsStarted          isStartedFunc
	Config             Config
	MultActionDelegate MultActionDelegateFunc
	ActionDelegate     ActionDelegateFunc
	EmitEvent          EmitEventFunc
	BroadcastEvent     EmitEventFunc
	HandleRemoteEvent  EmitEventFunc
	ServiceForAction   ServiceForActionFunc
	BrokerContext      BrokerContextFunc
	MiddlewareHandler  MiddlewareHandlerFunc
	PublishServices    PublishServicesFunc
	WaitFor            WaitForFunc
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

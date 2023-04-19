package nucleo

import (
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/Bendomey/nucleo-go/nucleo/util"
	log "github.com/sirupsen/logrus"
)

type ForEachFunc func(iterator func(key interface{}, value Payload) bool)

// Payload contains the data sent/received in actions. These utility functions helps developers gets the most out of it.
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

type ActionHandler func(context Context, params Payload) interface{}

// type EventHandler func(context Context, params Payload)
type CreatedFunc func(ServiceSchema, *log.Entry)
type LifecycleFunc func(BrokerContext, ServiceSchema)

type LoggerFunc func(name string, value string) *log.Entry
type isStartedFunc func() bool
type LocalNodeFunc func() Node
type InstanceIDFunc func() string
type ActionDelegateFunc func(context BrokerContext, opts ...Options) chan Payload
type EmitEventFunc func(context BrokerContext)
type ServiceForActionFunc func(string) []*ServiceSchema
type MultActionDelegateFunc func(callMaps map[string]map[string]interface{}) chan map[string]Payload
type BrokerContextFunc func() BrokerContext
type MiddlewareHandlerFunc func(name string, params interface{}) interface{}
type PublishFunc func(...interface{})
type WaitForFunc func(...string) error
type MiddlewareHandler func(params interface{}, next func(...interface{}))

type Middlewares map[string]MiddlewareHandler

type Middleware interface {
	CallHandlers(name string, params interface{}) interface{}
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

// ActionSchema is used by the validation engine to check if parameters sent to the action are valid.
type ActionSchema interface {
}

type ObjectSchema struct {
	Source interface{}
}

type Action struct {
	Name        string
	Handler     ActionHandler
	Schema      ActionSchema
	Settings    map[string]interface{}
	Description string
}

// TODO: Work on Events later
// type Event struct {
// 	Name    string
// 	Group   string
// 	Handler EventHandler
// }

type ServiceSchema struct {
	Name         string
	Version      string
	Dependencies []string
	Settings     map[string]interface{}
	Metadata     map[string]interface{}
	Hooks        map[string]interface{}
	Mixins       []Mixin
	Actions      []Action
	// Events       []Event
	Created CreatedFunc
	Started LifecycleFunc
	Stopped LifecycleFunc
}

type Mixin struct {
	Name         string
	Dependencies []string
	Settings     map[string]interface{}
	Metadata     map[string]interface{}
	Hooks        map[string]interface{}
	Actions      []Action
	// Events       []Event
	Created CreatedFunc
	Started LifecycleFunc
	Stopped LifecycleFunc
}

type TransporterFactoryFunc func() interface{}
type StrategyFactoryFunc func() interface{}

type Config struct {
	LogLevel                   string
	LogFormat                  string
	DiscoverNodeID             func() string
	Transporter                string
	TransporterFactory         TransporterFactoryFunc
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
	// RetryPolicy                RetryPolicy
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

var DefaultConfig = Config{
	LogLevel:                   "INFO",
	LogFormat:                  "TEXT",
	DiscoverNodeID:             discoverNodeID,
	Transporter:                "MEMORY",
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
	// RetryPolicy: RetryPolicy{
	// 	Enabled: false,
	// },
	RequestTimeout:            3 * time.Second,
	MCallTimeout:              5 * time.Second,
	WaitForNeighboursInterval: 200 * time.Millisecond,
}

// discoverNodeID - should return the node id for this machine
func discoverNodeID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "node-" + util.RandomString(2)
	}
	return fmt.Sprint(hostname, "-", util.RandomString(5))
}

// TODO: Work on Retryability later.
// type RetryPolicy struct {
// 	Enabled  bool
// 	Retries  int
// 	Delay    int
// 	MaxDelay int
// 	Factor   int
// 	Check    func(error) bool
// }

type Options struct {
	Meta Payload
	// NodeID string
}

type Context interface {
	//context methods used by services
	MCall(map[string]map[string]interface{}) chan map[string]Payload
	Call(actionName string, params interface{}, opts ...Options) chan Payload
	// Emit(eventName string, params interface{}, groups ...string)
	// Broadcast(eventName string, params interface{}, groups ...string)
	Logger() *log.Entry

	Payload() Payload
	Meta() Payload
}

type BrokerContext interface {
	Call(actionName string, params interface{}, opts ...Options) chan Payload
	// Emit(eventName string, params interface{}, groups ...string)

	ChildActionContext(actionName string, params Payload, opts ...Options) BrokerContext
	// ChildEventContext(eventName string, params Payload, groups []string, broadcast bool) BrokerContext

	ActionName() string
	// EventName() string
	Payload() Payload
	Groups() []string
	// IsBroadcast() bool
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

type BrokerDelegates struct {
	InstanceID         InstanceIDFunc
	LocalNode          LocalNodeFunc
	Logger             LoggerFunc
	IsStarted          isStartedFunc
	Config             Config
	MultActionDelegate MultActionDelegateFunc
	ActionDelegate     ActionDelegateFunc
	// EmitEvent          EmitEventFunc
	// BroadcastEvent     EmitEventFunc
	HandleRemoteEvent EmitEventFunc
	ServiceForAction  ServiceForActionFunc
	BrokerContext     BrokerContextFunc
	MiddlewareHandler MiddlewareHandlerFunc
	Publish           PublishFunc
	WaitFor           WaitForFunc
}

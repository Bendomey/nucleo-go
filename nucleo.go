package nucleo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

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

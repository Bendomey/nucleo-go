// Dump from https://github.com/moleculer-go/goemitter/blob/master/goemitter.go
package Emitter

import (
	"reflect"
	"strings"
	"sync"
)

// wildcard helper
func eventMatchPattern(eventName, pattern []rune) bool {
	for len(pattern) > 0 {
		switch pattern[0] {
		case '*':
			return eventMatchPattern(eventName, pattern[1:]) || (len(eventName) > 0 && eventMatchPattern(eventName[1:], pattern))

		default:
			if len(eventName) == 0 || eventName[0] != pattern[0] {
				return false
			}
		}

		eventName = eventName[1:]
		pattern = pattern[1:]
	}

	return len(eventName) == 0 && len(pattern) == 0
}

// Emitter - our listeners container
type Emitter struct {
	listeners map[interface{}][]Listener
	mutex     *sync.Mutex
}

// Listener - our callback container and whether it will run once or not
type Listener struct {
	callback func(...interface{})
	once     bool
}

// Construct() - create a new instance of Emitter
func Construct() *Emitter {
	return &Emitter{
		make(map[interface{}][]Listener),
		&sync.Mutex{},
	}
}

// Destruct() - free memory from an emitter instance
func (emitter *Emitter) Destruct() {
	emitter = nil
}

// AddListener() - register a new listener on the specified event
func (emitter *Emitter) AddListener(event string, callback func(...interface{})) *Emitter {
	return emitter.On(event, callback)
}

// On() - register a new listener on the specified event
func (emitter *Emitter) On(event string, callback func(...interface{})) *Emitter {
	emitter.mutex.Lock()
	if _, ok := emitter.listeners[event]; !ok {
		emitter.listeners[event] = []Listener{}
	}
	emitter.listeners[event] = append(emitter.listeners[event], Listener{callback, false})
	emitter.mutex.Unlock()

	emitter.EmitSync("newListener", []interface{}{event, callback})
	return emitter
}

// Once() - register a new one-time listener on the specified event
func (emitter *Emitter) Once(event string, callback func(...interface{})) *Emitter {
	emitter.mutex.Lock()
	if _, ok := emitter.listeners[event]; !ok {
		emitter.listeners[event] = []Listener{}
	}
	emitter.listeners[event] = append(emitter.listeners[event], Listener{callback, true})
	emitter.mutex.Unlock()

	emitter.EmitSync("newListener", []interface{}{event, callback})
	return emitter
}

// RemoveListeners() - remove the specified callback from the specified events' listeners
func (emitter *Emitter) RemoveListener(event string, callback func(...interface{})) *Emitter {
	return emitter.removeListenerInternal(event, callback, false)
}

func (emitter *Emitter) removeListenerInternal(event string, callback func(...interface{}), suppress bool) *Emitter {
	emitter.mutex.Lock()

	if _, ok := emitter.listeners[event]; !ok {
		emitter.mutex.Unlock()
		return emitter
	}

	for k, v := range emitter.listeners[event] {
		if reflect.ValueOf(v.callback).Pointer() == reflect.ValueOf(callback).Pointer() {
			emitter.listeners[event] = append(emitter.listeners[event][:k], emitter.listeners[event][k+1:]...)

			emitter.mutex.Unlock()

			if !suppress {
				emitter.EmitSync("removeListener", []interface{}{event, callback})
			}
			return emitter
		}
	}

	emitter.mutex.Unlock()

	return emitter
}

// RemoveAllListeners() - remove all listeners from (all/event)
func (emitter *Emitter) RemoveAllListeners(event interface{}) *Emitter {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	if event == nil {
		emitter.listeners = make(map[interface{}][]Listener)
		return emitter
	}
	if _, ok := emitter.listeners[event]; !ok {
		return emitter
	}
	delete(emitter.listeners, event)
	return emitter
}

// Listeners() - return an array with the registered listeners in the specified event
func (emitter *Emitter) Listeners(event string) []Listener {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	listeners := make([]Listener, 0)

	// add the ones that follow pattern
	for eventPattern, lis := range emitter.listeners {
		shouldAdd := false

		// add generic "**" bound listeners
		shouldAdd = shouldAdd || eventPattern.(string) == "**"
		// add listener bound on full name event
		shouldAdd = shouldAdd || eventPattern.(string) == event
		// add listeners that have matching wildcard pattern
		shouldAdd = shouldAdd ||
			(strings.Contains(eventPattern.(string), "*") &&
				eventMatchPattern([]rune(event), []rune(eventPattern.(string))))

		if shouldAdd {
			listeners = append(listeners, lis...)
		}
	}

	return listeners
}

// ListenersCount() - return the count of listeners in the speicifed event
func (emitter *Emitter) ListenersCount(event string) int {
	return len(emitter.Listeners(event))
}

// EmitSync() - run all listeners of the specified event in synchronous mode
func (emitter *Emitter) EmitSync(event string, args ...interface{}) *Emitter {
	for _, v := range emitter.Listeners(event) {
		if v.once {
			emitter.removeListenerInternal(event, v.callback, true)
		}
		v.callback(args...)
	}

	return emitter
}

// EmitAsync() - run all listeners of the specified event in asynchronous mode using goroutines
func (emitter *Emitter) EmitAsync(event string, args []interface{}) *Emitter {
	for _, v := range emitter.Listeners(event) {
		if v.once {
			emitter.removeListenerInternal(event, v.callback, true)
		}
		go v.callback(args...)
	}
	return emitter
}

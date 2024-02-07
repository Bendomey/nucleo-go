package cache

import (
	"github.com/Bendomey/nucleo-go"
)

type Cacher interface {
	Init()
	GetCacherName() string
	Middlewares() nucleo.Middlewares
	Close()
}

// Resolve returns a new instance of a cache based on the config
func Resolve(config nucleo.Config) (Cacher, bool) {
	if config.Cacher.Type == "" {
		return nil, true
	}

	cacher := config.Cacher

	switch cacher.Type {
	case nucleo.CacherMemory:
		memoryCacher := NewMemoryCacher(cacher)
		return memoryCacher, memoryCacher == nil
	case nucleo.CacherRedis:
		redisCacher := NewRedisCacher(cacher)
		return redisCacher, redisCacher == nil
	}

	return nil, true
}

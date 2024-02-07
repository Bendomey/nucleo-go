package cache

import (
	"github.com/Bendomey/nucleo-go"
	log "github.com/sirupsen/logrus"
)

type Memory struct {
	Prefix  string
	Ttl     int
	Monitor bool
	logger  *log.Entry
}

func NewMemoryCacher(config nucleo.CacherConfig) *Memory {
	return &Memory{
		Monitor: config.Monitor,
		Ttl:     config.Ttl,
		Prefix:  config.Prefix,
		logger:  log.WithField("cacher", "redis"),
	}
}

func (r *Memory) GetCacherName() string {
	return string(nucleo.CacherMemory)
}

func (r *Memory) Init() {
}

func (r *Memory) Middlewares() nucleo.Middlewares {
	return nucleo.Middlewares{}
}

func (r *Memory) Close() {
}

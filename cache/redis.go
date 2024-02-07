package cache

import (
	"github.com/Bendomey/nucleo-go"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

type Redis struct {
	connectionUrl string
	connection    redis.Options
	prefix        string
	ttl           int
	monitor       bool
	redisClient   *redis.Client
	isConnected   bool
	logger        *log.Entry
}

func NewRedisCacher(config nucleo.CacherConfig) *Redis {
	// Not enough redis config to create a redis cacher.
	if config.RedisConnectionUrl == "" && config.RedisConnection.Addr == "" {
		return nil
	}

	return &Redis{
		connectionUrl: config.RedisConnectionUrl,
		connection:    config.RedisConnection,
		prefix:        config.Prefix,
		ttl:           config.Ttl,
		monitor:       config.Monitor,
		logger:        log.WithField("cacher", "redis"),
	}
}

func (r *Redis) GetCacherName() string {
	return string(nucleo.CacherRedis)
}

func (r *Redis) Init() {
	if r.connectionUrl != "" {
		redisClient := redis.NewClient(&redis.Options{
			Addr: r.connectionUrl,
		})

		r.redisClient = redisClient
	}

	if r.connection.Addr != "" {
		redisClient := redis.NewClient(&r.connection)

		r.redisClient = redisClient
	}

	r.isConnected = true
	r.logger.Debugln("Redis Cacher created. Prefix: " + r.prefix)

}

func (r *Redis) Middlewares() nucleo.Middlewares {
	return nucleo.Middlewares{}
}

func (r *Redis) Close() {
	r.redisClient.Close()
	r.redisClient = nil
	r.isConnected = false
}

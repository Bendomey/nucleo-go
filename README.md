# nucleo-go
Progressive microservices framework for distributed systems development in golang.


> Inspired By [MoleculerJs](https://moleculer.services)

> Building on top of [MoleculerGo](https://github.com/moleculer-go)


## Features
- [x] Service Broker
- [x] Transit and Transport
- [x] Actions (request-reply)
- [x] Events
- [x] Mixins
- [x] Load balancing for actions and events (random round-robin)
- [x] Service registry & dynamic service discovery
- [x] Versioned services
- [x] Middlewares
- [x] Transporter: TCP, Nats, Kafka, AMQP
- [x] Serializers: JSON
- [x] Examples

- [ ] Standard Project Template
- [ ] CLI for Project Seed Generation
- [ ] Action validators: Fastest Validator, [Joi](https://github.com/softbrewery/gojoi)
- [ ] More Load balancing implementations (cpu-usage, latency)
- [ ] Fault tolerance features (Circuit Breaker, Bulkhead, Retry, Timeout, Fallback)
- [ ] Built-in caching solution (memory, Redis)
- [ ] More transporters (gRPC, Redis)
- [ ] More serializers (Avro, MsgPack, Protocol Buffer, Thrift)
- [ ] Testssssss

## External Support Features
- [ ] Database Adapters
- [ ] Gateway
- [ ] Metrics Plugins
- [ ] etc...

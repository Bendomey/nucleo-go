#  Dynamic Service Discovery

Nucleo framework has a built-in service responsible for node/service discovery and periodic heartbeat verification. The discovery is dynamic meaning that a node does not need to know anything about other nodes on startup. When a node/service/actions starts, it will announce it’s presence to the service discovery and that will be saved in the registry. In case of a node crash (or stop) other nodes will detect it and remove the affected services from their registry. This way the following requests will be routed to live nodes or users will be signaled.

# Service Registry

Nucleo has a built-in service registry module. It stores all information about services, actions, event listeners and nodes. When you call a service or emit an event, broker asks the registry to look up a node which executes the request. If there are multiple nodes which can serve the request, it uses load-balancing strategy to select the next node.


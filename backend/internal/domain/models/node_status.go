package domainmodels

type NodeStatus string

const (
	NodeStatusOffline NodeStatus = "OFFLINE"
	NodeStatusOnline  NodeStatus = "ONLINE"
)

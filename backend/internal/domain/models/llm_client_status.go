package domainmodels

type LlmClientStatus string

const (
	LlmClientStatusConnected    LlmClientStatus = "CONNECTED"
	LlmClientStatusDisconnected LlmClientStatus = "DISCONNECTED"
)

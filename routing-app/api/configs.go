package api

type HandlerConfig struct {
	HandlerType     string
	HostAddresses   []string
	MaxRetries      int
	TimeoutSeconds  int
	RecoverySeconds int
}

type Host struct {
	Address string
	Healthy bool
}

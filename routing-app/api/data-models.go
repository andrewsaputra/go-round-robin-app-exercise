package api

type AppConfig struct {
	RoutingAlgorithm string
	RequestHandling  RequestHandlingConfig
	HealthCheck      HealthCheckConfig
}

type RequestHandlingConfig struct {
	MaxRetries     int
	TimeoutSeconds int
}

type HealthCheckConfig struct {
	Path            string
	NumRequired     int
	IntervalSeconds int
	TimeoutSeconds  int
}

type Host struct {
	Address            string
	Healthy            bool
	RecentHealthChecks []bool
}

type HandlerResponse struct {
	Code    int
	Message string
	Error   error
}

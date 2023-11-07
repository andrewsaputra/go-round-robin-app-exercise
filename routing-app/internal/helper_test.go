package internal

import (
	"andrewsaputra/routing-app/api"
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockRoundTripper struct {
	mock.Mock
}

func (this *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := this.Called(req)

	return args.Get(0).(*http.Response), args.Error(1)
}

func Helper_ConstructMockHttpClient() *http.Client {
	roundTripper := new(MockRoundTripper)
	roundTripper.On("RoundTrip", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK}, nil)

	client := &http.Client{
		Transport: roundTripper,
	}

	return client
}

func Helper_ConstructHealthCheckConfig() api.HealthCheckConfig {
	return api.HealthCheckConfig{
		Path:            "/status",
		NumRequired:     1,
		IntervalSeconds: 1,
		TimeoutSeconds:  1,
	}
}

func Helper_ConstructHostManager() *HostManager {
	config := Helper_ConstructHealthCheckConfig()
	client := Helper_ConstructMockHttpClient()
	return ConstructHostManager(client, config)
}

func Helper_ConstructRoundRobinRouter() (*RoundRobinRouter, *HostManager) {
	client := Helper_ConstructMockHttpClient()
	hostManager := Helper_ConstructHostManager()
	router := ConstructRoundRobinRouter(client, hostManager, 0)
	return router, hostManager
}

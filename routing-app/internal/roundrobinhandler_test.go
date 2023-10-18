package internal

import (
	"andrewsaputra/routing-app/api"
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var connector = new(MockHttpConnector)

func setupMockRouter(config api.HandlerConfig) (*RoundRobinHandler, *gin.Engine) {
	handler, _ := CreateRoundRobinHandler(config, connector)
	router := gin.Default()
	router.POST("/routejson", handler.RouteJson)

	return handler, router
}

func TestRouteJson_ValidPayload_EchoSuccess(t *testing.T) {
	_, router := setupMockRouter(getBaseHandlerConfig())

	payload := []byte(`{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}`)

	mocker := connector.On("DoPost", mock.Anything, payload).
		Return(getHttpResponse(http.StatusOK, payload), nil)
	defer mocker.Unset()

	request, _ := http.NewRequest("POST", "/routejson", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, string(payload), string(response.Body.Bytes()))
}

func TestRouteJson_NonJsonPayload_ReturnBadRequest(t *testing.T) {
	_, router := setupMockRouter(getBaseHandlerConfig())

	payload := []byte("normal string")
	request, _ := http.NewRequest("POST", "/routejson", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRouteJson_NoHealthyHosts_ReturnInternalServerError(t *testing.T) {
	handler, router := setupMockRouter(getBaseHandlerConfig())
	for i, _ := range handler.Hosts {
		handler.Hosts[i].Healthy = false
	}

	payload := []byte(`{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}`)
	request, _ := http.NewRequest("POST", "/routejson", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestRouteJson_ConnectionFailure_ReturnInternalServerError(t *testing.T) {
	_, router := setupMockRouter(getBaseHandlerConfig())

	payload := []byte(`{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}`)
	mocker := connector.On("DoPost", mock.Anything, payload).
		Return(&http.Response{}, errors.New("mock connection failure"))
	defer mocker.Unset()

	request, _ := http.NewRequest("POST", "/routejson", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestRouteJson_EncounterHostError_ReturnInternalServerError(t *testing.T) {
	_, router := setupMockRouter(getBaseHandlerConfig())

	payload := []byte(`{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}`)
	mocker := connector.On("DoPost", mock.Anything, payload).
		Return(getHttpResponse(http.StatusInternalServerError, payload), nil)
	defer mocker.Unset()

	request, _ := http.NewRequest("POST", "/routejson", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestRouteJson_EncounterHostError_EchoSuccessAfterRetry(t *testing.T) {
	_, router := setupMockRouter(getBaseHandlerConfig())

	payload := []byte(`{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}`)
	mocker := connector.On("DoPost", mock.Anything, payload).
		Return(getHttpResponse(http.StatusInternalServerError, payload), nil).
		Once()

	mocker2 := connector.On("DoPost", mock.Anything, payload).
		Return(getHttpResponse(http.StatusOK, payload), nil).NotBefore(mocker)
	defer mocker2.Unset()

	request, _ := http.NewRequest("POST", "/routejson", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, string(payload), string(response.Body.Bytes()))
}

func TestHostCycling_AllHostsHealthy_ReturnSequential(t *testing.T) {
	config := getBaseHandlerConfig()
	handler, _ := CreateRoundRobinHandler(config, connector)
	n := len(config.HostAddresses)
	for i := 0; i < 3*n; i++ {
		host, _ := handler.nextHost()
		expected := config.HostAddresses[i%n]
		assert.Equal(t, expected, host.Address)
	}
}

func TestHostCycling_HasUnhealthyHost_SkipUnhealthyHost(t *testing.T) {
	config := getBaseHandlerConfig()
	handler, _ := CreateRoundRobinHandler(config, connector)
	handler.Hosts[1].Healthy = false
	validHosts := []string{config.HostAddresses[0], config.HostAddresses[2]}
	n := len(validHosts)
	for i := 0; i < 3*n; i++ {
		host, _ := handler.nextHost()
		assert.Equal(t, validHosts[i%n], host.Address)
	}
}

func TestHostCycling_NoHealthyHosts_ReturnError(t *testing.T) {
	config := getBaseHandlerConfig()
	handler, _ := CreateRoundRobinHandler(config, connector)
	for i, _ := range handler.Hosts {
		handler.Hosts[i].Healthy = false
	}

	host, err := handler.nextHost()
	assert.Nil(t, host)
	assert.Error(t, err)
}

func TestHostCycling_SingleHost_ReturnHost(t *testing.T) {
	config := getBaseHandlerConfig()
	config.HostAddresses = []string{"host1"}
	handler, _ := CreateRoundRobinHandler(config, connector)
	n := len(config.HostAddresses)
	for i := 0; i < 3*n; i++ {
		host, _ := handler.nextHost()
		expected := config.HostAddresses[i%n]
		assert.Equal(t, expected, host.Address)
	}
}

func TestMarkAsUnhealthy_RecoverySuccess(t *testing.T) {
	config := getBaseHandlerConfig()
	handler, _ := CreateRoundRobinHandler(config, connector)

	host, _ := handler.nextHost()
	assert.True(t, host.Healthy)

	handler.markHostAsUnhealthy(host)
	assert.False(t, host.Healthy)

	time.Sleep(handler.RecoveryInterval + 200*time.Millisecond)
	assert.True(t, host.Healthy)
}

type MockHttpConnector struct {
	mock.Mock
}

func (this *MockHttpConnector) DoPost(url string, body []byte) (*http.Response, error) {
	args := this.Called(url, body)
	return args.Get(0).(*http.Response), args.Error(1)
}

func getHttpResponse(statusCode int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

func getBaseHandlerConfig() api.HandlerConfig {
	return api.HandlerConfig{
		HandlerType:     "RoundRobin",
		HostAddresses:   []string{"host1", "host2", "host3"},
		MaxRetries:      2,
		TimeoutSeconds:  1,
		RecoverySeconds: 1,
	}
}

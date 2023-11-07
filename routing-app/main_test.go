package main

import (
	"andrewsaputra/routing-app/api"
	"andrewsaputra/routing-app/internal"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReadAppConfig_WithValidPath_ReturnConfig(t *testing.T) {
	config, err := readAppConfig("configs/appconfig.json")

	assert.NotNil(t, config)
	assert.Nil(t, err)
}

func TestReadAppConfig_WithInvalidPath_ReturnError(t *testing.T) {
	config, err := readAppConfig("invalid-path")

	assert.Nil(t, config)
	assert.NotNil(t, err)
}

func TestSetupAppHandler_WithRoundRobinAlgoritm_ReturnHandler(t *testing.T) {
	config := &api.AppConfig{
		RoutingAlgorithm: "RoundRobin",
		RequestHandling:  api.RequestHandlingConfig{MaxRetries: 0, TimeoutSeconds: 5},
		HealthCheck:      api.HealthCheckConfig{Path: "/status", NumRequired: 1, IntervalSeconds: 1, TimeoutSeconds: 1},
	}
	handler, err := setupHandler(config)

	assert.Nil(t, err)
	assert.IsType(t, &internal.RoundRobinRouter{}, handler.RequestRouter)
}

func TestSetupAppHandler_WithUnknownAlgoritm_ReturnError(t *testing.T) {
	config := &api.AppConfig{
		RoutingAlgorithm: "unknown",
		RequestHandling:  api.RequestHandlingConfig{MaxRetries: 0, TimeoutSeconds: 5},
		HealthCheck:      api.HealthCheckConfig{Path: "/status", NumRequired: 1, IntervalSeconds: 1, TimeoutSeconds: 1},
	}
	handler, err := setupHandler(config)

	assert.NotNil(t, err)
	assert.Nil(t, handler)
}

func TestSetupRouter_RegisterRoutes_StatusCheckSuccess(t *testing.T) {
	handler := new(MockHandler)
	router := setupRouter(handler)

	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/status", nil)
	router.ServeHTTP(response, request)

	var responseJson map[string]any
	json.Unmarshal(response.Body.Bytes(), &responseJson)

	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "Healthy", responseJson["status"])
}

func TestSetupRouter_RegisterRoutes_HandlerFunctionsCalled(t *testing.T) {
	handler := new(MockHandler)
	handler.On("RegisterHost", mock.Anything).Return()
	handler.On("DeregisterHost", mock.Anything).Return()
	handler.On("ForwardRequest", mock.Anything).Return()

	router := setupRouter(handler)
	payload := []byte(`{"key":"value"}`)

	request, _ := http.NewRequest("POST", "/registerhost", bytes.NewReader(payload))
	router.ServeHTTP(httptest.NewRecorder(), request)
	handler.AssertCalled(t, "RegisterHost", mock.Anything)

	request, _ = http.NewRequest("POST", "/deregisterhost", bytes.NewReader(payload))
	router.ServeHTTP(httptest.NewRecorder(), request)
	handler.AssertCalled(t, "DeregisterHost", mock.Anything)

	request, _ = http.NewRequest("POST", "/other", bytes.NewReader(payload))
	router.ServeHTTP(httptest.NewRecorder(), request)
	handler.AssertCalled(t, "ForwardRequest", mock.Anything)
}

type MockHandler struct {
	mock.Mock
}

func (this *MockHandler) RegisterHost(c *gin.Context) {
	this.Called(c)
}

func (this *MockHandler) DeregisterHost(c *gin.Context) {
	this.Called(c)
}

func (this *MockHandler) ForwardRequest(c *gin.Context) {
	this.Called(c)
}

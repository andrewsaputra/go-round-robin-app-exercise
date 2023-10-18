package main

import (
	"andrewsaputra/routing-app/api"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusCheck_ReturnHealthy(t *testing.T) {
	router := setupRouter()

	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/status", nil)
	router.ServeHTTP(response, request)

	var responseJson map[string]any
	json.Unmarshal(response.Body.Bytes(), &responseJson)

	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "Healthy", responseJson["status"])
}

func TestSetupHandler_RoundRobin_ReturnHandler(t *testing.T) {
	handler, err := setupHandler(GetBaseHandlerConfig())

	assert.NotNil(t, handler)
	assert.Nil(t, err)
}

func TestSetupHandler_HasInvalidConfigValue_ReturnError(t *testing.T) {
	config := GetBaseHandlerConfig()
	config.HostAddresses = []string{}
	handler, err := setupHandler(config)

	assert.Nil(t, handler)
	assert.Error(t, err)
}

func TestSetupHandler_UnknownHandlerType_ReturnError(t *testing.T) {
	config := GetBaseHandlerConfig()
	config.HandlerType = "Unknown"
	handler, err := setupHandler(config)

	assert.Nil(t, handler)
	assert.Error(t, err)
}

func GetBaseHandlerConfig() api.HandlerConfig {
	return api.HandlerConfig{
		HandlerType:     "RoundRobin",
		HostAddresses:   []string{"host1", "host2", "host3"},
		MaxRetries:      2,
		TimeoutSeconds:  1,
		RecoverySeconds: 1,
	}
}

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestGetPort_NoPortArgs_ReturnDefaultPort(t *testing.T) {
	port := getPort()
	assert.Equal(t, defaultPort, port)
}

func TestGetPort_HasPortArgs_ReturnSpecifiedPort(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	targetPort := "7777"
	os.Args = []string{"cmd", targetPort}
	port := getPort()
	assert.Equal(t, targetPort, port)
}

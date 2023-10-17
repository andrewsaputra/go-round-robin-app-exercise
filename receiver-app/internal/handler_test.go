package internal

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var router *gin.Engine

func TestMain(t *testing.M) {
	handler := ApiHandler{}
	router = gin.Default()
	router.POST("/echojson", handler.EchoJson)

	t.Run()
}

func TestEchoJson_ValidPayload_ReturnSuccess(t *testing.T) {
	payload := []byte(`{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}`)
	request, _ := http.NewRequest("POST", "/echojson", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, string(payload), string(response.Body.Bytes()))
}

func TestEchoJson_NonJsonPayload_ReturnBadRequest(t *testing.T) {
	payload := []byte(`not a json string`)
	request, _ := http.NewRequest("POST", "/echojson", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

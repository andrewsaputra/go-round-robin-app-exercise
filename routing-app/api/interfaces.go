package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RequestRouter interface {
	ForwardRequest(*http.Request) (*http.Response, error)
}

type Handler interface {
	RegisterHost(c *gin.Context)
	DeregisterHost(c *gin.Context)
	ForwardRequest(c *gin.Context)
}

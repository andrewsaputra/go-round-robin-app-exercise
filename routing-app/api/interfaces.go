package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	RouteJson(c *gin.Context)
}

type HttpConnector interface {
	DoPost(url string, body []byte) (*http.Response, error)
}

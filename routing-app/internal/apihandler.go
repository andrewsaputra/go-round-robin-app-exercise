package internal

import (
	"andrewsaputra/routing-app/api"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ModifyHostRequest struct {
	HostAddress string
}

func ConstructApiHandler(hostManager *HostManager, requestRouter api.RequestRouter) *ApiHandler {
	return &ApiHandler{
		HostManager:   hostManager,
		RequestRouter: requestRouter,
	}
}

type ApiHandler struct {
	HostManager   *HostManager
	RequestRouter api.RequestRouter
}

func (this *ApiHandler) RegisterHost(c *gin.Context) {
	var body ModifyHostRequest
	err := c.BindJSON(&body)
	if err != nil {
		this.handleResponse(c, api.HandlerResponse{
			Code:    http.StatusBadRequest,
			Message: "bad payload request",
		})
		return
	}

	this.handleResponse(c, this.HostManager.RegisterHost(body.HostAddress))
}

func (this *ApiHandler) DeregisterHost(c *gin.Context) {
	var body ModifyHostRequest
	err := c.BindJSON(&body)
	if err != nil {
		this.handleResponse(c, api.HandlerResponse{
			Code:    http.StatusBadRequest,
			Message: "bad payload request",
		})
		return
	}

	this.handleResponse(c, this.HostManager.DeregisterHost(body.HostAddress))
}

func (this *ApiHandler) ForwardRequest(c *gin.Context) {
	resp, err := this.RequestRouter.ForwardRequest(c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

// Private Functions

func (this *ApiHandler) handleResponse(c *gin.Context, response api.HandlerResponse) {
	if response.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": response.Error.Error()})
		return
	}

	c.JSON(response.Code, gin.H{"message": response.Message})
}

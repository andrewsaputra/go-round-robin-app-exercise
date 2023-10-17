package internal

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ApiHandler struct {
}

func (this *ApiHandler) EchoJson(c *gin.Context) {
	rawBytes, err := c.GetRawData()
	if err != nil || !json.Valid(rawBytes) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload must be valid json"})
		return
	}

	c.Data(http.StatusOK, "application/json", rawBytes)
}

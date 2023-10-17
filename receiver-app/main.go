package main

import (
	"andrewsaputra/receiver-app/internal"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const defaultPort = "4000"

var startTime time.Time = time.Now()

func main() {
	router := setupRouter()
	router.Run(":" + getPort())
}

func setupRouter() *gin.Engine {
	handler := internal.ApiHandler{}

	router := gin.Default()
	router.GET("/status", statusCheck)
	router.POST("/echojson", handler.EchoJson)

	return router
}

func getPort() string {
	if len(os.Args) >= 2 {
		if _, err := strconv.Atoi(os.Args[1]); err == nil {
			return os.Args[1]
		}
	}

	return defaultPort
}

func statusCheck(c *gin.Context) {
	response := map[string]any{}
	response["status"] = "Healthy"
	response["startedAt"] = startTime.Format(time.RFC1123Z)

	c.JSON(http.StatusOK, response)
}

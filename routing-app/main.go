package main

import (
	"andrewsaputra/routing-app/api"
	"andrewsaputra/routing-app/internal"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const defaultPort = "3000"

var startTime time.Time = time.Now()

func main() {
	rawConfig, err := os.ReadFile("configs/handlerconfig.json")
	if err != nil {
		log.Fatal(err)
	}

	var config api.HandlerConfig
	json.Unmarshal(rawConfig, &config)
	handler, err := setupHandler(config)
	if err != nil {
		log.Fatal(err)
	}

	router := setupRouter()
	router.POST("/routejson", handler.RouteJson)
	router.Run(":" + defaultPort)
}

func setupHandler(config api.HandlerConfig) (api.Handler, error) {
	switch config.HandlerType {
	case "RoundRobin":
		timeout := time.Duration(config.TimeoutSeconds) * time.Second
		httpClient := &http.Client{
			Timeout: timeout,
		}

		connector := internal.CreateHttpConnectorImpl(httpClient)
		handler, err := internal.CreateRoundRobinHandler(config, connector)
		if err != nil {
			return nil, err
		}

		return handler, nil
	default:
		message := fmt.Sprintf("unsupported handler type : %v\n", config.HandlerType)
		return nil, errors.New(message)
	}
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/status", statusCheck)
	return router
}

func statusCheck(c *gin.Context) {
	response := map[string]any{}
	response["status"] = "Healthy"
	response["startedAt"] = startTime.Format(time.RFC1123Z)

	c.JSON(http.StatusOK, response)
}

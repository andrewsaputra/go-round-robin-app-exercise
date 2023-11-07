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
	appConfig, err := readAppConfig("configs/appconfig.json")
	if err != nil {
		log.Fatal(err)
	}

	appHandler, err := setupHandler(appConfig)
	if err != nil {
		log.Fatal(err)
	}

	router := setupRouter(appHandler)
	router.Run(":" + defaultPort)
}

func readAppConfig(path string) (*api.AppConfig, error) {
	rawConfig, _ := os.ReadFile(path)
	var config *api.AppConfig
	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func setupHandler(config *api.AppConfig) (*internal.ApiHandler, error) {
	hostManager := internal.ConstructHostManager(
		&http.Client{
			Timeout: time.Duration(config.HealthCheck.TimeoutSeconds) * time.Second,
		},
		config.HealthCheck,
	)

	var requestRouter api.RequestRouter
	switch config.RoutingAlgorithm {
	case "RoundRobin":
		requestRouter = internal.ConstructRoundRobinRouter(
			&http.Client{
				Timeout: time.Duration(config.RequestHandling.TimeoutSeconds) * time.Second,
			},
			hostManager,
			config.RequestHandling.MaxRetries,
		)
	default:
		msg := fmt.Sprintln("unsupported routing algorithm", config.RoutingAlgorithm)
		return nil, errors.New(msg)
	}

	return internal.ConstructApiHandler(hostManager, requestRouter), nil
}

func setupRouter(handler api.Handler) *gin.Engine {
	router := gin.Default()
	router.GET("/status", statusCheck)
	router.POST("/registerhost", handler.RegisterHost)
	router.POST("/deregisterhost", handler.DeregisterHost)
	router.NoRoute(handler.ForwardRequest)

	return router
}

func statusCheck(c *gin.Context) {
	response := map[string]any{}
	response["status"] = "Healthy"
	response["startedAt"] = startTime.Format(time.RFC1123Z)

	c.JSON(http.StatusOK, response)
}

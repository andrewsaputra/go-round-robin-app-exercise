package internal

import (
	"andrewsaputra/routing-app/api"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RoundRobinHandler struct {
	Connector        api.HttpConnector
	Hosts            []api.Host
	RecoveryInterval time.Duration
	MaxRetries       int
	TargetIndex      int
}

func CreateRoundRobinHandler(config api.HandlerConfig, connector api.HttpConnector) (*RoundRobinHandler, error) {
	if len(config.HostAddresses) == 0 {
		return nil, errors.New("at least 1 host address must be provided")
	}

	hosts := []api.Host{}
	for _, addr := range config.HostAddresses {
		hosts = append(hosts, api.Host{Address: addr, Healthy: true})
	}

	return &RoundRobinHandler{
		Connector:        connector,
		Hosts:            hosts,
		RecoveryInterval: time.Duration(config.RecoverySeconds) * time.Second,
		MaxRetries:       config.MaxRetries,
		TargetIndex:      0,
	}, nil
}

func (this *RoundRobinHandler) RouteJson(c *gin.Context) {
	payload, err := c.GetRawData()
	if err != nil || !json.Valid(payload) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload must be valid json"})
		return
	}

	numAttempts := 0
	for numAttempts <= this.MaxRetries {
		host, err := this.nextHost()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no available healthy server to process request. please try again after a while."})
			return
		}

		response, err := this.Connector.DoPost(host.Address+"/echojson", payload)
		if err != nil {
			this.markHostAsUnhealthy(host)
			numAttempts++
			continue
		}
		defer response.Body.Close()

		responseBody, err := io.ReadAll(response.Body)
		if err != nil || response.StatusCode == http.StatusInternalServerError {
			this.markHostAsUnhealthy(host)
			numAttempts++
			continue
		}

		c.Data(response.StatusCode, "application/json", responseBody)
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "server encountered error. please try again after a while."})
}

func (this *RoundRobinHandler) nextHost() (*api.Host, error) {
	var host *api.Host
	numHosts := len(this.Hosts)

	switch numHosts {
	case 1:
		host = &this.Hosts[0]
	default:
		startingIndex := this.TargetIndex
		for {
			host = &this.Hosts[this.TargetIndex]
			this.TargetIndex++
			if this.TargetIndex >= numHosts {
				this.TargetIndex = 0
			}

			if host.Healthy || this.TargetIndex == startingIndex {
				break
			}
		}
	}

	if !host.Healthy {
		return nil, errors.New("no available healthy targets")
	}

	return host, nil
}

func (this *RoundRobinHandler) markHostAsUnhealthy(host *api.Host) {
	host.Healthy = false

	go func() {
		time.Sleep(this.RecoveryInterval)
		host.Healthy = true
	}()
}

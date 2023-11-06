package internal

import (
	"andrewsaputra/routing-app/api"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func ConstructHostManager(client *http.Client, healthCheckConfig api.HealthCheckConfig) *HostManager {
	manager := &HostManager{
		hosts:         []api.Host{},
		client:        client,
		numRequiredHC: healthCheckConfig.NumRequired,
		hcPath:        healthCheckConfig.Path,
	}

	hcInterval := time.Duration(healthCheckConfig.IntervalSeconds) * time.Second
	go manager.scheduleHealthChecks(hcInterval)

	return manager
}

type HostManager struct {
	hosts         []api.Host
	client        *http.Client
	numRequiredHC int
	hcPath        string
	lock          sync.RWMutex
}

func (this *HostManager) RegisterHost(hostAddress string) api.HandlerResponse {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, host := range this.hosts {
		if host.Address == hostAddress {
			return api.HandlerResponse{
				Code:    http.StatusBadRequest,
				Message: "Duplicate host address detected",
			}
		}
	}

	this.hosts = append(this.hosts, api.Host{
		Address:            hostAddress,
		Healthy:            false,
		RecentHealthChecks: []bool{},
	})

	return api.HandlerResponse{
		Code:    http.StatusOK,
		Message: "Successful registration",
	}
}

func (this *HostManager) DeregisterHost(hostAddress string) api.HandlerResponse {
	this.lock.Lock()
	defer this.lock.Unlock()

	for i, host := range this.hosts {
		if host.Address == hostAddress {
			this.hosts = append(this.hosts[:i], this.hosts[i+1:]...)
			return api.HandlerResponse{
				Code:    http.StatusOK,
				Message: "Successful deregistration",
			}
		}
	}

	return api.HandlerResponse{
		Code:    http.StatusBadRequest,
		Message: "Host address not found",
	}
}

func (this *HostManager) GetEligibleHosts() []api.Host {
	this.lock.RLock()
	defer this.lock.RUnlock()

	result := []api.Host{}
	for _, host := range this.hosts {
		if host.Healthy {
			result = append(result, host)
		}
	}

	if len(result) > 0 {
		return result
	}

	return this.hosts
}

// Private Functions

func (this *HostManager) scheduleHealthChecks(duration time.Duration) {
	ticker := time.NewTicker(duration)

	for _ = range ticker.C {
		for i, _ := range this.hosts {
			go this.evaluateHostHealth(&this.hosts[i])
		}
	}
}

func (this *HostManager) evaluateHostHealth(host *api.Host) {
	url := host.Address + this.hcPath
	response, err := this.client.Get(url)
	var isHealthy bool
	if err != nil {
		isHealthy = false
	} else {
		defer response.Body.Close()
		isHealthy = response.StatusCode == http.StatusOK
	}

	this.lock.Lock()
	defer this.lock.Unlock()

	if len(host.RecentHealthChecks) > 0 {
		curr := host.RecentHealthChecks[0]
		if isHealthy != curr {
			host.RecentHealthChecks = []bool{}
		}
	}
	host.RecentHealthChecks = append(host.RecentHealthChecks, isHealthy)

	if len(host.RecentHealthChecks) == this.numRequiredHC {
		isHealthy = host.RecentHealthChecks[0]
		if host.Healthy != isHealthy {
			host.Healthy = isHealthy
			fmt.Println("healthy status change to ", isHealthy, " for", host.Address)
		}

		host.RecentHealthChecks = []bool{}
	}
}

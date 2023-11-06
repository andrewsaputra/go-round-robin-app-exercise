package internal

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterHost_Valid_ReturnStatusOK(t *testing.T) {
	mgr := Helper_ConstructHostManager()

	response := mgr.RegisterHost("http://localhost:7777")

	assert.Equal(t, http.StatusOK, response.Code)
	assert.NotNil(t, response.Message)
	assert.NoError(t, response.Error)
}

func TestRegisterHost_Duplicate_ReturnStatusBadRequest(t *testing.T) {
	mgr := Helper_ConstructHostManager()

	hostAddress := "http://localhost:7777"
	mgr.RegisterHost(hostAddress)
	response := mgr.RegisterHost(hostAddress)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotNil(t, response.Message)
	assert.NoError(t, response.Error)
}

func TestDeRegisterHost_Valid_ReturnStatusOK(t *testing.T) {
	mgr := Helper_ConstructHostManager()

	hostAddress := "http://localhost:7777"
	mgr.RegisterHost(hostAddress)
	response := mgr.DeregisterHost(hostAddress)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.NotNil(t, response.Message)
	assert.NoError(t, response.Error)
}

func TestDeRegisterHost_NotRegistered_ReturnStatusBadRequest(t *testing.T) {
	mgr := Helper_ConstructHostManager()

	hostAddress := "http://localhost:7777"
	response := mgr.DeregisterHost(hostAddress)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotNil(t, response.Message)
	assert.NoError(t, response.Error)
}

func TestGetEligibleHosts_SomeHostsUnhealthy_ReturnHealthyHosts(t *testing.T) {
	mgr := Helper_ConstructHostManager()

	hostHealths := map[string]bool{}
	hostHealths["http://localhost:4001"] = true
	hostHealths["http://localhost:4002"] = false
	hostHealths["http://localhost:4003"] = true

	for addr, _ := range hostHealths {
		mgr.RegisterHost(addr)
	}

	numHealthy := 0
	for addr, healthy := range hostHealths {
		for i, _ := range mgr.hosts {
			host := &mgr.hosts[i]
			if host.Address == addr {
				host.Healthy = healthy
				break
			}
		}

		if healthy {
			numHealthy++
		}
	}

	eligibleHosts := mgr.GetEligibleHosts()

	assert.Equal(t, numHealthy, len(eligibleHosts))
	for _, host := range eligibleHosts {
		assert.True(t, host.Healthy)
	}
}

func TestGetEligibleHosts_AllHostsUnhealthy_ReturnAllHosts(t *testing.T) {
	mgr := Helper_ConstructHostManager()

	hostAddresses := []string{
		"http://localhost:4001",
		"http://localhost:4002",
		"http://localhost:4003",
	}

	for _, addr := range hostAddresses {
		mgr.RegisterHost(addr)
	}

	eligibleHosts := mgr.GetEligibleHosts()

	assert.Equal(t, len(hostAddresses), len(eligibleHosts))
	for i, host := range eligibleHosts {
		assert.Equal(t, hostAddresses[i], host.Address)
		assert.False(t, host.Healthy)
	}
}

func TestHealthCheckEvaluation_MultipleHostRegistered_EvaluationTriggered(t *testing.T) {
	config := Helper_ConstructHealthCheckConfig()
	config.IntervalSeconds = 1

	roundTripper := MockRoundTripper{}
	roundTripper.On("RoundTrip", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK}, nil)

	client := &http.Client{
		Transport: &roundTripper,
	}

	mgr := ConstructHostManager(client, config)

	hostAddresses := []string{"http://localhost:4001", "http://localhost:4002"}
	for _, addr := range hostAddresses {
		mgr.RegisterHost(addr)
	}

	timer := time.NewTimer(1100 * time.Millisecond)
	<-timer.C

	roundTripper.AssertNumberOfCalls(t, "RoundTrip", len(hostAddresses))
}

func TestHealthCheckEvaluation_HostReturnsHealthy_UpdateStatusToHealthy(t *testing.T) {
	config := Helper_ConstructHealthCheckConfig()
	config.IntervalSeconds = 1
	config.NumRequired = 2

	roundTripper := MockRoundTripper{}
	roundTripper.On("RoundTrip", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK}, nil)

	client := &http.Client{
		Transport: &roundTripper,
	}

	mgr := ConstructHostManager(client, config)

	hostAddresses := []string{"http://localhost:4001", "http://localhost:4002"}
	for _, addr := range hostAddresses {
		mgr.RegisterHost(addr)
	}

	for _, host := range mgr.hosts {
		assert.False(t, host.Healthy)
	}

	timer := time.NewTimer(1100 * time.Millisecond)
	<-timer.C

	roundTripper.AssertNumberOfCalls(t, "RoundTrip", len(hostAddresses))
	for _, host := range mgr.hosts {
		assert.False(t, host.Healthy)
	}

	timer = time.NewTimer(1100 * time.Millisecond)
	<-timer.C

	roundTripper.AssertNumberOfCalls(t, "RoundTrip", 2*len(hostAddresses))
	for _, host := range mgr.hosts {
		assert.True(t, host.Healthy)
	}
}

func TestHealthCheckEvaluation_HostUnreachable_UpdateStatusToUnhealthy(t *testing.T) {
	config := Helper_ConstructHealthCheckConfig()
	config.IntervalSeconds = 1
	config.NumRequired = 1

	roundTripper := MockRoundTripper{}
	roundTripper.On("RoundTrip", mock.Anything).
		Return(&http.Response{}, errors.New("host unreachable"))

	client := &http.Client{
		Transport: &roundTripper,
	}

	mgr := ConstructHostManager(client, config)
	mgr.RegisterHost("http://localhost:4001")
	mgr.hosts[0].Healthy = true
	assert.True(t, mgr.hosts[0].Healthy)

	timer := time.NewTimer(1100 * time.Millisecond)
	<-timer.C

	roundTripper.AssertNumberOfCalls(t, "RoundTrip", 1)
	assert.False(t, mgr.hosts[0].Healthy)
}

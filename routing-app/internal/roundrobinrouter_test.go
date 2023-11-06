package internal

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetNextTargetHost_MultipleEligibleHosts_ReturnWithRoundRobinAlgorithm(t *testing.T) {
	router, hostManager := Helper_ConstructRoundRobinRouter()

	hostAddresses := []string{"host1", "host2", "host3"}
	for _, addr := range hostAddresses {
		hostManager.RegisterHost(addr)
	}

	n := len(hostAddresses)
	for i := 0; i < 3*n; i++ {
		target, _ := router.getNextTargetHost()
		assert.Equal(t, hostAddresses[i%n], target)
	}
}

func TestGetNextTargetHost_NoEligibleHost_ReturnError(t *testing.T) {
	router, _ := Helper_ConstructRoundRobinRouter()

	_, err := router.getNextTargetHost()
	assert.Error(t, err)
}

func TestGetNextTargetHost_SingleHost_ReturnSameHost(t *testing.T) {
	router, hostManager := Helper_ConstructRoundRobinRouter()

	hostAddr := "host1"
	hostManager.RegisterHost(hostAddr)

	for i := 0; i < 10; i++ {
		target, _ := router.getNextTargetHost()
		assert.Equal(t, hostAddr, target)
	}
}

func TestForwardRequest_Standard_ReturnStatusOK(t *testing.T) {
	router, hostManager := Helper_ConstructRoundRobinRouter()
	hostManager.RegisterHost("host1")

	body := io.NopCloser(bytes.NewReader([]byte(`{"key" : "value"}`)))
	request, _ := http.NewRequest("POST", "/test", body)
	resp, err := router.ForwardRequest(request)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NoError(t, err)
}

func TestForwardRequest_NoAvailableHost_ReturnStatusOK(t *testing.T) {
	router, _ := Helper_ConstructRoundRobinRouter()

	request, _ := http.NewRequest("GET", "/test", nil)
	resp, err := router.ForwardRequest(request)

	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestForwardRequest_HostReturnError_ReturnStatusError(t *testing.T) {
	roundTripper := new(MockRoundTripper)
	roundTripper.On("RoundTrip", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusInternalServerError}, nil)

	client := &http.Client{
		Transport: roundTripper,
	}

	hostManager := Helper_ConstructHostManager()
	hostManager.RegisterHost("host1")

	router := ConstructRoundRobinRouter(client, hostManager, 0)

	body := io.NopCloser(bytes.NewReader([]byte(`{"key" : "value"}`)))
	request, _ := http.NewRequest("POST", "/test", body)
	resp, err := router.ForwardRequest(request)

	assert.Nil(t, resp)
	assert.Error(t, err)
	roundTripper.AssertNumberOfCalls(t, "RoundTrip", 1)
}

func TestForwardRequest_WithRetryAttempt_ReturnStatusOkAfterRetry(t *testing.T) {
	roundTripper := new(MockRoundTripper)
	mock1 := roundTripper.On("RoundTrip", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusInternalServerError}, nil).
		Times(2)

	roundTripper.On("RoundTrip", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK}, nil).
		NotBefore(mock1)

	client := &http.Client{
		Transport: roundTripper,
	}

	hostManager := Helper_ConstructHostManager()
	hostManager.RegisterHost("host1")
	hostManager.RegisterHost("host2")
	hostManager.RegisterHost("host3")

	router := ConstructRoundRobinRouter(client, hostManager, 2)

	body := io.NopCloser(bytes.NewReader([]byte(`{"key" : "value"}`)))
	request, _ := http.NewRequest("POST", "/test", body)
	resp, err := router.ForwardRequest(request)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NoError(t, err)
	roundTripper.AssertNumberOfCalls(t, "RoundTrip", 3)
}

/*

func (this *RoundRobinRouter) ForwardRequest(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	numAttempts := 0
	for numAttempts <= this.maxRetries {
		targetHost, err := this.getNextTargetHost()
		if err != nil {
			return nil, err
		}

		url := targetHost + req.URL.Path
		newReq, err := http.NewRequest(req.Method, url, io.NopCloser(bytes.NewReader(body)))
		if err != nil {
			numAttempts++
			continue
		}

		newReq.Header = req.Header

		fmt.Println("forwarding request to :", newReq.URL)
		resp, err := this.client.Do(newReq)
		if err != nil || resp.StatusCode == http.StatusInternalServerError {
			numAttempts++
			continue
		}

		return resp, nil
	}

	return nil, errors.New("request forwarding failed. please try again after a while.")
}
*/

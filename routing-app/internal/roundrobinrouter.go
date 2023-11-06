package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func ConstructRoundRobinRouter(client *http.Client, hostManager *HostManager, maxRetries int) *RoundRobinRouter {
	return &RoundRobinRouter{
		client:      client,
		hostManager: hostManager,
		maxRetries:  maxRetries,
		hostIndex:   0,
	}
}

type RoundRobinRouter struct {
	client      *http.Client
	hostManager *HostManager
	maxRetries  int
	hostIndex   int
}

func (this *RoundRobinRouter) ForwardRequest(req *http.Request) (*http.Response, error) {
	var body []byte
	var err error
	if req.Body != nil {
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}

	numAttempts := 0
	for numAttempts <= this.maxRetries {
		targetHost, err := this.getNextTargetHost()
		if err != nil {
			return nil, err
		}

		url := targetHost + req.URL.Path
		var newReq *http.Request
		if req.Body == nil {
			newReq, err = http.NewRequest(req.Method, url, nil)
		} else {
			newReq, err = http.NewRequest(req.Method, url, io.NopCloser(bytes.NewReader(body)))
		}
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

func (this *RoundRobinRouter) getNextTargetHost() (string, error) {
	hosts := this.hostManager.GetEligibleHosts()
	lenHosts := len(hosts)

	switch lenHosts {
	case 0:
		return "", errors.New("no available hosts")
	case 1:
		return hosts[0].Address, nil
	default:
		if this.hostIndex >= lenHosts {
			this.hostIndex = 0
		}

		host := hosts[this.hostIndex]
		this.hostIndex++
		return host.Address, nil
	}
}

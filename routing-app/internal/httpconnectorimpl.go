package internal

import (
	"bytes"
	"fmt"
	"net/http"
)

type HttpConnectorImpl struct {
	Client *http.Client
}

func CreateHttpConnectorImpl(client *http.Client) *HttpConnectorImpl {
	return &HttpConnectorImpl{
		Client: client,
	}
}

func (this *HttpConnectorImpl) DoPost(url string, body []byte) (*http.Response, error) {
	fmt.Printf("url : %v\n", url)
	response, err := this.Client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return response, nil
}

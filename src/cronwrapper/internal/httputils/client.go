package httputils

import (
	"net"
	"net/http"
)

// ClientConfig A simple config struct.
type ClientConfig struct{}

// GetClient The method returns an http client based on the config.
func GetClient(_ *ClientConfig) (*http.Client, error) {
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: DialTimeout,
		}).Dial,
		IdleConnTimeout: IdleConnTimeout,
	}
	client := &http.Client{
		Timeout:   RequestTimeout,
		Transport: tr,
	}

	return client, nil
}

package client

import (
	"net"
	"net/http"
	"time"
)

// NewHTTPClient returns a tuned HTTP client for API calls
func NewHTTPClient() *http.Client {
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
	}
	return &http.Client{
		Timeout:   5 * time.Second, // end-to-end request timeout
		Transport: tr,
	}
}
